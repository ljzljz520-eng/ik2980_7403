package model

import (
	"fmt"
	"sort"
	"strings"
)

type DeliveryPlan struct {
	Channel       string
	Batch         string
	Operator      string
	MessageCount  int
	Priority      int
	Acknowledged  int
	Rejected      int
	Closed        bool
	MessageNonces []string
}

type Segment struct {
	MessageID string
	Index     int
	Total     int
	Payload   string
	Checksum  string
}

type ChannelQuota struct {
	Channel string
	Limit   int
	Used    int
}

type AlertSummary struct {
	Channel      string
	Accepted     int
	Rejected     int
	Duplicate    int
	BadTag       int
	Unknown      int
	LastSequence int
}

func (p DeliveryPlan) Validate() error {
	if strings.TrimSpace(p.Channel) == "" || strings.TrimSpace(p.Batch) == "" {
		return fmt.Errorf("delivery plan channel and batch are required")
	}
	if p.Operator == "" {
		return fmt.Errorf("delivery plan operator is required")
	}
	if p.MessageCount < 1 {
		return fmt.Errorf("delivery plan requires at least one message")
	}
	if p.Priority < 1 || p.Priority > 5 {
		return fmt.Errorf("delivery priority must be between one and five")
	}
	if len(p.MessageNonces) != p.MessageCount {
		return fmt.Errorf("delivery plan nonce count does not match message count")
	}
	return nil
}

func (p DeliveryPlan) Remaining() int {
	remaining := p.MessageCount - p.Acknowledged - p.Rejected
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (p *DeliveryPlan) RecordAccepted() error {
	if p.Closed {
		return fmt.Errorf("delivery plan is closed")
	}
	if p.Remaining() == 0 {
		return fmt.Errorf("delivery plan has no remaining messages")
	}
	p.Acknowledged++
	return nil
}

func (p *DeliveryPlan) RecordRejected() error {
	if p.Closed {
		return fmt.Errorf("delivery plan is closed")
	}
	if p.Remaining() == 0 {
		return fmt.Errorf("delivery plan has no remaining messages")
	}
	p.Rejected++
	return nil
}

func (p *DeliveryPlan) Close() error {
	if p.Closed {
		return fmt.Errorf("delivery plan already closed")
	}
	p.Closed = true
	return nil
}

func (p DeliveryPlan) Status() string {
	if p.Closed {
		return "closed"
	}
	if p.Remaining() == 0 {
		return "complete"
	}
	return "open"
}

func SortPlans(plans []DeliveryPlan) []DeliveryPlan {
	ordered := append([]DeliveryPlan(nil), plans...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Priority == ordered[j].Priority {
			return ordered[i].Batch < ordered[j].Batch
		}
		return ordered[i].Priority > ordered[j].Priority

	})
	return ordered
}

func (q ChannelQuota) Available() int {
	available := q.Limit - q.Used
	if available < 0 {
		return 0
	}
	return available
}

func (q *ChannelQuota) Consume(amount int) error {
	if amount < 1 {
		return fmt.Errorf("quota consumption must be positive")
	}
	if amount > q.Available() {
		return fmt.Errorf("channel quota exceeded")
	}
	q.Used += amount
	return nil
}

func (q *ChannelQuota) Restore(amount int) error {
	if amount < 1 || amount > q.Used {
		return fmt.Errorf("invalid quota restore")
	}
	q.Used -= amount
	return nil
}

func (s AlertSummary) Severity() string {
	if s.Duplicate > 0 || s.BadTag > 0 {
		return "high"
	}
	if s.Unknown > 0 || s.Rejected > 0 {
		return "medium"
	}
	return "normal"
}

func (s AlertSummary) Total() int { return s.Accepted + s.Rejected }
