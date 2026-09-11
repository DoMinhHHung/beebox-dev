package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/passwordreset"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/session"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/user"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/verification"
)

type httpFakeClock struct {
	now time.Time
}

func (c *httpFakeClock) Now() time.Time {
	return c.now
}

type httpFakeUserRepo struct {
	findUser  user.User
	findErr   error
	createErr error
	created   user.User
}

func (f *httpFakeUserRepo) Create(_ context.Context, value user.User) error {
	f.created = value
	return f.createErr
}

func (f *httpFakeUserRepo) FindByIdentifier(context.Context, identity.Identifier) (user.User, error) {
	return f.findUser, f.findErr
}

type httpFakeCredentialRepo struct {
	createErr error
	findValue credential.Credential
	findErr   error
	updateErr error
	updated   credential.Credential
}

func (f *httpFakeCredentialRepo) Create(context.Context, credential.Credential) error {
	return f.createErr
}

func (f *httpFakeCredentialRepo) FindByUserID(context.Context, identity.Identifier) (credential.Credential, error) {
	return f.findValue, f.findErr
}

func (f *httpFakeCredentialRepo) Update(_ context.Context, value credential.Credential) error {
	f.updated = value
	return f.updateErr
}

type httpFakeHasher struct {
	hashValue string
	hashErr   error
	verifyErr error
}

func (f *httpFakeHasher) Hash(context.Context, string) (string, error) {
	return f.hashValue, f.hashErr
}

func (f *httpFakeHasher) Verify(context.Context, string, string) error {
	return f.verifyErr
}

type httpFakeSessionRepo struct {
	findValue session.Session
	findErr   error
	revokeErr error
}

func (f *httpFakeSessionRepo) Create(context.Context, session.Session) error {
	return nil
}

func (f *httpFakeSessionRepo) FindByID(context.Context, string) (session.Session, error) {
	return f.findValue, f.findErr
}

func (f *httpFakeSessionRepo) Revoke(context.Context, session.Session) error {
	return f.revokeErr
}

type httpFakeVerificationRepo struct {
	createErr error
	findValue verification.Verification
	findErr   error
	markErr   error
	marked    verification.Verification
}

func (f *httpFakeVerificationRepo) Create(context.Context, verification.Verification) error {
	return f.createErr
}

func (f *httpFakeVerificationRepo) FindPending(context.Context, identity.Identifier, verification.Type, string) (verification.Verification, error) {
	return f.findValue, f.findErr
}

func (f *httpFakeVerificationRepo) MarkUsed(_ context.Context, value verification.Verification) error {
	f.marked = value
	return f.markErr
}

type httpFakePasswordResetRepo struct {
	createErr error
	findValue passwordreset.PasswordReset
	findErr   error
	markErr   error
	marked    passwordreset.PasswordReset
}

func (f *httpFakePasswordResetRepo) Create(context.Context, passwordreset.PasswordReset) error {
	return f.createErr
}

func (f *httpFakePasswordResetRepo) FindByID(context.Context, string) (passwordreset.PasswordReset, error) {
	return f.findValue, f.findErr
}

func (f *httpFakePasswordResetRepo) FindPending(context.Context, identity.Identifier) (passwordreset.PasswordReset, error) {
	return f.findValue, f.findErr
}

func (f *httpFakePasswordResetRepo) MarkUsed(_ context.Context, value passwordreset.PasswordReset) error {
	f.marked = value
	return f.markErr
}

func testRouter(t *testing.T, deps Dependencies) *http.ServeMux {
	t.Helper()
	return NewRouter(deps)
}

