package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	httphandler "github.com/lucasalmeron/backtabgo/internal/http"
	mongostorage "github.com/lucasalmeron/backtabgo/pkg/storage/mongo"
)

var (
	mongoURI      = os.Getenv("MONGODB_URI")
	mongoDataBase = os.Getenv("MONGODB_DB")
)

type Server struct {
	srv    *http.Server
	router *mux.Router
	Addr   string
}

func (srv *Server) Init() {

	srv.Addr = ":" + os.Getenv("PORT")

	if os.Getenv("PORT") == "" {
		srv.Addr = "127.0.0.1:3500"
	}

	srv.router = mux.NewRouter().StrictSlash(true)

	// Only matches if domain is "www.example.com".
	//router.Host("www.example.com")
	httphandler.InitRoomHandler(srv.router)
	httphandler.InitDeckHandler(srv.router)
	httphandler.InitCardHandler(srv.router)

	srv.srv = &http.Server{
		Handler:        srv.router,
		Addr:           srv.Addr,
		WriteTimeout:   15 * time.Second,
		ReadTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MiB
	}
}

func (srv *Server) ConnectMongoDB() error {
	return mongostorage.NewMongoDBConnection(mongoURI, mongoDataBase)
}

func (s *Server) StartAndListen() {
	go s.waitForShutdown()

	fmt.Printf("Server started on %s. CTRL+C for shutdown.\n", s.Addr)
	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("ListenAndServe Error: ", err)
	}

	fmt.Println("Shutdown Success.")
}

func (s *Server) waitForShutdown() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	fmt.Println("\nShutdown started...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown Error: %v", err)
	}
}
