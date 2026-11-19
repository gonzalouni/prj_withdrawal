// Package cli implementa el CLI interactivo (codeLAB Semana-7c) sobre el servicio.
package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"prj_withdrawal/internal/dominio"
	"prj_withdrawal/internal/servicio"
)

// Menu corre el menú interactivo leyendo de in y escribiendo en out.
func Menu(s *servicio.Servicio, in io.Reader, out io.Writer) error {
	sc := bufio.NewScanner(in)
	leer := func(prompt string) (string, bool) {
		fmt.Fprint(out, prompt)
		if !sc.Scan() {
			return "", false
		}
		return strings.TrimSpace(sc.Text()), true
	}

	fmt.Fprintln(out, "=== Cajero Automático (prj_withdrawal v3.0.0) ===")
	var titular string
	for {
		n, ok := leer("Titular de la cuenta: ")
		if !ok {
			return nil
		}
		if _, err := s.Saldo(n); err != nil {
			fmt.Fprintln(out, "Cuenta no encontrada. Intente nuevamente.")
			continue
		}
		titular = n
		break
	}

	for {
		fmt.Fprintf(out, "\n[%s] 1) Consultar saldo  2) Retirar efectivo  3) Abonar efectivo  4) Historial  0) Salir\n", titular)
		op, ok := leer("Opción: ")
		if !ok {
			return nil
		}
		switch op {
		case "0":
			fmt.Fprintln(out, "Gracias por usar el cajero.")
			return nil
		case "1":
			v, _ := s.Saldo(titular)
			fmt.Fprintf(out, "Saldo disponible: S/ %s\n", v)
		case "2", "3":
			tipo, verbo := dominio.OpRetiro, "retirar"
			if op == "3" {
				tipo, verbo = dominio.OpAbono, "abonar"
			}
			monto, ok := leer("Monto a " + verbo + ": ")
			if !ok {
				return nil
			}
			tx, _ := s.Ejecutar(titular, tipo, monto)
			ImprimirTx(out, tx)
		case "4":
			h, err := s.Historial(titular)
			if err != nil {
				fmt.Fprintln(out, "No se pudo leer el historial:", err)
				continue
			}
			for _, t := range h {
				fmt.Fprintf(out, "  %s %s %-6s S/ %9s %-9s %s\n", t.ID, t.Fecha, t.Tipo, t.Monto, t.Estado, t.Motivo)
			}
			if len(h) == 0 {
				fmt.Fprintln(out, "  (sin transacciones)")
			}
		default:
			fmt.Fprintln(out, "Opción inválida.")
		}
	}
}

// ImprimirTx muestra el resultado de una transacción en una línea.
func ImprimirTx(out io.Writer, tx dominio.Transaccion) {
	fmt.Fprintf(out, "Transacción %s [%s] %s S/ %s", tx.ID, tx.Estado, tx.Tipo, tx.Monto)
	if tx.Estado == dominio.EstadoExitosa {
		fmt.Fprintf(out, " | saldo %s -> %s\n", tx.SaldoAntes, tx.SaldoDespues)
		return
	}
	fmt.Fprintf(out, " | %s\n", tx.Motivo)
}
