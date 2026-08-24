package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"emergencycomms/internal/model"
)

type AuditDocument struct {
	Channel string             `json:"channel"`
	Batch   string             `json:"batch"`
	Counts  map[string]int     `json:"counts"`
	Entries []model.AuditEntry `json:"entries"`
}

func Document(channel, batch string, entries []model.AuditEntry) AuditDocument {
	counts := make(map[string]int)
	for _, entry := range entries {
		counts[entry.Outcome]++
	}
	return AuditDocument{Channel: channel, Batch: batch, Counts: counts, Entries: append([]model.AuditEntry(nil), entries...)}
}

func JSONDocument(document AuditDocument) (string, error) {
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func WriteCSV(writer io.Writer, entries []model.AuditEntry) error {
	if writer == nil {
		return fmt.Errorf("csv writer is required")
	}
	output := csv.NewWriter(writer)
	if err := output.Write([]string{"sequence", "channel", "batch", "nonce", "outcome", "reason"}); err != nil {
		return err
	}
	ordered := append([]model.AuditEntry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Sequence < ordered[j].Sequence })
	for _, entry := range ordered {
		if err := output.Write([]string{strconv.Itoa(entry.Sequence), entry.Channel, entry.Batch, entry.Nonce, entry.Outcome, entry.Reason}); err != nil {
			return err
		}
	}
	output.Flush()
	return output.Error()
}

func RenderTimeline(entries []model.AuditEntry) string {
	ordered := append([]model.AuditEntry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Sequence < ordered[j].Sequence })
	lines := make([]string, 0, len(ordered))
	for _, entry := range ordered {
		lines = append(lines, fmt.Sprintf("%d %s %s", entry.Sequence, strings.ToUpper(entry.Outcome), entry.Nonce))
	}
	return strings.Join(lines, "\n")
}

func RenderAlert(summary model.AlertSummary) string {
	return fmt.Sprintf("%s accepted=%d rejected=%d duplicate=%d bad-tag=%d unknown=%d", summary.Severity(), summary.Accepted, summary.Rejected, summary.Duplicate, summary.BadTag, summary.Unknown)
}
