package dominio

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestParseDinero(t *testing.T) {
	validos := map[string]Dinero{
		"200": 20000, "200.5": 20050, "200,50": 20050, "0.01": 1, ".5": 50,
		"-20": -2000, "+7": 700, " 12.30 ": 1230, "0": 0,
	}
	for in, esp := range validos {
		got, err := ParseDinero(in)
		if err != nil || got != esp {
			t.Errorf("ParseDinero(%q) = %d, %v; esperaba %d", in, got, err, esp)
		}
	}
	for _, in := range []string{"", "abc", "1e3", "NaN", "Inf", "1.234", "--1", "1.2.3", "12345678901234", ".", "-"} {
		if _, err := ParseDinero(in); !errors.Is(err, ErrMontoInvalido) {
			t.Errorf("ParseDinero(%q) debía fallar con ErrMontoInvalido, obtuvo %v", in, err)
		}
	}
}

func TestDineroStringYJSON(t *testing.T) {
	if s := Dinero(33456).String(); s != "334.56" {
		t.Errorf("String = %s", s)
	}
	if s := Dinero(-5).String(); s != "-0.05" {
		t.Errorf("String = %s", s)
	}
	b, _ := json.Marshal(Cuenta{"Alice", 23456})
	if string(b) != `{"nombre":"Alice","saldo":234.56}` {
		t.Errorf("JSON = %s", b)
	}
	var c Cuenta
	if err := json.Unmarshal([]byte(`{"nombre":"Bob","saldo":500}`), &c); err != nil || c.Saldo != 50000 {
		t.Errorf("Unmarshal = %+v, %v", c, err)
	}
}

func TestCuentaRetirar(t *testing.T) {
	casos := []struct {
		nombre string
		saldo  Dinero
		monto  Dinero
		err    error
		final  Dinero
	}{
		{"exitoso dentro del saldo", 23456, 20000, nil, 3456},
		{"retiro exacto del saldo", 5000, 5000, nil, 0},
		{"excede el saldo", 19876, 20000, ErrSaldoInsuficiente, 19876},
		{"monto negativo", 30000, -2000, ErrMontoInvalido, 30000},
		{"monto cero", 30000, 0, ErrMontoInvalido, 30000},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			c := Cuenta{"X", tc.saldo}
			if err := c.Retirar(tc.monto); !errors.Is(err, tc.err) || c.Saldo != tc.final {
				t.Errorf("err=%v saldo=%d; esperaba err=%v saldo=%d", err, c.Saldo, tc.err, tc.final)
			}
		})
	}
}

func TestCuentaDepositar(t *testing.T) {
	c := Cuenta{"Alice", 23456}
	if err := c.Depositar(10000); err != nil || c.Saldo != 33456 {
		t.Fatalf("abono exitoso falló: %v %d", err, c.Saldo)
	}
	for _, m := range []Dinero{0, -5000} {
		if err := c.Depositar(m); !errors.Is(err, ErrMontoInvalido) || c.Saldo != 33456 {
			t.Errorf("monto %d debía rechazarse sin cambiar el saldo", m)
		}
	}
	if err := c.Depositar(Dinero(1<<63 - 1)); !errors.Is(err, ErrMontoInvalido) {
		t.Errorf("el desbordamiento debía rechazarse")
	}
}

// Regresión del defecto de float64 (0.1 + 0.2 != 0.3) que motivó el uso de centavos.
func TestSinErrorDeRedondeo(t *testing.T) {
	c := Cuenta{"X", 0}
	a, _ := ParseDinero("0.1")
	b, _ := ParseDinero("0.2")
	_ = c.Depositar(a)
	_ = c.Depositar(b)
	if c.Saldo != 30 {
		t.Errorf("0.1 + 0.2 debe ser exactamente 0.30, obtuvo %s", c.Saldo)
	}
}
