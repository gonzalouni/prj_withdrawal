package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"prj_withdrawal/internal/almacen"
	"prj_withdrawal/internal/servicio"
)

func TestMenuInteractivo(t *testing.T) {
	dir := t.TempDir()
	a := almacen.JSON{RutaCuentas: filepath.Join(dir, "c.json"), RutaLog: filepath.Join(dir, "l.json")}
	_ = os.WriteFile(a.RutaCuentas, []byte(`[{"nombre":"Alice","saldo":234.56}]`), 0o644)
	s, err := servicio.Nuevo(a)
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	// titular inexistente, luego Alice: abono, retiro excesivo, monto ilegible, saldo, historial, salir
	entrada := strings.NewReader("Nadie\nAlice\n3\n100\n2\n1000\n2\nabc\n1\n4\n0\n")
	if err := Menu(s, entrada, &out); err != nil {
		t.Fatal(err)
	}
	txt := out.String()
	for _, esperado := range []string{"Cuenta no encontrada", "EXITOSA", "RECHAZADA",
		"Saldo insuficiente. Intente con un monto menor", "[ERROR]", "Saldo disponible: S/ 334.56"} {
		if !strings.Contains(txt, esperado) {
			t.Errorf("la salida no contiene %q", esperado)
		}
	}
	if h, _ := s.Historial(""); len(h) != 3 {
		t.Errorf("deben auditarse 3 intentos, hay %d", len(h))
	}
}
