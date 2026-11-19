package servicio

// Pruebas BDD: cada Scenario de features/*.feature se traduce a un subtest
// con la misma estructura Given / When / Then.

import (
	"testing"

	"prj_withdrawal/internal/dominio"
)

func TestFeatureWithdrawingCash(t *testing.T) {
	t.Run("Rule: Customers cannot withdraw more than their balance", func(t *testing.T) {
		t.Run("Scenario: Successful withdrawal within balance", func(t *testing.T) {
			s, _ := nuevoSvc(t)                                        // Given Alice has 234.56
			tx, err := s.Ejecutar("Alice", dominio.OpRetiro, "200.00") // When Alice tries to withdraw 200.00
			if err != nil || tx.Estado != dominio.EstadoExitosa {      // Then the withdrawal is successful
				t.Fatalf("tx=%+v err=%v", tx, err)
			}
			if saldoDe(t, s, "Alice") != "34.56" {
				t.Error("saldo final incorrecto")
			}
		})
		t.Run("Scenario: Declined withdrawal in excess of balance", func(t *testing.T) {
			s, _ := nuevoSvc(t)                                        // Given Hamza has 198.76
			tx, err := s.Ejecutar("Hamza", dominio.OpRetiro, "200.00") // When Hamza tries to withdraw 200.00
			if err != nil || tx.Estado != dominio.EstadoRechazada {    // Then the withdrawal is declined
				t.Fatalf("tx=%+v err=%v", tx, err)
			}
			if tx.Motivo != "Saldo insuficiente. Intente con un monto menor" {
				t.Errorf("mensaje US-02 incorrecto: %q", tx.Motivo)
			}
			if saldoDe(t, s, "Hamza") != "198.76" {
				t.Error("el saldo no debe cambiar")
			}
		})
	})
}

func TestFeatureDepositingCash(t *testing.T) {
	t.Run("Rule: Customers can only deposit amounts greater than zero", func(t *testing.T) {
		t.Run("Scenario: Successful deposit of a positive amount", func(t *testing.T) {
			s, _ := nuevoSvc(t)
			tx, err := s.Ejecutar("Alice", dominio.OpAbono, "100.00")
			if err != nil || tx.Estado != dominio.EstadoExitosa || saldoDe(t, s, "Alice") != "334.56" {
				t.Fatalf("tx=%+v err=%v", tx, err)
			}
		})
		for nombre, monto := range map[string]string{
			"Scenario: Rejected deposit of a negative amount": "-50.00",
			"Scenario: Rejected deposit of zero":              "0.00",
		} {
			t.Run(nombre, func(t *testing.T) {
				s, _ := nuevoSvc(t)
				tx, err := s.Ejecutar("Hamza", dominio.OpAbono, monto)
				if err == nil || tx.Estado != dominio.EstadoError || saldoDe(t, s, "Hamza") != "198.76" {
					t.Fatalf("tx=%+v err=%v", tx, err)
				}
			})
		}
	})
}

func TestFeatureTransactionIntegrity(t *testing.T) {
	casos := []struct {
		escenario string
		cuenta    string
		op        dominio.Operacion
		monto     string
		estado    dominio.Estado
		saldo     string
	}{
		{"Successful operation is logged", "Alice", dominio.OpRetiro, "200.00", dominio.EstadoExitosa, "34.56"},
		{"Declined operation is logged and leaves the balance untouched", "Hamza", dominio.OpRetiro, "200.00", dominio.EstadoRechazada, "198.76"},
		{"Invalid operation is logged as error", "Pedro", dominio.OpRetiro, "-20.00", dominio.EstadoError, "300.00"},
	}
	for _, tc := range casos {
		t.Run("Scenario: "+tc.escenario, func(t *testing.T) {
			s, _ := nuevoSvc(t)
			_, _ = s.Ejecutar(tc.cuenta, tc.op, tc.monto)
			h, _ := s.Historial(tc.cuenta)
			if len(h) != 1 || h[0].Estado != tc.estado {
				t.Fatalf("log: %+v", h)
			}
			if saldoDe(t, s, tc.cuenta) != tc.saldo {
				t.Errorf("saldo = %s, esperaba %s", saldoDe(t, s, tc.cuenta), tc.saldo)
			}
		})
	}
	// "Scenario: Failure while persisting rolls everything back" -> TestAtomicidadRollbackSiFallaElLog
}
