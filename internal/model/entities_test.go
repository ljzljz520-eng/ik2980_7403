package model

import "testing"

func TestChannelRecordValidation(t *testing.T) {
	valid := ChannelRecord{Number: "A", Name: "Alpha", Secret: "secret", Active: true}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid channel: %v", err)
	}
	valid.Secret = ""
	if err := valid.Validate(); err == nil {
		t.Fatal("expected missing secret error")
	}
}

func TestProtectedMessageValidationAndProgress(t *testing.T) {
	message := ProtectedMessage{ChannelNumber: "A", BatchNumber: "B", Nonce: "N", Body: "ok"}
	if err := message.Validate(); err != nil {
		t.Fatalf("valid message: %v", err)
	}
	batch := NewBatch("A", "B", 2)
	if batch.Progress() != "0/2 accepted" {
		t.Fatalf("unexpected progress: %s", batch.Progress())
	}
	batch.Closed = true
	if batch.Progress() != "closed" {
		t.Fatalf("unexpected closed progress: %s", batch.Progress())
	}
}
