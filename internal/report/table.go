package report

import (
	"fmt"
	"strings"

	"emergencycomms/internal/service"
)

func RenderSummaries(summaries []service.ChannelSummary) string {
	lines := make([]string, 0, len(summaries))
	for _, summary := range summaries {
		lines = append(lines, service.FormatSummary(summary))
	}
	return strings.Join(lines, "\n")
}

func RenderHeader(title string) string {
	return fmt.Sprintf("== %s ==", strings.TrimSpace(title))
}

func RenderKeyValue(key, value string) string {
	return fmt.Sprintf("%-16s %s", key, value)
}

func RenderCounts(counts map[string]int) string {
	return fmt.Sprintf("accepted=%d rejected=%d", counts["accepted"], counts["rejected"])
}
