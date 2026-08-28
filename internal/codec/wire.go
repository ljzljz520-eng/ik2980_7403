package codec

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"emergencycomms/internal/model"
)

type WireMessage struct {
	ChannelNumber string
	BatchNumber   string
	Nonce         string
	Body          string
	Tag           string
	Sequence      int
}

func Encode(message model.ProtectedMessage, secret string) (WireMessage, error) {
	if err := message.Validate(); err != nil {
		return WireMessage{}, err
	}
	if secret == "" {
		return WireMessage{}, fmt.Errorf("encoding secret is required")
	}
	return WireMessage{ChannelNumber: message.ChannelNumber, BatchNumber: message.BatchNumber, Nonce: message.Nonce, Body: message.Body, Sequence: message.Sequence, Tag: TagFor(message.ChannelNumber, message.BatchNumber, message.Nonce, message.Body, secret)}, nil
}

func Decode(wire WireMessage) model.ProtectedMessage {
	return model.ProtectedMessage{ChannelNumber: wire.ChannelNumber, BatchNumber: wire.BatchNumber, Nonce: wire.Nonce, Body: wire.Body, Sequence: wire.Sequence}
}

func TagFor(channel, batch, nonce, body, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.Join([]string{channel, batch, nonce, body}, "|")))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyTag(wire WireMessage, secret string) bool {
	if secret == "" || wire.Tag == "" {
		return false
	}
	expected := TagFor(wire.ChannelNumber, wire.BatchNumber, wire.Nonce, wire.Body, secret)
	return hmac.Equal([]byte(expected), []byte(wire.Tag))
}

func Canonical(wire WireMessage) string {
	return strings.Join([]string{wire.ChannelNumber, wire.BatchNumber, wire.Nonce, wire.Body, wire.Tag}, "|")
}

func ParseFields(fields []string) (WireMessage, error) {
	if len(fields) != 5 {
		return WireMessage{}, fmt.Errorf("wire message requires five fields")
	}
	return WireMessage{ChannelNumber: fields[0], BatchNumber: fields[1], Nonce: fields[2], Body: fields[3], Tag: fields[4]}, nil
}

func SplitCanonical(value string) []string { return strings.Split(value, "|") }
