package security

import (
	"testing"

	"emergencycomms/internal/model"
)

func TestChannelAuthenticatorLifecycle(t *testing.T) {
	auth := NewChannelAuthenticator()
	channel := model.ChannelRecord{Number: "A", Name: "Alpha", Secret: "secret", Active: true}
	if err := auth.Authenticate(channel, "secret"); err != nil {
		t.Fatalf("authentication failed: %v", err)
	}
	channel.Active = false
	if err := auth.Authenticate(channel, "secret"); err != ErrInactiveChannel {
		t.Fatalf("expected inactive error, got %v", err)
	}
}

func TestPolicyLimits(t *testing.T) {
	policy := DefaultPolicy()
	batch := model.NewBatch("A", "B", 1)
	message := model.ProtectedMessage{ChannelNumber: "A", BatchNumber: "B", Nonce: "N", Body: "ok"}
	if err := policy.Check(message, batch); err != nil {
		t.Fatalf("policy rejected valid message: %v", err)
	}
	batch.Closed = true
	if err := policy.Check(message, batch); err != ErrBatchClosed {
		t.Fatalf("expected closed batch error, got %v", err)
	}
}
