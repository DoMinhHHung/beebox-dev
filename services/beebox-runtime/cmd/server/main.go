package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/authcap"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/projectresolve"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/infrastructure/config"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/infrastructure/identityclient"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/infrastructure/projectclient"
	interfaceshttp "github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/interfaces/http"
)

func main() {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		log.Fatalf("beebox-runtime: invalid configuration: %v", err)
	}

	projectHTTP := projectclient.New(cfg.ProjectBaseURL, cfg.InternalToken, cfg.HTTPTimeout, nil)
	identityHTTP := identityclient.New(cfg.IdentityBaseURL, cfg.HTTPTimeout, nil)
	resolve := projectresolve.NewService(projectHTTP, projectHTTP)
	sessions := authcap.NewService(identityHTTP)
	router := interfaceshttp.NewRouter(resolve, sessions)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("beebox-runtime: listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("beebox-runtime: server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("beebox-runtime: shutdown error: %v", err)
	}
}
