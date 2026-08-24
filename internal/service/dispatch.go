package service

import (
	"fmt"
	"sort"

	"emergencycomms/internal/model"
	"emergencycomms/internal/security"
)

type DispatchReport struct {
	Plan        model.DeliveryPlan
	Accepted    int
	Rejected    int
	Transmitted int
	Failures    []string
}

func (a *Application) BuildPlan(channel, batch, operator string, priority int, bodies []string) (model.DeliveryPlan, error) {
	if operator == "" {
		return model.DeliveryPlan{}, fmt.Errorf("operator is required")
	}
	nonces := make([]string, 0, len(bodies))
	for index := range bodies {
		if bodies[index] == "" {
			return model.DeliveryPlan{}, fmt.Errorf("body %d is empty", index)
		}
		nonces = append(nonces, fmt.Sprintf("%s-%02d", batch, index+1))
	}
	plan := model.DeliveryPlan{Channel: model.NormalizeChannel(channel), Batch: model.NormalizeBatch(batch), Operator: operator, MessageCount: len(bodies), Priority: priority, MessageNonces: nonces}
	if err := plan.Validate(); err != nil {
		return model.DeliveryPlan{}, err
	}
	return plan, nil
}

func (a *Application) Dispatch(plan model.DeliveryPlan, bodies []string, secret string) (DispatchReport, error) {
	if err := plan.Validate(); err != nil {
		return DispatchReport{}, err
	}
	if len(bodies) != len(plan.MessageNonces) {
		return DispatchReport{}, fmt.Errorf("dispatch bodies do not match plan")
	}
	report := DispatchReport{Plan: plan, Failures: make([]string, 0)}
	for index, body := range bodies {
		wire, err := a.Send(plan.Channel, plan.Batch, plan.MessageNonces[index], body, secret)
		if err != nil {
			report.Failures = append(report.Failures, err.Error())
			report.Rejected++
			_ = plan.RecordRejected()
			continue
		}
		report.Transmitted++
		result, receiveErr := a.Receive(wire)
		if receiveErr != nil {
			report.Failures = append(report.Failures, receiveErr.Error())
			report.Rejected++
			_ = plan.RecordRejected()
			continue
		}
		if result.Accepted {
			report.Accepted++
			_ = plan.RecordAccepted()
		} else {
			report.Rejected++
			_ = plan.RecordRejected()
		}
	}
	report.Plan = plan
	return report, nil
}

func (a *Application) Admit(channel string, window *security.AdmissionWindow) error {
	if window == nil {
		return fmt.Errorf("admission window is required")
	}
	return window.Admit(model.NormalizeChannel(channel))
}

func SortReports(reports []DispatchReport) []DispatchReport {
	ordered := append([]DispatchReport(nil), reports...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Accepted == ordered[j].Accepted {
			return ordered[i].Plan.Batch < ordered[j].Plan.Batch
		}
		return ordered[i].Accepted > ordered[j].Accepted
	})
	return ordered
}

func (r DispatchReport) Completion() string {
	if r.Accepted == r.Plan.MessageCount {
		return "complete"
	}
	if r.Accepted == 0 {
		return "failed"
	}
	return "partial"
}

func (r DispatchReport) Labels() []string {
	labels := make([]string, 0, len(r.Failures)+1)
	labels = append(labels, fmt.Sprintf("%s:%d/%d", r.Completion(), r.Accepted, r.Plan.MessageCount))
	for _, failure := range r.Failures {
		labels = append(labels, failure)
	}
	return labels
}
