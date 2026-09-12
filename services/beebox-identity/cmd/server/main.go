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

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/infrastructure/config"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/infrastructure/delivery"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/infrastructure/postgres"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/infrastructure/security"
	interfaceshttp "github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/interfaces/http"
)

func main() {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		log.Fatalf("beebox-identity: invalid configuration: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("beebox-identity: postgres connection failed: %v", err)
	}
	defer pool.Close()

	users := postgres.NewUserRepository(pool)
	credentials := postgres.NewCredentialRepository(pool)
	sessions := postgres.NewSessionRepository(pool)
	verifications := postgres.NewVerificationRepository(pool)
	passwordResets := postgres.NewPasswordResetRepository(pool)
	tx := postgres.NewTransactor(pool)
	hasher := security.NewBcryptPasswordHasher()
	clock := postgres.SystemClock{}

	mailer := delivery.NewSMTPMailer(delivery.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUser,
		Password: cfg.SMTPPass,
		From:     cfg.SMTPFrom,
	})
	sms := delivery.NewTwilioSMSSender(delivery.TwilioSMSConfig{
		AccountSID: cfg.TwilioAccountSID,
		AuthToken:  cfg.TwilioAuthToken,
		From:       cfg.TwilioFromNumber,
	})

	signUp := auth.NewSignUpService(users, credentials, hasher, clock, tx)
	signIn := auth.NewSignInService(users, credentials, sessions, hasher, clock)
	revokeSession := auth.NewRevokeSessionService(sessions, clock)
	authenticateSession := auth.NewAuthenticateSessionService(sessions, clock)
	requestVerification := auth.NewRequestVerificationService(users, verifications, mailer, sms, clock, cfg.VerificationCodeSecret)
	verify := auth.NewVerifyService(verifications, clock, cfg.VerificationCodeSecret)
	requestPasswordReset := auth.NewRequestPasswordResetService(users, passwordResets, mailer, clock)
	resetPassword := auth.NewResetPasswordService(passwordResets, credentials, sessions, hasher, clock, tx)

	router := interfaceshttp.NewRouter(interfaceshttp.Dependencies{
		SignUp:               signUp,
		SignIn:               signIn,
		RevokeSession:        revokeSession,
		AuthenticateSession:  authenticateSession,
		RequestVerification:  requestVerification,
		Verify:               verify,
		RequestPasswordReset: requestPasswordReset,
		ResetPassword:        resetPassword,
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

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
