package servicio

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"prj_withdrawal/internal/almacen"
	"prj_withdrawal/internal/dominio"
)

const semilla = `[
 {"nombre":"Alice","saldo":234.56},
 {"nombre":"Hamza","saldo":198.76},
 {"nombre":"Pedro","saldo":300.00}
]`

func nuevoAlmacen(t *testing.T) almacen.JSON {
	t.Helper()
	dir := t.TempDir()
	a := almacen.JSON{RutaCuentas: filepath.Join(dir, "cuentas.json"), RutaLog: filepath.Join(dir, "log.json")}
	if err := os.WriteFile(a.RutaCuentas, []byte(semilla), 0o644); err != nil {
		t.Fatal(err)
	}
	return a
}

func nuevoSvc(t *testing.T) (*Servicio, almacen.JSON) {
	t.Helper()
	a := nuevoAlmacen(t)
	s, err := Nuevo(a)
	if err != nil {
		t.Fatal(err)
	}
	return s, a
}

func saldoDe(t *testing.T, s *Servicio, n string) string {
	t.Helper()
	v, err := s.Saldo(n)
	if err != nil {
		t.Fatal(err)
	}
	return v.String()
}

func TestMontoIlegibleQuedaAuditadoComoError(t *testing.T) {
	for _, m := range []string{"abc", "", "NaN", "1e9", "1.999"} {
		s, _ := nuevoSvc(t)
		tx, err := s.Ejecutar("Pedro", dominio.OpRetiro, m)
		if !errors.Is(err, dominio.ErrMontoInvalido) || tx.Estado != dominio.EstadoError {
			t.Errorf("monto %q: esperaba ERROR, tx=%+v err=%v", m, tx, err)
		}
		h, _ := s.Historial("")
		if len(h) != 1 || saldoDe(t, s, "Pedro") != "300.00" {
			t.Errorf("monto %q: debe auditarse y no cambiar el saldo", m)
		}
	}
}

func TestCuentaInexistenteYOperacionDesconocida(t *testing.T) {
	s, _ := nuevoSvc(t)
	if _, err := s.Ejecutar("Nadie", dominio.OpAbono, "10"); !errors.Is(err, dominio.ErrCuentaNoExiste) {
		t.Errorf("cuenta inexistente: %v", err)
	}
	if _, err := s.Ejecutar("Alice", "TRANSFERIR", "10"); !errors.Is(err, dominio.ErrOperacion) {
		t.Errorf("operación desconocida: %v", err)
	}
	if h, _ := s.Historial(""); len(h) != 2 {
		t.Errorf("ambos intentos deben quedar en el log, hay %d", len(h))
	}
}

func TestPersistenciaYSecuenciaEntreReinicios(t *testing.T) {
	s, a := nuevoSvc(t)
	if _, err := s.Ejecutar("Alice", dominio.OpAbono, "100"); err != nil {
		t.Fatal(err)
	}
	s2, err := Nuevo(a)
	if err != nil {
		t.Fatal(err)
	}
	if saldoDe(t, s2, "Alice") != "334.56" {
		t.Error("el saldo no persistió")
	}
	if tx, _ := s2.Ejecutar("Alice", dominio.OpAbono, "1"); tx.ID != "TX-000002" {
		t.Errorf("ID esperado TX-000002, obtuvo %s", tx.ID)
	}
}

// --- Inyección de fallos para probar la atomicidad (RN-03) ---

type fallaLog struct{ almacen.JSON }

func (fallaLog) EscribirLog([]dominio.Transaccion) error { return errors.New("disco lleno") }

type fallaCuentas struct{ almacen.JSON }

func (fallaCuentas) GuardarCuentas([]dominio.Cuenta) error { return errors.New("disco lleno") }

func TestAtomicidadRollbackSiFallaElLog(t *testing.T) {
	a := nuevoAlmacen(t)
	s, _ := Nuevo(fallaLog{a})
	tx, err := s.Ejecutar("Alice", dominio.OpAbono, "100")
	if !errors.Is(err, dominio.ErrPersistencia) || tx.Estado != dominio.EstadoError {
		t.Fatalf("esperaba ERROR de persistencia, tx=%+v err=%v", tx, err)
	}
	if saldoDe(t, s, "Alice") != "234.56" {
		t.Error("rollback en memoria falló")
	}
	cs, _ := a.CargarCuentas()
	if cs[0].Saldo != 23456 {
		t.Errorf("rollback en disco falló: %+v", cs)
	}
}

func TestAtomicidadRollbackSiFallaGuardarSaldo(t *testing.T) {
	s, _ := Nuevo(fallaCuentas{nuevoAlmacen(t)})
	tx, err := s.Ejecutar("Hamza", dominio.OpRetiro, "50")
	if !errors.Is(err, dominio.ErrPersistencia) || tx.Estado != dominio.EstadoError {
		t.Fatalf("tx=%+v err=%v", tx, err)
	}
	if saldoDe(t, s, "Hamza") != "198.76" {
		t.Error("el saldo debe quedar intacto")
	}
}

func TestConcurrenciaMantieneConsistencia(t *testing.T) {
	s, _ := nuevoSvc(t)
	const n = 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = s.Ejecutar("Pedro", dominio.OpAbono, "1") }()
	}
	wg.Wait()
	if saldoDe(t, s, "Pedro") != "350.00" {
		t.Errorf("saldo inconsistente: %s", saldoDe(t, s, "Pedro"))
	}
	h, _ := s.Historial("")
	ids := map[string]bool{}
	for _, x := range h {
		ids[x.ID] = true
	}
	if len(h) != n || len(ids) != n {
		t.Errorf("log inconsistente: %d entradas, %d ids únicos", len(h), len(ids))
	}
}

func TestHistorialFiltraPorCuenta(t *testing.T) {
	s, _ := nuevoSvc(t)
	_, _ = s.Ejecutar("Alice", dominio.OpAbono, "1")
	_, _ = s.Ejecutar("Pedro", dominio.OpAbono, "1")
	if h, _ := s.Historial("Alice"); len(h) != 1 || h[0].Cuenta != "Alice" {
		t.Errorf("filtro incorrecto: %+v", h)
	}
}
