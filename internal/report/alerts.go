package report

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"emergencycomms/internal/model"
)

type AlertRow struct {
	Channel  string
	Batch    string
	Sequence int
	Nonce    string
	Severity string
	Message  string
}

func BuildAlerts(entries []model.AuditEntry) []AlertRow {
	alerts := make([]AlertRow, 0)
	for _, entry := range entries {
		if entry.Outcome == "accepted" {
			continue
		}
		severity := "medium"
		message := "report rejected by policy"
		lower := strings.ToLower(entry.Reason)
		if strings.Contains(lower, "duplicate") || strings.Contains(lower, "bad-tag") {
			severity = "high"
		}
		if strings.Contains(lower, "unknown-channel") {
			message = "report addressed to an unregistered channel"
		} else if strings.Contains(lower, "duplicate") {
			message = "repeated nonce requires operator review"
		} else if strings.Contains(lower, "bad-tag") {
			message = "authentication tag failed verification"
		}
		alerts = append(alerts, AlertRow{Channel: entry.Channel, Batch: entry.Batch, Sequence: entry.Sequence, Nonce: entry.Nonce, Severity: severity, Message: message})
	}
	sort.SliceStable(alerts, func(i, j int) bool {
		if alerts[i].Severity == alerts[j].Severity {
			return alerts[i].Sequence < alerts[j].Sequence
		}
		return alerts[i].Severity == "high"
	})
	return alerts
}

func RenderAlerts(alerts []AlertRow) string {
	if len(alerts) == 0 {
		return "no alerts"
	}
	lines := make([]string, 0, len(alerts))
	for _, alert := range alerts {
		lines = append(lines, fmt.Sprintf("[%s] %s/%s #%d %s", strings.ToUpper(alert.Severity), alert.Channel, alert.Batch, alert.Sequence, alert.Message))
	}
	return strings.Join(lines, "\n")
}

func AlertCountBySeverity(alerts []AlertRow) map[string]int {
	counts := make(map[string]int)
	for _, alert := range alerts {
		counts[alert.Severity]++
	}
	return counts
}

func HasCriticalAlert(alerts []AlertRow) bool {
	for _, alert := range alerts {
		if alert.Severity == "high" {
			return true
		}
	}
	return false
}

func RenderOperatorBrief(channel string, alerts []AlertRow) string {
	counts := AlertCountBySeverity(alerts)
	state := "clear"
	if HasCriticalAlert(alerts) {
		state = "review-required"
	} else if len(alerts) > 0 {
		state = "monitor"
	}
	return fmt.Sprintf("%s %s high=%d medium=%d", channel, state, counts["high"], counts["medium"])
}

func FilterAlerts(alerts []AlertRow, severity string) []AlertRow {
	if severity == "" {
		return append([]AlertRow(nil), alerts...)
	}
	filtered := make([]AlertRow, 0)
	for _, alert := range alerts {
		if alert.Severity == severity {
			filtered = append(filtered, alert)
		}
	}
	return filtered
}

func EncodeAlerts(alerts []AlertRow) (string, error) {
	data, err := json.Marshal(alerts)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func AlertSummaryLine(alerts []AlertRow) string {
	counts := AlertCountBySeverity(alerts)
	return fmt.Sprintf("alerts=%d high=%d medium=%d", len(alerts), counts["high"], counts["medium"])
}

func EscalationReason(alert AlertRow) string {
	if alert.Severity == "high" {
		return "escalate to communications security officer"
	}
	return "retain for shift review"
}

func SortAlertsByNonce(alerts []AlertRow) []AlertRow {
	ordered := append([]AlertRow(nil), alerts...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Nonce == ordered[j].Nonce {
			return ordered[i].Sequence < ordered[j].Sequence
		}
		return ordered[i].Nonce < ordered[j].Nonce
	})
	return ordered
}
