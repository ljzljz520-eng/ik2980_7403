package report

import (
	"fmt"
	"sort"
	"strings"

	"emergencycomms/internal/model"
)

type ChannelMetrics struct {
	Channel        string
	Accepted       int
	Rejected       int
	Duplicate      int
	Authentication int
	LastSequence   int
}

func BuildMetrics(entries []model.AuditEntry) []ChannelMetrics {
	byChannel := make(map[string]ChannelMetrics)
	for _, entry := range entries {
		metrics := byChannel[entry.Channel]
		metrics.Channel = entry.Channel
		if entry.Outcome == "accepted" {
			metrics.Accepted++
		}
		if entry.Outcome == "rejected" {
			metrics.Rejected++
		}
		if strings.Contains(strings.ToLower(entry.Reason), "duplicate") {
			metrics.Duplicate++
		}
		if strings.Contains(strings.ToLower(entry.Reason), "tag") {
			metrics.Authentication++
		}
		if entry.Sequence > metrics.LastSequence {
			metrics.LastSequence = entry.Sequence
		}
		byChannel[entry.Channel] = metrics
	}
	metrics := make([]ChannelMetrics, 0, len(byChannel))
	for _, value := range byChannel {
		metrics = append(metrics, value)
	}
	sort.Slice(metrics, func(i, j int) bool { return metrics[i].Channel < metrics[j].Channel })
	return metrics
}

func (m ChannelMetrics) Total() int { return m.Accepted + m.Rejected }

func (m ChannelMetrics) AcceptanceRate() string {
	if m.Total() == 0 {
		return "0%"
	}
	return fmt.Sprintf("%d%%", m.Accepted*100/m.Total())
}

func RenderMetrics(metrics []ChannelMetrics) string {
	lines := make([]string, 0, len(metrics)+1)
	lines = append(lines, "channel accepted rejected total rate")
	for _, value := range metrics {
		lines = append(lines, fmt.Sprintf("%s %d %d %d %s", value.Channel, value.Accepted, value.Rejected, value.Total(), value.AcceptanceRate()))
	}
	return strings.Join(lines, "\n")
}

func RenderFindingCounts(counts map[string]int) string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, counts[key]))
	}
	return strings.Join(parts, " ")
}

func RenderPlan(plan model.DeliveryPlan) string {
	return fmt.Sprintf("plan %s/%s operator=%s priority=%d status=%s remaining=%d", plan.Channel, plan.Batch, plan.Operator, plan.Priority, plan.Status(), plan.Remaining())
}

func SummarizeEntries(entries []model.AuditEntry) string {
	if len(entries) == 0 {
		return "no audit entries"
	}
	metrics := BuildMetrics(entries)
	return RenderMetrics(metrics)
}

func SortEntriesByReason(entries []model.AuditEntry) []model.AuditEntry {
	ordered := append([]model.AuditEntry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Reason == ordered[j].Reason {
			return ordered[i].Sequence < ordered[j].Sequence
		}
		return ordered[i].Reason < ordered[j].Reason
	})
	return ordered
}

func CountByReason(entries []model.AuditEntry) map[string]int {
	counts := make(map[string]int)
	for _, entry := range entries {
		counts[entry.Reason]++
	}
	return counts
}

func IsHealthy(metrics ChannelMetrics) bool {
	return metrics.Rejected == 0 && metrics.Accepted > 0
}

func RenderHealth(metrics []ChannelMetrics) string {
	lines := make([]string, 0, len(metrics))
	for _, value := range metrics {
		state := "attention"
		if IsHealthy(value) {
			state = "healthy"
		}
		lines = append(lines, fmt.Sprintf("%s %s", value.Channel, state))
	}
	return strings.Join(lines, "\n")
}

func Totals(metrics []ChannelMetrics) ChannelMetrics {
	total := ChannelMetrics{Channel: "ALL"}
	for _, value := range metrics {
		total.Accepted += value.Accepted
		total.Rejected += value.Rejected
		total.Duplicate += value.Duplicate
		total.Authentication += value.Authentication
		if value.LastSequence > total.LastSequence {
			total.LastSequence = value.LastSequence
		}
	}
	return total
}

func RenderTotals(metrics []ChannelMetrics) string {
	total := Totals(metrics)
	return fmt.Sprintf("total accepted=%d rejected=%d rate=%s", total.Accepted, total.Rejected, total.AcceptanceRate())
}

func FilterHealthy(metrics []ChannelMetrics) []ChannelMetrics {
	healthy := make([]ChannelMetrics, 0)
	for _, value := range metrics {
		if IsHealthy(value) {
			healthy = append(healthy, value)
		}
	}
	return healthy
}

func FindMetric(metrics []ChannelMetrics, channel string) (ChannelMetrics, bool) {
	for _, value := range metrics {
		if value.Channel == channel {
			return value, true
		}
	}
	return ChannelMetrics{}, false
}

func SortMetricsByAcceptance(metrics []ChannelMetrics) []ChannelMetrics {
	ordered := append([]ChannelMetrics(nil), metrics...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].AcceptanceRate() == ordered[j].AcceptanceRate() {
			return ordered[i].Channel < ordered[j].Channel
		}
		return ordered[i].Accepted > ordered[j].Accepted
	})
	return ordered
}

func MetricStatus(value ChannelMetrics) string {
	if value.Accepted == 0 && value.Rejected > 0 {
		return "blocked"
	}
	if value.Rejected > value.Accepted {
		return "degraded"
	}
	return "operational"
}

func IsEscalated(value ChannelMetrics) bool {
	return value.Duplicate > 0 || value.Authentication > 0
}

func RenderMetricStatus(value ChannelMetrics) string {
	return fmt.Sprintf("%s %s", value.Channel, MetricStatus(value))
}

func EscalationLabel(value ChannelMetrics) string {
	if IsEscalated(value) {
		return "security-review"
	}
	return "routine"
}

func ShouldPage(value ChannelMetrics) bool {
	return value.Rejected >= 3 || value.Duplicate >= 2
}

func PageReason(value ChannelMetrics) string {
	if value.Duplicate >= 2 {
		return "multiple replay attempts"
	}
	if value.Rejected >= 3 {
		return "rejection threshold exceeded"
	}
	return "no page"
}

func IsQuiet(value ChannelMetrics) bool { return value.Accepted > 0 && value.Rejected == 0 }

func HasRejections(value ChannelMetrics) bool { return value.Rejected > 0 }

func HasAcceptances(value ChannelMetrics) bool { return value.Accepted > 0 }

func NeedsAttention(value ChannelMetrics) bool { return !IsQuiet(value) }
