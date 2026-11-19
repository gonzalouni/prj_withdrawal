// Package servicio orquesta los casos de uso (Retiro, Abono) garantizando
// atomicidad (RN-03) y auditoría de todo intento (US-04).
package servicio

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"prj_withdrawal/internal/almacen"
	"prj_withdrawal/internal/dominio"
)

// Servicio es seguro para uso concurrente.
type Servicio struct {
	mu      sync.Mutex
	alm     almacen.Almacen
	cuentas map[string]*dominio.Cuenta
	seq     int
	Reloj   func() time.Time
}

// Nuevo carga las cuentas y la secuencia de IDs desde el almacén.
func Nuevo(alm almacen.Almacen) (*Servicio, error) {
	lista, err := alm.CargarCuentas()
	if err != nil {
		return nil, err
	}
	s := &Servicio{alm: alm, cuentas: map[string]*dominio.Cuenta{}, Reloj: time.Now}
	for _, c := range lista {
		if c.Nombre == "" || c.Saldo < 0 {
			return nil, fmt.Errorf("cuenta inválida en la fuente de datos: %+v", c)
		}
		if _, dup := s.cuentas[c.Nombre]; dup {
			return nil, fmt.Errorf("cuenta duplicada: %s", c.Nombre)
		}
		cc := c
		s.cuentas[c.Nombre] = &cc
	}
	hist, err := alm.LeerLog()
	if err != nil {
		return nil, err
	}
	s.seq = len(hist)
	return s, nil
}

// Cuentas devuelve una copia ordenada de las cuentas.
func (s *Servicio) Cuentas() []dominio.Cuenta {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.copiaCuentas()
}

// Saldo devuelve el saldo actual de una cuenta.
func (s *Servicio) Saldo(nombre string) (dominio.Dinero, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.cuentas[nombre]
	if !ok {
		return 0, dominio.ErrCuentaNoExiste
	}
	return c.Saldo, nil
}

// Historial devuelve las transacciones; si nombre != "" filtra por cuenta.
func (s *Servicio) Historial(nombre string) ([]dominio.Transaccion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, err := s.alm.LeerLog()
	if err != nil || nombre == "" {
		return h, err
	}
	out := []dominio.Transaccion{}
	for _, t := range h {
		if t.Cuenta == nombre {
			out = append(out, t)
		}
	}
	return out, nil
}

// Ejecutar procesa una operación. El monto llega como texto para que un monto
// ilegible también quede auditado como ERROR.
//
//	EXITOSA   -> error == nil
//	RECHAZADA -> error == nil (regla de negocio)
//	ERROR     -> error != nil (validación o persistencia)
//
// Si falla guardar el saldo o el log, el saldo se revierte (memoria y disco).
func (s *Servicio) Ejecutar(nombre string, op dominio.Operacion, monto string) (dominio.Transaccion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cuenta := s.cuentas[nombre]
	tx := dominio.Transaccion{
		ID:     fmt.Sprintf("TX-%06d", s.seq+1),
		Fecha:  s.Reloj().UTC().Format(time.RFC3339),
		Cuenta: nombre,
		Tipo:   op,
	}
	if cuenta != nil {
		tx.SaldoAntes, tx.SaldoDespues = cuenta.Saldo, cuenta.Saldo
	}

	m, errMonto := dominio.ParseDinero(monto)
	tx.Monto = m
	antes := tx.SaldoAntes

	err := aplicar(cuenta, op, m, errMonto)
	switch {
	case err == nil:
		tx.Estado, tx.Motivo, tx.SaldoDespues = dominio.EstadoExitosa, "Operación completada", cuenta.Saldo
		if gerr := s.guardarCuentas(); gerr != nil {
			cuenta.Saldo = antes
			return s.fallo(tx, antes, gerr, "no se pudo guardar el saldo")
		}
		return s.registrar(tx, cuenta, antes)
	case errors.Is(err, dominio.ErrSaldoInsuficiente):
		tx.Estado, tx.Motivo = dominio.EstadoRechazada, dominio.MensajeSaldoInsuficiente
		return s.registrar(tx, nil, antes)
	default:
		tx.Estado, tx.Motivo = dominio.EstadoError, err.Error()
		t, rerr := s.registrar(tx, nil, antes)
		if rerr != nil {
			return t, rerr
		}
		return t, err
	}
}

// aplicar valida y ejecuta la operación sobre la entidad (sin E/S).
func aplicar(c *dominio.Cuenta, op dominio.Operacion, m dominio.Dinero, errMonto error) error {
	switch {
	case op != dominio.OpRetiro && op != dominio.OpAbono:
		return dominio.ErrOperacion
	case errMonto != nil:
		return errMonto
	case c == nil:
		return dominio.ErrCuentaNoExiste
	case op == dominio.OpRetiro:
		return c.Retirar(m)
	default:
		return c.Depositar(m)
	}
}

// registrar escribe la transacción en el log; si falla y la cuenta ya fue
// modificada, revierte el saldo en memoria y en disco.
func (s *Servicio) registrar(tx dominio.Transaccion, cuenta *dominio.Cuenta, antes dominio.Dinero) (dominio.Transaccion, error) {
	hist, err := s.alm.LeerLog()
	if err == nil {
		err = s.alm.EscribirLog(append(hist, tx))
	}
	if err != nil {
		if cuenta != nil {
			cuenta.Saldo = antes
			_ = s.guardarCuentas() // mejor esfuerzo: restaurar disco
		}
		return s.fallo(tx, antes, err, "no se pudo registrar la auditoría")
	}
	s.seq++
	return tx, nil
}

func (s *Servicio) fallo(tx dominio.Transaccion, antes dominio.Dinero, err error, motivo string) (dominio.Transaccion, error) {
	tx.Estado, tx.SaldoDespues, tx.Motivo = dominio.EstadoError, antes, motivo+": "+err.Error()
	return tx, fmt.Errorf("%w: %v", dominio.ErrPersistencia, err)
}

func (s *Servicio) guardarCuentas() error {
	return s.alm.GuardarCuentas(s.copiaCuentas())
}

// copiaCuentas requiere el lock tomado por el llamador.
func (s *Servicio) copiaCuentas() []dominio.Cuenta {
	out := make([]dominio.Cuenta, 0, len(s.cuentas))
	for _, c := range s.cuentas {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Nombre < out[j].Nombre })
	return out
}
