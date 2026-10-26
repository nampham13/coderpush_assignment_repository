package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	address := flag.String("addr", ":8080", "HTTP listen address")
	database := flag.String("data", "./data/catalog.json", "path to the local catalog database")
	flag.Parse()

	store, err := OpenStore(*database)
	if err != nil {
		log.Fatalf("open catalog: %v", err)
	}
	server := &http.Server{
		Addr: *address, Handler: NewHandler(store),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second,
	}

	stopped := make(chan os.Signal, 1)
	signal.Notify(stopped, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stopped
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}
	}()

	log.Printf("supplier catalog listening on %s; database: %s", *address, *database)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve HTTP: %v", err)
	}
}
