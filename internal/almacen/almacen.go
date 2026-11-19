// Package almacen abstrae la fuente de datos (cuentas) y el log de
// auditoría. La implementación JSON escribe de forma atómica.
package almacen

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"prj_withdrawal/internal/dominio"
)

// Almacen es la interfaz que consume el servicio (inversión de dependencias).
type Almacen interface {
	CargarCuentas() ([]dominio.Cuenta, error)
	GuardarCuentas([]dominio.Cuenta) error
	LeerLog() ([]dominio.Transaccion, error)
	EscribirLog([]dominio.Transaccion) error
}

// JSON implementa Almacen sobre dos archivos JSON.
type JSON struct {
	RutaCuentas string
	RutaLog     string
}

// InicializarDesdeSemilla copia la semilla a RutaCuentas si esta no existe.
func (a JSON) InicializarDesdeSemilla(semilla string) error {
	if _, err := os.Stat(a.RutaCuentas); err == nil {
		return nil
	}
	src, err := os.Open(semilla)
	if err != nil {
		return fmt.Errorf("no existe %s ni la semilla %s: %w", a.RutaCuentas, semilla, err)
	}
	defer src.Close()
	if err := os.MkdirAll(filepath.Dir(a.RutaCuentas), 0o755); err != nil {
		return err
	}
	raw, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	return escribirAtomico(a.RutaCuentas, raw)
}

func (a JSON) CargarCuentas() ([]dominio.Cuenta, error) {
	raw, err := os.ReadFile(a.RutaCuentas)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer la fuente de datos: %w", err)
	}
	var l []dominio.Cuenta
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, fmt.Errorf("fuente de datos con JSON inválido: %w", err)
	}
	return l, nil
}

func (a JSON) GuardarCuentas(l []dominio.Cuenta) error {
	raw, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return escribirAtomico(a.RutaCuentas, raw)
}

func (a JSON) LeerLog() ([]dominio.Transaccion, error) {
	raw, err := os.ReadFile(a.RutaLog)
	if errors.Is(err, os.ErrNotExist) || (err == nil && len(raw) == 0) {
		return []dominio.Transaccion{}, nil
	}
	if err != nil {
		return nil, err
	}
	var l []dominio.Transaccion
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, fmt.Errorf("log con JSON inválido: %w", err)
	}
	return l, nil
}

func (a JSON) EscribirLog(l []dominio.Transaccion) error {
	raw, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return escribirAtomico(a.RutaLog, raw)
}

// escribirAtomico escribe a un temporal, sincroniza y renombra (todo o nada).
func escribirAtomico(ruta string, datos []byte) error {
	tmp := ruta + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err = f.Write(datos); err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, ruta)
}
