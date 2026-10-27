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

	"example.com/supplier-catalog/internal/catalog/httpapi"
	seedpkg "example.com/supplier-catalog/internal/catalog/seed"
	"example.com/supplier-catalog/internal/catalog/service"
)

func main() {
	address := flag.String("addr", ":8080", "HTTP listen address")
	database := flag.String("data", "./data/catalog.db", "path to the local SQLite database")
	seed := flag.String("seed", "", "load products from a SQLite mock database and exit")
	flag.Parse()

	store, err := service.NewStoreFactory(*database).Create()
	if err != nil {
		log.Fatalf("open catalog: %v", err)
	}
	if *seed != "" {
		count, err := seedpkg.Import(store, *seed)
		if err != nil {
			log.Fatalf("seed catalog: %v", err)
		}
		log.Printf("seeded %d products from %s", count, *seed)
		return
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close catalog database: %v", err)
		}
	}()
	server := &http.Server{
		Addr: *address, Handler: httpapi.NewHandlerFactory(store).Create(),
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
