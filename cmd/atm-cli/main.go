// atm-cli: Cajero Automático en línea de comandos (interactivo o por flags).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"prj_withdrawal/internal/almacen"
	"prj_withdrawal/internal/cli"
	"prj_withdrawal/internal/dominio"
	"prj_withdrawal/internal/servicio"
)

func main() {
	datos := flag.String("datos", "data/cuentas.json", "Fuente de datos JSON de cuentas")
	logf := flag.String("log", "data/transacciones.json", "Log de transacciones JSON")
	semilla := flag.String("semilla", "data/cuentas.seed.json", "Datos iniciales si -datos no existe")
	nombre := flag.String("nombre", "", "Titular (modo no interactivo)")
	op := flag.String("op", "", "Operación no interactiva: 'retiro' o 'abono'")
	monto := flag.String("monto", "", "Monto de la operación, p. ej. 150.50")
	historial := flag.Bool("historial", false, "Imprime el log de transacciones en JSON")
	flag.Parse()

	alm := almacen.JSON{RutaCuentas: *datos, RutaLog: *logf}
	if err := alm.InicializarDesdeSemilla(*semilla); err != nil {
		salir(err)
	}
	svc, err := servicio.Nuevo(alm)
	if err != nil {
		salir(err)
	}

	switch {
	case *historial:
		h, err := svc.Historial(*nombre)
		if err != nil {
			salir(err)
		}
		imprimirJSON(h)
	case *op != "":
		tipo := dominio.Operacion(map[string]string{"retiro": "RETIRO", "abono": "ABONO"}[*op])
		tx, err := svc.Ejecutar(*nombre, tipo, *monto)
		imprimirJSON(tx)
		if err != nil {
			os.Exit(2)
		}
	default:
		if err := cli.Menu(svc, os.Stdin, os.Stdout); err != nil {
			salir(err)
		}
	}
}

func imprimirJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func salir(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
