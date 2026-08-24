package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"emergencycomms/internal/model"
)

type AuditFilter struct {
	Outcome string
	Nonce   string
	Reason  string
	From    int
	To      int
}

type AuditDigest struct {
	Channel       string
	Batch         string
	Total         int
	Accepted      int
	Rejected      int
	FirstSequence int
	LastSequence  int
	Reasons       map[string]int
}

func (a *Application) QueryAudits(channel, batch string, filter AuditFilter) ([]model.AuditEntry, error) {
	entries, err := a.db.ListAudits(model.NormalizeChannel(channel), model.NormalizeBatch(batch))
	if err != nil {
		return nil, err
	}
	return FilterAudits(entries, filter), nil
}

func FilterAudits(entries []model.AuditEntry, filter AuditFilter) []model.AuditEntry {
	filtered := make([]model.AuditEntry, 0, len(entries))
	for _, entry := range entries {
		if filter.Outcome != "" && entry.Outcome != filter.Outcome {
			continue
		}
		if filter.Nonce != "" && entry.Nonce != filter.Nonce {
			continue
		}
		if filter.Reason != "" && !strings.Contains(strings.ToLower(entry.Reason), strings.ToLower(filter.Reason)) {
			continue
		}
		if filter.From > 0 && entry.Sequence < filter.From {
			continue
		}
		if filter.To > 0 && entry.Sequence > filter.To {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func MakeAuditDigest(channel, batch string, entries []model.AuditEntry) AuditDigest {
	digest := AuditDigest{Channel: channel, Batch: batch, Total: len(entries), Reasons: make(map[string]int)}
	if len(entries) == 0 {
		return digest
	}
	digest.FirstSequence = entries[0].Sequence
	for _, entry := range entries {
		if entry.Outcome == "accepted" {
			digest.Accepted++
		} else if entry.Outcome == "rejected" {
			digest.Rejected++
		}
		if entry.Sequence < digest.FirstSequence {
			digest.FirstSequence = entry.Sequence
		}
		if entry.Sequence > digest.LastSequence {
			digest.LastSequence = entry.Sequence
		}
		digest.Reasons[entry.Reason]++
	}
	return digest
}

func (d AuditDigest) Completion() string {
	if d.Total == 0 {
		return "empty"
	}
	if d.Rejected == 0 {
		return "clean"
	}
	if d.Accepted == 0 {
		return "blocked"
	}
	return "mixed"
}

func (d AuditDigest) String() string {
	return fmt.Sprintf("%s/%s %s total=%d accepted=%d rejected=%d sequence=%d-%d", d.Channel, d.Batch, d.Completion(), d.Total, d.Accepted, d.Rejected, d.FirstSequence, d.LastSequence)
}

func ParseAuditFilter(values map[string]string) (AuditFilter, error) {
	filter := AuditFilter{Outcome: values["outcome"], Nonce: values["nonce"], Reason: values["reason"]}
	for name, target := range map[string]*int{"from": &filter.From, "to": &filter.To} {
		value := values[name]
		if value == "" {
			continue
		}
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			return AuditFilter{}, fmt.Errorf("invalid %s sequence", name)
		}
		*target = parsed
	}
	if filter.From > 0 && filter.To > 0 && filter.From > filter.To {
		return AuditFilter{}, fmt.Errorf("audit sequence range is reversed")
	}
	return filter, nil
}

func ValidateAuditChain(entries []model.AuditEntry) error {
	if len(entries) == 0 {
		return nil
	}
	ordered := append([]model.AuditEntry(nil), entries...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Sequence < ordered[j].Sequence })
	previous := 0
	for _, entry := range ordered {
		if entry.Sequence < 1 {
			return fmt.Errorf("audit sequence must be positive")
		}
		if entry.Sequence == previous {
			return fmt.Errorf("audit sequence is duplicated")
		}
		if entry.ID == "" || entry.Channel == "" || entry.Batch == "" {
			return fmt.Errorf("audit identity is incomplete")
		}
		previous = entry.Sequence
	}
	return nil
}

func (a *Application) AuditDigest(channel, batch string) (AuditDigest, error) {
	entries, err := a.db.ListAudits(model.NormalizeChannel(channel), model.NormalizeBatch(batch))
	if err != nil {
		return AuditDigest{}, err
	}
	if err := ValidateAuditChain(entries); err != nil {
		return AuditDigest{}, err
	}
	return MakeAuditDigest(model.NormalizeChannel(channel), model.NormalizeBatch(batch), entries), nil
}

func OutcomeExplanation(entry model.AuditEntry) string {
	if entry.Outcome == "accepted" {
		return "authenticated report accepted for operational handling"
	}
	if strings.Contains(entry.Reason, "duplicate") {
		return "replay protection rejected a previously seen nonce"
	}
	if strings.Contains(entry.Reason, "bad-tag") {
		return "authentication tag did not match the channel secret"
	}
	return "report rejected by channel policy"
}

func SeverityForEntry(entry model.AuditEntry) string {
	if entry.Outcome == "accepted" {
		return "info"
	}
	if strings.Contains(entry.Reason, "duplicate") || strings.Contains(entry.Reason, "bad-tag") {
		return "high"
	}
	return "medium"
}

func SortAuditEntries(entries []model.AuditEntry, descending bool) []model.AuditEntry {
	ordered := append([]model.AuditEntry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if descending {
			return ordered[i].Sequence > ordered[j].Sequence
		}
		return ordered[i].Sequence < ordered[j].Sequence
	})
	return ordered
}

func (a *Application) Latest(channel, batch string) (model.AuditEntry, error) {
	return a.db.LatestAudit(model.NormalizeChannel(channel), model.NormalizeBatch(batch))
}
