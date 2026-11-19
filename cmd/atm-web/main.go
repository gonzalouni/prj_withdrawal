// atm-web: App WEB del Cajero Automático (API REST + interfaz embebida).
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"prj_withdrawal/internal/almacen"
	"prj_withdrawal/internal/api"
	"prj_withdrawal/internal/servicio"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "Dirección de escucha")
	datos := flag.String("datos", "data/cuentas.json", "Fuente de datos JSON de cuentas")
	logf := flag.String("log", "data/transacciones.json", "Log de transacciones JSON")
	semilla := flag.String("semilla", "data/cuentas.seed.json", "Datos iniciales si -datos no existe")
	flag.Parse()

	alm := almacen.JSON{RutaCuentas: *datos, RutaLog: *logf}
	if err := alm.InicializarDesdeSemilla(*semilla); err != nil {
		log.Fatal(err)
	}
	svc, err := servicio.Nuevo(alm)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              *addr,
		Handler:           api.Nuevo(svc),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	go func() {
		log.Printf("Cajero Automático en http://%s", *addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
