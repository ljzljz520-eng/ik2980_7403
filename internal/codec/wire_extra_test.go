package codec

import (
	"testing"

	"emergencycomms/internal/model"
)

func TestTransportRoundTripAndShape(t *testing.T) {
	wire := WireMessage{ChannelNumber: "A", BatchNumber: "B", Nonce: "N", Body: "body", Tag: "tag"}
	encoded := EncodeTransport(wire)
	decoded, err := DecodeTransport(encoded)
	if err != nil || !IsWellFormed(decoded) {
		t.Fatalf("transport round trip failed: %v", err)
	}
	if _, err := DecodeTransport("invalid"); err == nil {
		t.Fatal("invalid transport should fail")
	}
}

func TestFragmentAndAssembleFrames(t *testing.T) {
	message := model.ProtectedMessage{ChannelNumber: "A", BatchNumber: "B", Nonce: "N", Body: "abcdefgh"}
	segments, err := Fragment(message, 3)
	if err != nil || len(segments) != 3 {
		t.Fatalf("fragment: %v %d", err, len(segments))
	}
	encoded := EncodeFrame(segments[1])
	decoded, err := DecodeFrame(encoded)
	if err != nil || decoded.Payload != "def" {
		t.Fatalf("frame round trip: %v %+v", err, decoded)
	}
	assembled, err := Assemble(segments)
	if err != nil || assembled != message.Body {
		t.Fatalf("assemble: %v %s", err, assembled)
	}
}
