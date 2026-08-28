package codec

import (
	"testing"

	"emergencycomms/internal/model"
)

func TestEncodeAndVerifyWire(t *testing.T) {
	message := model.ProtectedMessage{ChannelNumber: "A", BatchNumber: "B", Nonce: "N", Body: "report"}
	wire, err := Encode(message, "secret")
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !VerifyTag(wire, "secret") {
		t.Fatal("tag should verify")
	}
	wire.Body = "tampered"
	if VerifyTag(wire, "secret") {
		t.Fatal("tampered tag should fail")
	}
}

func TestCanonicalRoundTrip(t *testing.T) {
	wire := WireMessage{ChannelNumber: "A", BatchNumber: "B", Nonce: "N", Body: "body", Tag: "tag"}
	decoded, err := ParseFields(SplitCanonical(Canonical(wire)))
	if err != nil || decoded.Body != wire.Body {
		t.Fatalf("round trip failed: %v", err)
	}
}
