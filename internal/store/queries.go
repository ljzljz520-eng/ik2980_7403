package store

import (
	"fmt"
	"strings"

	"emergencycomms/internal/model"
)

func (d *Database) CountOutcomes(channel, batch string) (map[string]int, error) {
	entries, err := d.ListAudits(channel, batch)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, entry := range entries {
		counts[entry.Outcome]++
	}
	return counts, nil
}

func (d *Database) FindNonce(channel, nonce string) (model.AuditEntry, bool, error) {
	entries, err := d.ListChannelAudits(channel)
	if err != nil {
		return model.AuditEntry{}, false, err
	}
	for _, entry := range entries {
		if entry.Nonce == nonce {
			return entry, true, nil
		}
	}
	return model.AuditEntry{}, false, nil
}

func (d *Database) LatestAudit(channel, batch string) (model.AuditEntry, error) {
	entries, err := d.ListAudits(channel, batch)
	if err != nil {
		return model.AuditEntry{}, err
	}
	if len(entries) == 0 {
		return model.AuditEntry{}, fmt.Errorf("no audits for %s/%s", channel, batch)
	}
	return entries[len(entries)-1], nil
}

func (d *Database) SearchReasons(channel, phrase string) ([]model.AuditEntry, error) {
	entries, err := d.ListChannelAudits(channel)
	if err != nil {
		return nil, err
	}
	matches := make([]model.AuditEntry, 0)
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Reason), strings.ToLower(phrase)) {
			matches = append(matches, entry)
		}
	}
	return matches, nil
}

func (d *Database) PurgeBatchAudits(channel, batch string) (int, error) {
	entries, err := d.ListAudits(channel, batch)
	if err != nil {
		return 0, err
	}
	for _, entry := range entries {
		if err := d.RemoveAudit(entry); err != nil {
			return 0, err
		}
	}
	return len(entries), nil
}

func (d *Database) RejectedEntries(channel string) ([]model.AuditEntry, error) {
	entries, err := d.ListChannelAudits(channel)
	if err != nil {
		return nil, err
	}
	rejected := make([]model.AuditEntry, 0)
	for _, entry := range entries {
		if entry.IsRejected() {
			rejected = append(rejected, entry)
		}
	}
	return rejected, nil
}
