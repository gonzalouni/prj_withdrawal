package dominio

import (
	"errors"
	"math"
)

// Errores de dominio.
var (
	ErrMontoInvalido     = errors.New("monto inválido: debe ser un número mayor que cero con máximo 2 decimales")
	ErrSaldoInsuficiente = errors.New(MensajeSaldoInsuficiente)
	ErrCuentaNoExiste    = errors.New("la cuenta no existe")
	ErrOperacion         = errors.New("operación no soportada")
	ErrPersistencia      = errors.New("fallo de persistencia: la operación fue revertida")
)

// Cuenta es la entidad de negocio de los Features Gherkin.
type Cuenta struct {
	Nombre string `json:"nombre"`
	Saldo  Dinero `json:"saldo"`
}

// Retirar aplica RN-01/RN-02: monto > 0 y no mayor al saldo.
func (c *Cuenta) Retirar(m Dinero) error {
	if m <= 0 {
		return ErrMontoInvalido
	}
	if m > c.Saldo {
		return ErrSaldoInsuficiente
	}
	c.Saldo -= m
	return nil
}

// Depositar aplica US-01b: solo montos mayores que cero.
func (c *Cuenta) Depositar(m Dinero) error {
	if m <= 0 || int64(m) > math.MaxInt64-int64(c.Saldo) {
		return ErrMontoInvalido
	}
	c.Saldo += m
	return nil
}
