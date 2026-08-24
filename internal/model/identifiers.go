package model

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func NormalizeChannel(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func NormalizeBatch(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func NormalizeNonce(value string) string {
	return strings.TrimSpace(value)
}

func AuditID(channel, batch, nonce string) string {
	sum := sha256.Sum256([]byte(channel + "|" + batch + "|" + nonce))
	return hex.EncodeToString(sum[:8])
}

func Fingerprint(message ProtectedMessage) string {
	sum := sha256.Sum256([]byte(message.ChannelNumber + ":" + message.BatchNumber + ":" + message.Nonce + ":" + message.Body))
	return hex.EncodeToString(sum[:])
}

func SortKey(entry AuditEntry) string {
	return entry.Channel + "|" + entry.Batch + "|" + entry.ID
}
