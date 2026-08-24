package security

import (
	"errors"
	"testing"
)

func TestReplayGuardDetectsAndForgetsNonce(t *testing.T) {
	guard := NewReplayGuard()
	if err := guard.Reserve("A", "N"); err != nil {
		t.Fatalf("first reserve: %v", err)
	}
	err := guard.Reserve("A", "N")
	if !errors.Is(err, ErrDuplicateNonce) {
		t.Fatalf("expected duplicate sentinel: %v", err)
	}
	if !guard.Contains("A", "N") || guard.Count("A") != 1 {
		t.Fatal("nonce should remain reserved")
	}
	if !guard.Forget("A", "N") || guard.Contains("A", "N") {
		t.Fatal("nonce should be forgotten")
	}
}

func TestAdmissionWindowCapacityAndReset(t *testing.T) {
	window := NewAdmissionWindow()
	if err := window.SetLimit("A", 1); err != nil {
		t.Fatalf("set limit: %v", err)
	}
	if err := window.Admit("A"); err != nil {
		t.Fatalf("first admit: %v", err)
	}
	if err := window.Admit("A"); err == nil {
		t.Fatal("second admit should be denied")
	}
	if !window.Release("A") || window.Available("A") != 1 {
		t.Fatal("release should restore capacity")
	}
	window.Reset("A")
	if _, accepted, denied := window.Snapshot("A"); accepted != 0 || denied != 0 {
		t.Fatal("reset should clear counters")
	}
}
