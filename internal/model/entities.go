package model

import "fmt"

type ChannelRecord struct {
	Number   string `json:"number"`
	Name     string `json:"name"`
	Secret   string `json:"secret"`
	Active   bool   `json:"active"`
	Revision int    `json:"revision"`
}

type BatchRecord struct {
	Number       string `json:"number"`
	Channel      string `json:"channel"`
	Expected     int    `json:"expected"`
	Accepted     int    `json:"accepted"`
	Rejected     int    `json:"rejected"`
	Closed       bool   `json:"closed"`
	LastSequence int    `json:"last_sequence"`
}

type ProtectedMessage struct {
	ChannelNumber string `json:"channel_number"`
	BatchNumber   string `json:"batch_number"`
	Nonce         string `json:"nonce"`
	Body          string `json:"body"`
	Sequence      int    `json:"sequence"`
}

type AuditEntry struct {
	ID          string `json:"id"`
	Channel     string `json:"channel"`
	Batch       string `json:"batch"`
	Nonce       string `json:"nonce"`
	Outcome     string `json:"outcome"`
	Reason      string `json:"reason"`
	Sequence    int    `json:"sequence"`
	Fingerprint string `json:"fingerprint"`
}

type ReceiveResult struct {
	Accepted bool
	Outcome  string
	Reason   string
	Entry    AuditEntry
}

func (c ChannelRecord) Validate() error {
	if c.Number == "" {
		return fmt.Errorf("channel number is required")
	}
	if c.Name == "" {
		return fmt.Errorf("channel name is required")
	}
	if c.Secret == "" {
		return fmt.Errorf("channel secret is required")
	}
	return nil
}

func (b BatchRecord) Validate() error {
	if b.Number == "" || b.Channel == "" {
		return fmt.Errorf("batch number and channel are required")
	}
	if b.Expected < 0 {
		return fmt.Errorf("expected count cannot be negative")
	}
	if b.Accepted < 0 || b.Rejected < 0 {
		return fmt.Errorf("batch counters cannot be negative")
	}
	return nil
}

func (m ProtectedMessage) Validate() error {
	if m.ChannelNumber == "" || m.BatchNumber == "" {
		return fmt.Errorf("message channel and batch are required")
	}
	if m.Nonce == "" {
		return fmt.Errorf("message nonce is required")
	}
	if m.Body == "" {
		return fmt.Errorf("message body is required")
	}
	if len(m.Body) > 280 {
		return fmt.Errorf("message body exceeds short report limit")
	}
	return nil
}

func (a AuditEntry) IsRejected() bool { return a.Outcome == "rejected" }

func NewBatch(channel, number string, expected int) BatchRecord {
	return BatchRecord{Number: number, Channel: channel, Expected: expected}
}

func (b BatchRecord) Progress() string {
	if b.Closed {
		return "closed"
	}
	if b.Expected == 0 {
		return "open"
	}
	return fmt.Sprintf("%d/%d accepted", b.Accepted, b.Expected)
}
