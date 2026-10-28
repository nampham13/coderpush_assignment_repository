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
	defaultAddress := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		defaultAddress = ":" + port
	}
	address := flag.String("addr", defaultAddress, "HTTP listen address (defaults to $PORT when set)")
	database := flag.String("data", "./data/catalog.db", "path to the local SQLite database")
	seed := flag.String("seed", "", "load products from a SQLite mock database and exit")
	static := flag.String("static", "", "directory of built web UI files to serve at /")
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
	handler := httpapi.NewHandlerFactory(store).Create()
	if *static != "" {
		handler = withStaticFiles(handler, *static)
	}
	server := &http.Server{
		Addr: *address, Handler: handler,
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

// withStaticFiles serves the built web UI at / and routes API paths to api.
func withStaticFiles(api http.Handler, directory string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(directory)))
	for _, path := range []string{"/health", "/catalog", "/updates", "/updates/", "/demo"} {
		mux.Handle(path, api)
	}
	return mux
}
