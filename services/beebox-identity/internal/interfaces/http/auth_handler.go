package http

import (
	"net/http"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
)

type authHandler struct {
	signUp               *auth.SignUpService
	signIn               *auth.SignInService
	revokeSession        *auth.RevokeSessionService
	requestVerification  *auth.RequestVerificationService
	verify               *auth.VerifyService
	requestPasswordReset *auth.RequestPasswordResetService
	resetPassword        *auth.ResetPasswordService
}

type signUpRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type signUpResponse struct {
	UserID string `json:"user_id"`
}

func (h *authHandler) handleSignUp(w http.ResponseWriter, r *http.Request) {
	var req signUpRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	result, err := h.signUp.SignUp(r.Context(), auth.SignUpInput{
		Identifier: req.Identifier,
		Password:   req.Password,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, signUpResponse{UserID: result.UserID.String()})
}

type signInRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type signInResponse struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (h *authHandler) handleSignIn(w http.ResponseWriter, r *http.Request) {
	var req signInRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	result, err := h.signIn.SignIn(r.Context(), auth.SignInInput{
		Identifier: req.Identifier,
		Password:   req.Password,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, signInResponse{
		UserID:    result.UserID.String(),
		SessionID: result.SessionID,
		ExpiresAt: result.ExpiresAt,
	})
}

func (h *authHandler) handleSignOut(w http.ResponseWriter, r *http.Request) {
	token, err := extractBearerToken(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.revokeSession.RevokeSession(r.Context(), auth.RevokeSessionInput{
		SessionID: token,
	}); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type requestVerificationRequest struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	Target string `json:"target"`
}

type requestVerificationResponse struct {
	VerificationID string    `json:"verification_id"`
	ExpiresAt      time.Time `json:"expires_at"`
}

func (h *authHandler) handleRequestVerification(w http.ResponseWriter, r *http.Request) {
	var req requestVerificationRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	result, err := h.requestVerification.RequestVerification(r.Context(), auth.RequestVerificationInput{
		UserID: req.UserID,
		Type:   req.Type,
		Target: req.Target,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, requestVerificationResponse{
		VerificationID: result.VerificationID,
		ExpiresAt:      result.ExpiresAt,
	})
}

type verifyRequest struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	Target string `json:"target"`
	Code   string `json:"code"`
}

type verifyResponse struct {
	VerificationID string `json:"verification_id"`
	UserID         string `json:"user_id"`
	Type           string `json:"type"`
	Target         string `json:"target"`
}

func (h *authHandler) handleVerify(w http.ResponseWriter, r *http.Request) {
	var req verifyRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	result, err := h.verify.Verify(r.Context(), auth.VerifyInput{
		UserID: req.UserID,
		Type:   req.Type,
		Target: req.Target,
		Code:   req.Code,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, verifyResponse{
		VerificationID: result.VerificationID,
		UserID:         result.UserID.String(),
		Type:           string(result.Type),
		Target:         result.Target,
	})
}

type requestPasswordResetRequest struct {
	Identifier string `json:"identifier"`
}

type requestPasswordResetResponse struct {
	Status string `json:"status"`
}

func (h *authHandler) handleRequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req requestPasswordResetRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	result, err := h.requestPasswordReset.RequestPasswordReset(r.Context(), auth.RequestPasswordResetInput{
		Identifier: req.Identifier,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, requestPasswordResetResponse{Status: result.Status})
}

type resetPasswordRequest struct {
	ResetID     string `json:"reset_id"`
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type resetPasswordResponse struct {
	UserID string `json:"user_id"`
}

func (h *authHandler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	result, err := h.resetPassword.ResetPassword(r.Context(), auth.ResetPasswordInput{
		ResetID:     req.ResetID,
		Token:       req.Token,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resetPasswordResponse{UserID: result.UserID.String()})
}
