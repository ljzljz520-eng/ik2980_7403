package report

import (
	"fmt"
	"strings"

	"emergencycomms/internal/model"
)

func RenderReceive(result model.ReceiveResult) string {
	if result.Accepted {
		return fmt.Sprintf("ACCEPTED %s #%d", result.Entry.Channel, result.Entry.Sequence)
	}
	return fmt.Sprintf("REJECTED %s", result.Reason)
}

func RenderAudit(entries []model.AuditEntry) string {
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		lines = append(lines, fmt.Sprintf("%03d %s %s %s", entry.Sequence, entry.Nonce, strings.ToUpper(entry.Outcome), entry.Reason))
	}
	return strings.Join(lines, "\n")
}

func RenderBatch(batch model.BatchRecord) string {
	return fmt.Sprintf("batch %s channel=%s progress=%s accepted=%d rejected=%d", batch.Number, batch.Channel, batch.Progress(), batch.Accepted, batch.Rejected)
}

func RenderChannel(channel model.ChannelRecord) string {
	state := "inactive"
	if channel.Active {
		state = "active"
	}
	return fmt.Sprintf("channel %s %s revision=%d", channel.Number, state, channel.Revision)
}

func RenderSummary(lines []string) string { return strings.Join(lines, "\n") }

func OutcomeLabel(result model.ReceiveResult) string {
	if result.Accepted {
		return "accepted"
	}
	return "rejected"
}
