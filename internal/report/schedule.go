package report

import (
	"fmt"
	"sort"
	"strings"

	"emergencycomms/internal/model"
)

type PlanRow struct {
	Channel   string
	Batch     string
	Operator  string
	Priority  int
	Status    string
	Remaining int
}

func BuildPlanRows(plans []model.DeliveryPlan) []PlanRow {
	rows := make([]PlanRow, 0, len(plans))
	for _, plan := range plans {
		rows = append(rows, PlanRow{Channel: plan.Channel, Batch: plan.Batch, Operator: plan.Operator, Priority: plan.Priority, Status: plan.Status(), Remaining: plan.Remaining()})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Priority == rows[j].Priority {
			return rows[i].Batch < rows[j].Batch
		}
		return rows[i].Priority > rows[j].Priority
	})
	return rows
}

func RenderPlanRows(rows []PlanRow) string {
	if len(rows) == 0 {
		return "no delivery plans"
	}
	lines := make([]string, 0, len(rows)+1)
	lines = append(lines, "channel batch operator priority status remaining")
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("%s %s %s %d %s %d", row.Channel, row.Batch, row.Operator, row.Priority, row.Status, row.Remaining))
	}
	return strings.Join(lines, "\n")
}

func FilterPlanRows(rows []PlanRow, status string) []PlanRow {
	if status == "" {
		return append([]PlanRow(nil), rows...)
	}
	filtered := make([]PlanRow, 0)
	for _, row := range rows {
		if row.Status == status {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func PlanRowsByOperator(rows []PlanRow, operator string) []PlanRow {
	filtered := make([]PlanRow, 0)
	for _, row := range rows {
		if row.Operator == operator {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func PlanPriorityLabel(priority int) string {
	switch priority {
	case 5:
		return "critical"
	case 4:
		return "urgent"
	case 3:
		return "normal"
	case 2:
		return "deferred"
	default:
		return "routine"
	}
}

func (row PlanRow) Label() string {
	return fmt.Sprintf("%s %s %s", row.Channel, row.Batch, PlanPriorityLabel(row.Priority))
}

func CountPlanStatuses(rows []PlanRow) map[string]int {
	counts := make(map[string]int)
	for _, row := range rows {
		counts[row.Status]++
	}
	return counts
}

func HasOpenPlans(rows []PlanRow) bool {
	for _, row := range rows {
		if row.Status == "open" {
			return true
		}
	}
	return false
}

func PlanSummary(rows []PlanRow) string {
	counts := CountPlanStatuses(rows)
	return fmt.Sprintf("plans=%d open=%d complete=%d closed=%d", len(rows), counts["open"], counts["complete"], counts["closed"])
}
