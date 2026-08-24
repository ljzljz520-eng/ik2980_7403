package service

import (
	"fmt"
	"sort"
	"strings"

	"emergencycomms/internal/codec"
	"emergencycomms/internal/model"
	"emergencycomms/internal/security"
)

type WireFinding struct {
	Code     string
	Severity string
	Detail   string
}

type WireInspection struct {
	Message  model.ProtectedMessage
	Findings []WireFinding
	Ready    bool
}

func (a *Application) InspectWire(wire codec.WireMessage) WireInspection {
	message := codec.Decode(wire)
	inspection := WireInspection{Message: message, Findings: make([]WireFinding, 0), Ready: true}
	if !codec.IsWellFormed(wire) {
		inspection.add("shape", "high", "wire fields are incomplete")
	}
	if err := message.Validate(); err != nil {
		inspection.add("message", "high", err.Error())
	}
	channel, err := a.db.GetChannel(model.NormalizeChannel(wire.ChannelNumber))
	if err != nil {
		inspection.add("channel", "high", security.ErrUnknownChannel.Error())
	} else {
		if !channel.Active {
			inspection.add("inactive", "high", security.ErrInactiveChannel.Error())
		}
		if !codec.VerifyTag(wire, channel.Secret) {
			inspection.add("tag", "high", security.ErrBadTag.Error())
		}
	}
	if len(inspection.Findings) > 0 {
		inspection.Ready = false
	}
	return inspection
}

func (i *WireInspection) add(code, severity, detail string) {
	i.Findings = append(i.Findings, WireFinding{Code: code, Severity: severity, Detail: detail})
}

func (i WireInspection) HighestSeverity() string {
	for _, severity := range []string{"high", "medium", "low"} {
		for _, finding := range i.Findings {
			if finding.Severity == severity {
				return severity
			}
		}
	}
	return "none"
}

func (i WireInspection) Codes() []string {
	codes := make([]string, 0, len(i.Findings))
	for _, finding := range i.Findings {
		codes = append(codes, finding.Code)
	}
	sort.Strings(codes)
	return codes
}

func (i WireInspection) Summary() string {
	if i.Ready {
		return "ready"
	}
	return fmt.Sprintf("blocked:%s:%s", i.HighestSeverity(), strings.Join(i.Codes(), ","))
}

func ClassifyReason(reason string) string {
	lower := strings.ToLower(reason)
	switch {
	case strings.Contains(lower, "duplicate"):
		return "replay"
	case strings.Contains(lower, "tag"):
		return "integrity"
	case strings.Contains(lower, "channel"):
		return "routing"
	case strings.Contains(lower, "batch"):
		return "lifecycle"
	default:
		return "policy"
	}
}

func (a *Application) ReviewBatch(channel, batch string) ([]WireFinding, error) {
	entries, err := a.db.ListAudits(model.NormalizeChannel(channel), model.NormalizeBatch(batch))
	if err != nil {
		return nil, err
	}
	findings := make([]WireFinding, 0)
	for _, entry := range entries {
		if entry.IsRejected() {
			findings = append(findings, WireFinding{Code: ClassifyReason(entry.Reason), Severity: "medium", Detail: entry.Reason})
		}
	}
	return findings, nil
}

func GroupFindings(findings []WireFinding) map[string]int {
	counts := make(map[string]int)
	for _, finding := range findings {
		counts[finding.Code]++
	}
	return counts
}

func SortFindings(findings []WireFinding) []WireFinding {
	ordered := append([]WireFinding(nil), findings...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Severity == ordered[j].Severity {
			return ordered[i].Code < ordered[j].Code
		}
		return ordered[i].Severity < ordered[j].Severity
	})
	return ordered
}

func SafetyLabel(inspection WireInspection) string {
	if inspection.Ready {
		return "PASS"
	}
	return "BLOCK " + inspection.HighestSeverity()
}

func RequiresOperatorReview(findings []WireFinding) bool {
	for _, finding := range findings {
		if finding.Severity == "high" {
			return true
		}
	}
	return false
}
