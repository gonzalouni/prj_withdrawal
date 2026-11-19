// Package api expone el servicio como API REST JSON y sirve la App WEB
// (HTML/CSS/JS embebidos en el binario).
package api

import (
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"prj_withdrawal/internal/dominio"
	"prj_withdrawal/internal/servicio"
)

//go:embed web
var webFS embed.FS

type API struct{ svc *servicio.Servicio }

// Nuevo construye el http.Handler completo (API + estáticos).
func Nuevo(svc *servicio.Servicio) http.Handler {
	a := &API{svc: svc}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]string{"estado": "ok"})
	})
	mux.HandleFunc("GET /api/cuentas", a.cuentas)
	mux.HandleFunc("GET /api/cuentas/{nombre}", a.cuenta)
	mux.HandleFunc("POST /api/operaciones", a.operacion)
	mux.HandleFunc("GET /api/transacciones", a.transacciones)
	sub, _ := fs.Sub(webFS, "web")
	mux.Handle("GET /", http.FileServerFS(sub))
	return cabeceras(mux)
}

func cabeceras(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		h.ServeHTTP(w, r)
	})
}

func escribirJSON(w http.ResponseWriter, codigo int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	_ = json.NewEncoder(w).Encode(v)
}

func errorJSON(w http.ResponseWriter, codigo int, msg string) {
	escribirJSON(w, codigo, map[string]string{"error": msg})
}

func (a *API) cuentas(w http.ResponseWriter, _ *http.Request) {
	escribirJSON(w, http.StatusOK, a.svc.Cuentas())
}

func (a *API) cuenta(w http.ResponseWriter, r *http.Request) {
	nombre := r.PathValue("nombre")
	saldo, err := a.svc.Saldo(nombre)
	if err != nil {
		errorJSON(w, http.StatusNotFound, err.Error())
		return
	}
	escribirJSON(w, http.StatusOK, dominio.Cuenta{Nombre: nombre, Saldo: saldo})
}

type solicitud struct {
	Cuenta string          `json:"cuenta"`
	Tipo   string          `json:"tipo"`
	Monto  json.RawMessage `json:"monto"`
}

func (a *API) operacion(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req solicitud
	if err := dec.Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}
	monto := strings.Trim(string(req.Monto), `" `)
	tx, err := a.svc.Ejecutar(req.Cuenta, dominio.Operacion(strings.ToUpper(req.Tipo)), monto)
	escribirJSON(w, codigoHTTP(tx, err), tx)
}

// codigoHTTP: 200 exitosa, 422 rechazada (regla de negocio), 404 cuenta
// inexistente, 500 persistencia, 400 cualquier otra validación.
func codigoHTTP(tx dominio.Transaccion, err error) int {
	switch {
	case tx.Estado == dominio.EstadoExitosa:
		return http.StatusOK
	case tx.Estado == dominio.EstadoRechazada:
		return http.StatusUnprocessableEntity
	case errors.Is(err, dominio.ErrCuentaNoExiste):
		return http.StatusNotFound
	case errors.Is(err, dominio.ErrPersistencia):
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}

func (a *API) transacciones(w http.ResponseWriter, r *http.Request) {
	h, err := a.svc.Historial(r.URL.Query().Get("cuenta"))
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	escribirJSON(w, http.StatusOK, h)
}
