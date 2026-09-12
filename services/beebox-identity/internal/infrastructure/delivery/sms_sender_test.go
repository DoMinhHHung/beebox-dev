package delivery

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
)

func TestTwilioSMSSenderSuccess(t *testing.T) {
	var gotAuth string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	sender := NewTwilioSMSSender(TwilioSMSConfig{
		AccountSID: "ACtest",
		AuthToken:  "token",
		From:       "+15550001111",
		APIBaseURL: server.URL,
		HTTPClient: server.Client(),
	})

	err := sender.SendVerificationSMS(context.Background(), auth.VerificationDeliveryMessage{
		UserID: "user-1",
		Type:   "phone",
		Target: "+15551234567",
		Code:   "123456",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth == "" {
		t.Fatal("expected basic auth")
	}
	if !strings.Contains(gotBody, "To=%2B15551234567") {
		t.Fatalf("unexpected body: %s", gotBody)
	}
	if !strings.Contains(gotBody, "123456") {
		t.Fatalf("expected code in provider body: %s", gotBody)
	}
}

func TestTwilioSMSSenderProviderFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	sender := NewTwilioSMSSender(TwilioSMSConfig{
		AccountSID: "ACtest",
		AuthToken:  "token",
		From:       "+15550001111",
		APIBaseURL: server.URL,
		HTTPClient: server.Client(),
	})

	err := sender.SendVerificationSMS(context.Background(), auth.VerificationDeliveryMessage{
		Target: "+15551234567",
		Code:   "123456",
	})
	if err == nil {
		t.Fatal("expected provider failure")
	}
}

func TestTwilioSMSSenderMissingConfig(t *testing.T) {
	sender := NewTwilioSMSSender(TwilioSMSConfig{})
	err := sender.SendVerificationSMS(context.Background(), auth.VerificationDeliveryMessage{
		Target: "+15551234567",
		Code:   "123456",
	})
	if err == nil || !strings.Contains(err.Error(), "sms configuration incomplete") {
		t.Fatalf("expected configuration incomplete, got %v", err)
	}
}

func TestTwilioSMSSenderInvalidPayload(t *testing.T) {
	sender := NewTwilioSMSSender(TwilioSMSConfig{
		AccountSID: "ACtest",
		AuthToken:  "token",
		From:       "+15550001111",
	})
	err := sender.SendVerificationSMS(context.Background(), auth.VerificationDeliveryMessage{})
	if err == nil {
		t.Fatal("expected invalid payload error")
	}
}

func TestTwilioSMSSenderContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	sender := NewTwilioSMSSender(TwilioSMSConfig{
		AccountSID: "ACtest",
		AuthToken:  "token",
		From:       "+15550001111",
		APIBaseURL: server.URL,
		HTTPClient: server.Client(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sender.SendVerificationSMS(ctx, auth.VerificationDeliveryMessage{
		Target: "+15551234567",
		Code:   "123456",
	})
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
