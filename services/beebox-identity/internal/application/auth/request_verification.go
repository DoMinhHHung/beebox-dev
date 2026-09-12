package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/verification"
)

const (
	defaultVerificationTTL = 15 * time.Minute
	verificationCodeDigits = 6
)

type RequestVerificationInput struct {
	UserID string
	Type   string
	Target string
}

type RequestVerificationResult struct {
	VerificationID string
	ExpiresAt      time.Time
}

type RequestVerificationService struct {
	users         UserRepository
	verifications VerificationRepository
	email         VerificationMailer
	sms           VerificationSMSSender
	clock         Clock
	codeSecret    string
	ttl           time.Duration
	randReader    io.Reader
}

func NewRequestVerificationService(
	users UserRepository,
	verifications VerificationRepository,
	email VerificationMailer,
	sms VerificationSMSSender,
	clock Clock,
	codeSecret string,
) *RequestVerificationService {
	return &RequestVerificationService{
		users:         users,
		verifications: verifications,
		email:         email,
		sms:           sms,
		clock:         clock,
		codeSecret:    codeSecret,
		ttl:           defaultVerificationTTL,
		randReader:    rand.Reader,
	}
}

func (s *RequestVerificationService) RequestVerification(ctx context.Context, input RequestVerificationInput) (RequestVerificationResult, error) {
	userID, err := identity.NewIdentifier(input.UserID)
	if err != nil {
		return RequestVerificationResult{}, apperror.Wrap(apperror.CodeValidation, "invalid request verification input", err)
	}
	vtype := verification.Type(strings.TrimSpace(input.Type))
	if !vtype.Valid() {
		return RequestVerificationResult{}, apperror.New(apperror.CodeValidation, "invalid request verification input")
	}
	target := strings.TrimSpace(input.Target)
	if target == "" {
		return RequestVerificationResult{}, apperror.New(apperror.CodeValidation, "invalid request verification input")
	}

	_, err = s.users.FindByIdentifier(ctx, userID)
	switch {
	case err == nil:
	case errors.Is(err, ErrNotFound):
		return s.publicSuccess(ctx)
	case err != nil:
		return RequestVerificationResult{}, translateRepositoryError(err)
	}

	code, err := generateNumericCode(s.randReader, verificationCodeDigits)
	if err != nil {
		return RequestVerificationResult{}, apperror.Wrap(apperror.CodeInternal, "verification code generation failed", err)
	}
	codeHash := hashVerificationCode(s.codeSecret, code)

	id, err := generateVerificationID(s.randReader)
	if err != nil {
		return RequestVerificationResult{}, apperror.Wrap(apperror.CodeInternal, "verification id generation failed", err)
	}

	now := s.clock.Now()
	expiresAt := now.Add(s.ttl)
	v, err := verification.New(id, userID, vtype, target, codeHash, now, expiresAt)
	if err != nil {
		return RequestVerificationResult{}, translateVerificationDomainError(err)
	}

	if err := s.verifications.Create(ctx, v); err != nil {
		return RequestVerificationResult{}, translateRepositoryError(err)
	}

	message := VerificationDeliveryMessage{
		UserID: userID.String(),
		Type:   string(vtype),
		Target: target,
		Code:   code,
	}
	switch vtype {
	case verification.TypeEmail:
		if s.email != nil {
			_ = s.email.SendVerificationEmail(ctx, message)
		}
	case verification.TypePhone:
		if s.sms != nil {
			_ = s.sms.SendVerificationSMS(ctx, message)
		}
	}

	return RequestVerificationResult{
		VerificationID: v.ID(),
		ExpiresAt:      v.ExpiresAt(),
	}, nil
}

func (s *RequestVerificationService) publicSuccess(ctx context.Context) (RequestVerificationResult, error) {
	id, err := generateVerificationID(s.randReader)
	if err != nil {
		return RequestVerificationResult{}, apperror.Wrap(apperror.CodeInternal, "verification id generation failed", err)
	}
	now := s.clock.Now()
	return RequestVerificationResult{
		VerificationID: id,
		ExpiresAt:      now.Add(s.ttl),
	}, nil
}

func generateNumericCode(r io.Reader, digits int) (string, error) {
	if digits <= 0 {
		return "", fmt.Errorf("invalid digits")
	}
	max := 1
	for i := 0; i < digits; i++ {
		max *= 10
	}
	buf := make([]byte, 4)
	for {
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", err
		}
		n := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
		limit := (^uint32(0) / uint32(max)) * uint32(max)
		if n >= limit {
			continue
		}
		return fmt.Sprintf("%0*d", digits, int(n)%max), nil
	}
}

func generateVerificationID(r io.Reader) (string, error) {
	buf := make([]byte, 16)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashVerificationCode(secret, code string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}

func translateVerificationDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidVerification):
		return apperror.Wrap(apperror.CodeValidation, "invalid verification", err)
	case errors.Is(err, domain.ErrVerificationUsed):
		return apperror.New(apperror.CodeConflict, "verification already used")
	case errors.Is(err, domain.ErrVerificationExpired):
		return apperror.New(apperror.CodeConflict, "verification expired")
	default:
		return apperror.Wrap(apperror.CodeInternal, "internal error", err)
	}
}
