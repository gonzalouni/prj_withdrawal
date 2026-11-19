// Package dominio contiene las entidades y reglas de negocio puras
// (sin E/S): Dinero, Cuenta y Transaccion.
package dominio

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Dinero representa un monto en la unidad mínima (centavos) como entero,
// tal como exige el Modelo de Dominio ("cast entero, unidad mínima") para
// evitar errores de redondeo de float64.
type Dinero int64

// ParseDinero convierte texto decimal ("200", "200.5", "200,50") a centavos.
// Rechaza notación científica, NaN/Inf y más de 2 decimales.
func ParseDinero(texto string) (Dinero, error) {
	s := strings.TrimSpace(strings.Replace(texto, ",", ".", 1))
	neg := false
	switch {
	case strings.HasPrefix(s, "-"):
		neg, s = true, s[1:]
	case strings.HasPrefix(s, "+"):
		s = s[1:]
	}
	ent, dec, _ := strings.Cut(s, ".")
	if ent == "" && dec == "" {
		return 0, ErrMontoInvalido
	}
	if ent == "" {
		ent = "0"
	}
	if len(ent) > 12 || len(dec) > 2 || !soloDigitos(ent) || !soloDigitos(dec) {
		return 0, ErrMontoInvalido
	}
	e, err := strconv.ParseInt(ent, 10, 64)
	if err != nil {
		return 0, ErrMontoInvalido
	}
	var d int64
	if dec != "" {
		d, _ = strconv.ParseInt((dec + "0")[:2], 10, 64)
	}
	v := e*100 + d
	if neg {
		v = -v
	}
	return Dinero(v), nil
}

func soloDigitos(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// DesdeFloat convierte un float64 (p. ej. de pruebas) a centavos redondeando.
func DesdeFloat(f float64) Dinero { return Dinero(math.Round(f * 100)) }

// String devuelve el monto con 2 decimales, p. ej. "334.56".
func (d Dinero) String() string {
	signo, v := "", int64(d)
	if v < 0 {
		signo, v = "-", -v
	}
	return fmt.Sprintf("%s%d.%02d", signo, v/100, v%100)
}

// MarshalJSON escribe el monto como número decimal (compatible con v2.0.0).
func (d Dinero) MarshalJSON() ([]byte, error) { return []byte(d.String()), nil }

// UnmarshalJSON acepta números (234.56) o cadenas ("234.56").
func (d *Dinero) UnmarshalJSON(b []byte) error {
	v, err := ParseDinero(strings.Trim(string(b), `"`))
	if err != nil {
		return err
	}
	*d = v
	return nil
}
