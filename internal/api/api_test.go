package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"prj_withdrawal/internal/almacen"
	"prj_withdrawal/internal/dominio"
	"prj_withdrawal/internal/servicio"
)

func servidor(t *testing.T) *httptest.Server {
	t.Helper()
	dir := t.TempDir()
	a := almacen.JSON{RutaCuentas: filepath.Join(dir, "c.json"), RutaLog: filepath.Join(dir, "l.json")}
	_ = os.WriteFile(a.RutaCuentas, []byte(`[{"nombre":"Alice","saldo":234.56},{"nombre":"Hamza","saldo":198.76}]`), 0o644)
	svc, err := servicio.Nuevo(a)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(Nuevo(svc))
	t.Cleanup(ts.Close)
	return ts
}

func post(t *testing.T, ts *httptest.Server, cuerpo string) (int, dominio.Transaccion, string) {
	t.Helper()
	r, err := http.Post(ts.URL+"/api/operaciones", "application/json", strings.NewReader(cuerpo))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	var tx dominio.Transaccion
	var raw map[string]any
	b, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(b, &tx)
	_ = json.Unmarshal(b, &raw)
	e, _ := raw["error"].(string)
	return r.StatusCode, tx, e
}

func TestOperacionesYCodigosHTTP(t *testing.T) {
	ts := servidor(t)
	casos := []struct {
		nombre string
		cuerpo string
		codigo int
		estado dominio.Estado
	}{
		{"retiro exitoso", `{"cuenta":"Alice","tipo":"retiro","monto":"200.00"}`, 200, dominio.EstadoExitosa},
		{"monto numérico JSON", `{"cuenta":"Alice","tipo":"ABONO","monto":10.5}`, 200, dominio.EstadoExitosa},
		{"saldo insuficiente", `{"cuenta":"Hamza","tipo":"RETIRO","monto":"200"}`, 422, dominio.EstadoRechazada},
		{"monto negativo", `{"cuenta":"Hamza","tipo":"ABONO","monto":"-5"}`, 400, dominio.EstadoError},
		{"cuenta inexistente", `{"cuenta":"Nadie","tipo":"ABONO","monto":"5"}`, 404, dominio.EstadoError},
		{"operación desconocida", `{"cuenta":"Alice","tipo":"X","monto":"5"}`, 400, dominio.EstadoError},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			codigo, tx, _ := post(t, ts, tc.cuerpo)
			if codigo != tc.codigo || tx.Estado != tc.estado {
				t.Errorf("código=%d estado=%s; esperaba %d %s", codigo, tx.Estado, tc.codigo, tc.estado)
			}
		})
	}
}

func TestJSONMalformadoNoSeAudita(t *testing.T) {
	ts := servidor(t)
	for _, cuerpo := range []string{`{no json`, `{"cuenta":"Alice","extra":1}`} {
		if codigo, _, e := post(t, ts, cuerpo); codigo != 400 || e == "" {
			t.Errorf("cuerpo %q: código=%d error=%q", cuerpo, codigo, e)
		}
	}
	r, _ := http.Get(ts.URL + "/api/transacciones")
	var h []dominio.Transaccion
	_ = json.NewDecoder(r.Body).Decode(&h)
	if len(h) != 0 {
		t.Errorf("las solicitudes ilegibles no deben crear transacciones: %+v", h)
	}
}

func TestConsultasYEstaticos(t *testing.T) {
	ts := servidor(t)
	post(t, ts, `{"cuenta":"Alice","tipo":"ABONO","monto":"1"}`)

	r, _ := http.Get(ts.URL + "/api/cuentas/Alice")
	var c dominio.Cuenta
	_ = json.NewDecoder(r.Body).Decode(&c)
	if r.StatusCode != 200 || c.Saldo != 23556 {
		t.Errorf("cuenta: %d %+v", r.StatusCode, c)
	}
	if r, _ := http.Get(ts.URL + "/api/cuentas/Nadie"); r.StatusCode != 404 {
		t.Errorf("cuenta inexistente debe ser 404, fue %d", r.StatusCode)
	}
	r, _ = http.Get(ts.URL + "/api/transacciones?cuenta=Hamza")
	var h []dominio.Transaccion
	_ = json.NewDecoder(r.Body).Decode(&h)
	if len(h) != 0 {
		t.Errorf("filtro por cuenta: %+v", h)
	}
	r, _ = http.Get(ts.URL + "/")
	if r.StatusCode != 200 || !strings.HasPrefix(r.Header.Get("Content-Type"), "text/html") {
		t.Errorf("index: %d %s", r.StatusCode, r.Header.Get("Content-Type"))
	}
	if r.Header.Get("Content-Security-Policy") == "" || r.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("faltan cabeceras de seguridad")
	}
	if r, _ := http.Get(ts.URL + "/healthz"); r.StatusCode != 200 {
		t.Error("healthz")
	}
	if r, _ := http.Post(ts.URL+"/api/cuentas", "application/json", nil); r.StatusCode != 405 {
		t.Errorf("POST /api/cuentas debe ser 405, fue %d", r.StatusCode)
	}
}
