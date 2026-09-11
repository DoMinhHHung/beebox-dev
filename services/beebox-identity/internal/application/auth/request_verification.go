package auth

import (
	"context"
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
	Code           string
	ExpiresAt      time.Time
}

type RequestVerificationService struct {
	verifications VerificationRepository
	clock         Clock
	ttl           time.Duration
	randReader    io.Reader
}

func NewRequestVerificationService(verifications VerificationRepository, clock Clock) *RequestVerificationService {
	return &RequestVerificationService{
		verifications: verifications,
		clock:         clock,
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

	code, err := generateNumericCode(s.randReader, verificationCodeDigits)
	if err != nil {
		return RequestVerificationResult{}, apperror.Wrap(apperror.CodeInternal, "verification code generation failed", err)
	}
	codeHash := hashVerificationCode(code)

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

	return RequestVerificationResult{
		VerificationID: v.ID(),
		Code:           code,
		ExpiresAt:      v.ExpiresAt(),
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

func hashVerificationCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
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