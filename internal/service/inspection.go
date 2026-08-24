package service

import (
	"fmt"
	"sort"
	"strings"

	"emergencycomms/internal/model"
)

type ChannelSummary struct {
	Channel string
	Status  string
	Batches int
	Audits  int
}

func (a *Application) Summaries() ([]ChannelSummary, error) {
	channels, err := a.db.ListChannels()
	if err != nil {
		return nil, err
	}
	summaries := make([]ChannelSummary, 0, len(channels))
	for _, channel := range channels {
		batches, batchErr := a.db.ListBatches(channel.Number)
		if batchErr != nil {
			return nil, batchErr
		}
		audits, auditErr := a.db.ListChannelAudits(channel.Number)
		if auditErr != nil {
			return nil, auditErr
		}
		summaries = append(summaries, ChannelSummary{Channel: channel.Number, Status: channelStatus(channel), Batches: len(batches), Audits: len(audits)})
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].Channel < summaries[j].Channel })
	return summaries, nil
}

func channelStatus(channel model.ChannelRecord) string {
	if !channel.Active {
		return "inactive"
	}
	return "active"
}

func FormatSummary(summary ChannelSummary) string {
	return fmt.Sprintf("%s %s batches=%d audits=%d", summary.Channel, summary.Status, summary.Batches, summary.Audits)
}

func FilterSummaries(summaries []ChannelSummary, status string) []ChannelSummary {
	filtered := make([]ChannelSummary, 0)
	for _, summary := range summaries {
		if status == "" || strings.EqualFold(summary.Status, status) {
			filtered = append(filtered, summary)
		}
	}
	return filtered
}

func AuditCounts(entries []model.AuditEntry) map[string]int {
	counts := make(map[string]int)
	for _, entry := range entries {
		counts[entry.Outcome]++
	}
	return counts
}

func (a *Application) ExplainBatch(channel, batch string) (string, error) {
	record, err := a.Batch(channel, batch)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s", record.Number, record.Progress()), nil
}
