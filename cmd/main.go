package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	httphandler "github.com/lucasalmeron/backtabgo/internal/http"
	storage "github.com/lucasalmeron/backtabgo/pkg/storage"
)

func main() {

	err := storage.NewMongoDBConnection()
	if err != nil {
		log.Fatal("MongoDb Connection Error: %v", err)
	}

	var addr string

	if os.Getenv("PORT") != "" {
		addr = ":" + os.Getenv("PORT")
	} else {
		addr = "127.0.0.1:3500"
	}

	srv := &http.Server{
		Handler: httphandler.InitHttpHandler(),
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout:   15 * time.Second,
		ReadTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MiB
	}

	// Canal para señalar conexiones inactivas cerradas.
	conxCerradas := make(chan struct{})
	// Lanzamos goroutine para esperar señal y llamar Shutdown.
	go waitForShutdown(conxCerradas, srv)

	fmt.Printf("Server started on %s. CTRL+C for shutdown.\n", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("ListenAndServe Error: %v", err)
	}
	// Esperamos a que el shut down termine al cerrar todas las conexiones.
	<-conxCerradas
	fmt.Println("Shutdown Success.")
}

// waitForShutdown para detectar señales de interrupción al proceso y hacer Shutdown.
func waitForShutdown(conxCerradas chan struct{}, srv *http.Server) {
	// Canal para recibir señal de interrupción.
	sigint := make(chan os.Signal, 1)
	// Escuchamos por una señal de interrupción del OS (SIGINT).
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	// Si llegamos aquí, recibimos la señal, iniciamos shut down.
	// Noten se puede usar un Context para posible límite de tiempo.
	fmt.Println("\nShutdown started...")
	// Límite de tiempo para el Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		// Error aquí tiene que ser cerrando conexiones.
		log.Printf("Shutdown Error: %v", err)
	}

	// Cerramos el canal, señalando conexiones ya cerradas.
	close(conxCerradas)
}
