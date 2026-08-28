package security

import (
	"fmt"
	"strings"

	"emergencycomms/internal/model"
)

type Policy struct {
	MaxBodyLength int
	RequireBatch  bool
}

func DefaultPolicy() Policy { return Policy{MaxBodyLength: 280, RequireBatch: true} }

func (p Policy) Check(message model.ProtectedMessage, batch model.BatchRecord) error {
	if p.RequireBatch && batch.Number == "" {
		return fmt.Errorf("batch is required")
	}
	if message.Body == "" || strings.TrimSpace(message.Body) == "" {
		return fmt.Errorf("empty report body")
	}
	if p.MaxBodyLength > 0 && len(message.Body) > p.MaxBodyLength {
		return fmt.Errorf("report exceeds policy limit")
	}
	if batch.Closed {
		return ErrBatchClosed
	}
	if batch.Expected > 0 && batch.Accepted >= batch.Expected {
		return fmt.Errorf("batch capacity reached")
	}
	return nil
}

func (p Policy) Summary() string {
	return fmt.Sprintf("body<=%d batch-required=%t", p.MaxBodyLength, p.RequireBatch)
}

func (p Policy) AcceptsLength(length int) bool {
	return length > 0 && (p.MaxBodyLength == 0 || length <= p.MaxBodyLength)
}

func (p Policy) WithLimit(limit int) Policy {
	if limit <= 0 {
		return p
	}
	p.MaxBodyLength = limit
	return p
}
