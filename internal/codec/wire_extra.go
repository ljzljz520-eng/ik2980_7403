package codec

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func EncodeTransport(wire WireMessage) string {
	return base64.RawStdEncoding.EncodeToString([]byte(Canonical(wire)))
}

func DecodeTransport(value string) (WireMessage, error) {
	if strings.TrimSpace(value) == "" {
		return WireMessage{}, fmt.Errorf("empty transport payload")
	}
	decoded, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil {
		return WireMessage{}, fmt.Errorf("transport decode: %w", err)
	}
	return ParseFields(SplitCanonical(string(decoded)))
}

func IsWellFormed(wire WireMessage) bool {
	if wire.ChannelNumber == "" || wire.BatchNumber == "" || wire.Nonce == "" {
		return false
	}
	return wire.Body != "" && wire.Tag != ""
}

func WithSequence(wire WireMessage, sequence int) WireMessage {
	wire.Sequence = sequence
	return wire
}

func Describe(wire WireMessage) string {
	return fmt.Sprintf("%s/%s nonce=%s sequence=%d", wire.ChannelNumber, wire.BatchNumber, wire.Nonce, wire.Sequence)
}
