package almacen

import (
	"os"
	"path/filepath"
	"testing"

	"prj_withdrawal/internal/dominio"
)

func TestSemillaYRoundTrip(t *testing.T) {
	dir := t.TempDir()
	semilla := filepath.Join(dir, "seed.json")
	_ = os.WriteFile(semilla, []byte(`[{"nombre":"Alice","saldo":234.56}]`), 0o644)
	a := JSON{RutaCuentas: filepath.Join(dir, "d", "cuentas.json"), RutaLog: filepath.Join(dir, "d", "log.json")}

	if err := a.InicializarDesdeSemilla(semilla); err != nil {
		t.Fatal(err)
	}
	cs, err := a.CargarCuentas()
	if err != nil || len(cs) != 1 || cs[0].Saldo != 23456 {
		t.Fatalf("carga: %+v %v", cs, err)
	}
	cs[0].Saldo = 100
	if err := a.GuardarCuentas(cs); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(a.RutaCuentas + ".tmp"); err == nil {
		t.Error("no debe quedar archivo temporal tras la escritura atómica")
	}
	// una segunda inicialización NO debe pisar los datos existentes
	_ = a.InicializarDesdeSemilla(semilla)
	cs, _ = a.CargarCuentas()
	if cs[0].Saldo != 100 {
		t.Errorf("la semilla pisó los datos: %+v", cs)
	}
}

func TestLogVacioYJSONInvalido(t *testing.T) {
	dir := t.TempDir()
	a := JSON{RutaCuentas: filepath.Join(dir, "c.json"), RutaLog: filepath.Join(dir, "log.json")}
	if l, err := a.LeerLog(); err != nil || len(l) != 0 {
		t.Errorf("log inexistente debe ser vacío: %v %v", l, err)
	}
	_ = os.WriteFile(a.RutaLog, []byte("{no es json"), 0o644)
	if _, err := a.LeerLog(); err == nil {
		t.Error("JSON inválido debía devolver error")
	}
	_ = a.EscribirLog([]dominio.Transaccion{{ID: "TX-000001"}})
	if l, err := a.LeerLog(); err != nil || len(l) != 1 {
		t.Errorf("round-trip del log: %v %v", l, err)
	}
}
