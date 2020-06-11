package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "127.0.0.1:3500", "Dirección IP y puerto")

var upgrader = websocket.Upgrader{} // use default options

func main() {

	srv := &http.Server{
		//Handler: router,
		Addr: *addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout:   15 * time.Second,
		ReadTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MiB
	}

	http.HandleFunc("/echo", echo)

	// Canal para señalar conexiones inactivas cerradas.
	conxCerradas := make(chan struct{})
	// Lanzamos goroutine para esperar señal y llamar Shutdown.
	go waitForShutdown(conxCerradas, srv)

	// Lanzamos el Server y estamos pendientes por si hay shut down.
	fmt.Printf("Servidor HTTPS en puerto %s listo. CTRL+C para detener.\n", *addr)
	/*
		// Certificado y key para producción
		cert, key := "tls/prod/cert.pem", "tls/prod/key.pem"
		if *dev {
			// Archivos para certificado y key en modo de desarrollo local, generados por
			// /usr/local/go/src/crypto/tls/generate_cert.go --host localhost
			cert, key = "tls/dev/cert.pem", "tls/dev/key.pem"
		}*/
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error iniciando el Server. Posible conflicto de puerto, permisos, etc.
		log.Fatal("Error durante ListenAndServe: %v", err)
	}

	// Esperamos a que el shut down termine al cerrar todas las conexiones.
	<-conxCerradas
	fmt.Println("Shut down del servidor HTTPS completado exitosamente.")
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
	fmt.Println("\nShut down del servidor HTTPS iniciado...")
	// Límite de tiempo para el Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		// Error aquí tiene que ser cerrando conexiones.
		log.Printf("Error durante Shutdown: %v", err)
	}

	// Cerramos el canal, señalando conexiones ya cerradas.
	close(conxCerradas)
}

func echo(w http.ResponseWriter, r *http.Request) {

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
