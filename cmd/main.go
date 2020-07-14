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
	mongostorage "github.com/lucasalmeron/backtabgo/pkg/storage/mongo"
)

var (
	mongoURI      = os.Getenv("MONGODB_URI")
	mongoDataBase = os.Getenv("MONGODB_DB")
	addr          = ":" + os.Getenv("PORT")
)

func main() {

	if os.Getenv("MONGODB_URI") == "" {
		mongoURI = fmt.Sprintf("mongodb://localhost:27017")
	}
	if os.Getenv("MONGODB_DB") == "" {
		mongoDataBase = "taboogame"
	}

	if os.Getenv("PORT") == "" {
		addr = "127.0.0.1:3500"
	}

	err := mongostorage.NewMongoDBConnection(mongoURI, mongoDataBase)
	if err != nil {
		log.Fatal("MongoDb Connection Error: ", err)
	}

	srv := &http.Server{
		Handler: httphandler.Init(),
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout:   15 * time.Second,
		ReadTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MiB
	}

	go waitForShutdown(srv)

	fmt.Printf("Server started on %s. CTRL+C for shutdown.\n", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("ListenAndServe Error: ", err)
	}

	fmt.Println("Shutdown Success.")
}

func waitForShutdown(srv *http.Server) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	fmt.Println("\nShutdown started...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown Error: %v", err)
	}
}