func postJSON(t *testing.T, mux *http.ServeMux, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else if s, ok := body.(string); ok {
		reader = bytes.NewReader([]byte(s))
	} else {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(http.MethodPost, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	NewRouter(Dependencies{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected content type")
	}
	if rec.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected body %q", rec.Body.String())
	}
}

func TestSignUpSuccess(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	users := &httpFakeUserRepo{findErr: auth.ErrNotFound}
	credentials := &httpFakeCredentialRepo{}
	hasher := &httpFakeHasher{hashValue: "hash"}
	svc := auth.NewSignUpService(users, credentials, hasher, &httpFakeClock{now: now})
	mux := testRouter(t, Dependencies{SignUp: svc})

	rec := postJSON(t, mux, "/auth/signup", map[string]string{
		"identifier": "user-1",
		"password":   "secret",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body signUpResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.UserID != "user-1" {
		t.Fatalf("unexpected user_id %q", body.UserID)
	}
}

func TestSignUpValidation(t *testing.T) {
	svc := auth.NewSignUpService(&httpFakeUserRepo{}, &httpFakeCredentialRepo{}, &httpFakeHasher{}, &httpFakeClock{now: time.Now().UTC()})
	mux := testRouter(t, Dependencies{SignUp: svc})
	rec := postJSON(t, mux, "/auth/signup", map[string]string{"identifier": "", "password": ""})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "VALIDATION")
}

func TestSignInSuccess(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	u, _ := user.New(userID, now)
	c, _ := credential.NewPassword(userID, "hash", now)
	users := &httpFakeUserRepo{findUser: u}
	credentials := &httpFakeCredentialRepo{findValue: c}
	hasher := &httpFakeHasher{}
	svc := auth.NewSignInService(users, credentials, hasher, clock)
	mux := testRouter(t, Dependencies{SignIn: svc})

	rec := postJSON(t, mux, "/auth/signin", map[string]string{
		"identifier": "user-1",
		"password":   "secret",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body signInResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.UserID != "user-1" {
		t.Fatalf("unexpected user_id")
	}
}

func TestSignInUnauthenticated(t *testing.T) {
	users := &httpFakeUserRepo{findErr: auth.ErrNotFound}
	svc := auth.NewSignInService(users, &httpFakeCredentialRepo{}, &httpFakeHasher{})
	mux := testRouter(t, Dependencies{SignIn: svc})
	rec := postJSON(t, mux, "/auth/signin", map[string]string{
		"identifier": "missing",
		"password":   "secret",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "UNAUTHENTICATED")
}

func TestSignOutSuccess(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	s, _ := session.New("session-1", userID, now.Add(-time.Hour), now.Add(time.Hour))
	sessions := &httpFakeSessionRepo{findValue: s}
	svc := auth.NewRevokeSessionService(sessions, &httpFakeClock{now: now})
	mux := testRouter(t, Dependencies{RevokeSession: svc})

	rec := postJSON(t, mux, "/auth/signout", map[string]string{"session_id": "session-1"})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRequestVerificationSuccess(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := auth.NewRequestVerificationService(&httpFakeVerificationRepo{}, &httpFakeClock{now: now})
	mux := testRouter(t, Dependencies{RequestVerification: svc})

	rec := postJSON(t, mux, "/auth/verification/request", map[string]string{
		"user_id": "user-1",
		"type":    "email",
		"target":  "user@example.com",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body requestVerificationResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.VerificationID == "" || body.Code == "" {
		t.Fatalf("expected verification payload, got %+v", body)
	}
}

func TestVerifyFailure(t *testing.T) {
	svc := auth.NewVerifyService(&httpFakeVerificationRepo{findErr: auth.ErrNotFound}, &httpFakeClock{now: time.Now().UTC()})
	mux := testRouter(t, Dependencies{Verify: svc})
	rec := postJSON(t, mux, "/auth/verification/verify", map[string]string{
		"user_id": "user-1",
		"type":    "email",
		"target":  "a@b.com",
		"code":    "123456",
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "NOT_FOUND")
}

func TestRequestPasswordResetKnownUser(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	u, _ := user.New(userID, now)
	users := &httpFakeUserRepo{findUser: u}
	resets := &httpFakePasswordResetRepo{}
	svc := auth.NewRequestPasswordResetService(users, resets, &httpFakeClock{now: now})
	mux := testRouter(t, Dependencies{RequestPasswordReset: svc})

	rec := postJSON(t, mux, "/auth/password-reset/request", map[string]string{"identifier": "user-1"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body requestPasswordResetResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ResetID == "" || body.Token == "" {
		t.Fatalf("expected reset payload, got %+v", body)
	}
}

func TestRequestPasswordResetEnumerationProtection(t *testing.T) {
	users := &httpFakeUserRepo{findErr: auth.ErrNotFound}
	svc := auth.NewRequestPasswordResetService(users, &httpFakePasswordResetRepo{}, &httpFakeClock{now: time.Now().UTC()})
	mux := testRouter(t, Dependencies{RequestPasswordReset: svc})

	rec := postJSON(t, mux, "/auth/password-reset/request", map[string]string{"identifier": "missing"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for unknown identity, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body requestPasswordResetResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "accepted" {
		t.Fatalf("expected accepted status, got %+v", body)
	}
	if body.Token != "" || body.ResetID != "" {
		t.Fatal("must not leak reset material for unknown identity")
	}
}

func TestResetPasswordSuccess(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	token := "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
	sum := authHashToken(token)
	reset, _ := passwordreset.New("reset-1", userID, sum, now.Add(-time.Minute), now.Add(time.Hour))
	cred, _ := credential.NewPassword(userID, "old", now.Add(-time.Hour))
	resets := &httpFakePasswordResetRepo{findValue: reset}
	credentials := &httpFakeCredentialRepo{findValue: cred}
	hasher := &httpFakeHasher{hashValue: "new-hash"}
	svc := auth.NewResetPasswordService(resets, credentials, hasher, &httpFakeClock{now: now})
	mux := testRouter(t, Dependencies{ResetPassword: svc})

	rec := postJSON(t, mux, "/auth/password-reset/reset", map[string]string{
		"reset_id":     "reset-1",
		"token":        token,
		"new_password": "new-secret",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body resetPasswordResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.UserID != "user-1" {
		t.Fatalf("unexpected user_id")
	}
}

func TestMalformedJSON(t *testing.T) {
	svc := auth.NewSignUpService(&httpFakeUserRepo{}, &httpFakeCredentialRepo{}, &httpFakeHasher{}, &httpFakeClock{now: time.Now().UTC()})
	mux := testRouter(t, Dependencies{SignUp: svc})
	rec := postJSON(t, mux, "/auth/signup", "{not-json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "VALIDATION")
}

func TestApplicationErrorMappingConflict(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	u, _ := user.New(userID, now)
	users := &httpFakeUserRepo{findUser: u}
	svc := auth.NewSignUpService(users, &httpFakeCredentialRepo{}, &httpFakeHasher{hashValue: "h"}, &httpFakeClock{now: now})
	mux := testRouter(t, Dependencies{SignUp: svc})
	rec := postJSON(t, mux, "/auth/signup", map[string]string{"identifier": "user-1", "password": "secret"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "CONFLICT")
}

func TestInternalErrorGenericMessage(t *testing.T) {
	users := &httpFakeUserRepo{findErr: errors.New("db exploded")}
	svc := auth.NewSignInService(users, &httpFakeCredentialRepo{}, &httpFakeHasher{})
	mux := testRouter(t, Dependencies{SignIn: svc})
	rec := postJSON(t, mux, "/auth/signin", map[string]string{"identifier": "user-1", "password": "secret"})
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 for dependency failure, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "DEPENDENCY_FAILURE")
}

func TestNoPanicOnEmptyBody(t *testing.T) {
	svc := auth.NewSignUpService(&httpFakeUserRepo{}, &httpFakeCredentialRepo{}, &httpFakeHasher{}, &httpFakeClock{now: time.Now().UTC()})
	mux := testRouter(t, Dependencies{SignUp: svc})
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", strings.NewReader(""))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, code string) {
	t.Helper()
	var env errorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode error envelope: %v body=%s", err, rec.Body.String())
	}
	if env.Error.Code != code {
		t.Fatalf("expected code %s, got %s message=%s", code, env.Error.Code, env.Error.Message)
	}
	if env.Error.Message == "" {
		t.Fatal("expected non-empty message")
	}
}

func authHashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
