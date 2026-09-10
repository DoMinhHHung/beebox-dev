package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/infrastructure/config"
	interfaceshttp "github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/interfaces/http"
)

func main() {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		log.Fatalf("beebox-identity: invalid configuration: %v", err)
	}

	router := interfaceshttp.NewRouter()

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("beebox-identity: listening on %s", server.Addr)
		serveErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("beebox-identity: server error: %v", err)
		}
	case <-ctx.Done():
		stop()
		log.Println("beebox-identity: shutting down")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("beebox-identity: graceful shutdown failed: %v", err)
		}
	}
}
