package http

import (
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
)

type Dependencies struct {
	SignUp               *auth.SignUpService
	SignIn               *auth.SignInService
	RevokeSession        *auth.RevokeSessionService
	AuthenticateSession  *auth.AuthenticateSessionService
	RequestVerification  *auth.RequestVerificationService
	Verify               *auth.VerifyService
	RequestPasswordReset *auth.RequestPasswordResetService
	ResetPassword        *auth.ResetPasswordService
}

func NewRouter(deps Dependencies) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealthz)

	h := &authHandler{
		signUp:               deps.SignUp,
		signIn:               deps.SignIn,
		revokeSession:        deps.RevokeSession,
		requestVerification:  deps.RequestVerification,
		verify:               deps.Verify,
		requestPasswordReset: deps.RequestPasswordReset,
		resetPassword:        deps.ResetPassword,
	}

	mux.HandleFunc("POST /auth/signup", h.handleSignUp)
	mux.HandleFunc("POST /auth/signin", h.handleSignIn)
	mux.HandleFunc("POST /auth/verification/request", h.handleRequestVerification)
	mux.HandleFunc("POST /auth/verification/verify", h.handleVerify)
	mux.HandleFunc("POST /auth/password-reset/request", h.handleRequestPasswordReset)
	mux.HandleFunc("POST /auth/password-reset/reset", h.handleResetPassword)

	if deps.AuthenticateSession != nil {
		mux.HandleFunc("POST /auth/signout", requireAuthentication(deps.AuthenticateSession, h.handleSignOut))
	} else {
		mux.HandleFunc("POST /auth/signout", h.handleSignOut)
	}

	return mux
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
