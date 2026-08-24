package model

import "testing"

func TestIdentifierNormalizationAndFingerprint(t *testing.T) {
	if NormalizeChannel(" alpha-01 ") != "ALPHA-01" {
		t.Fatal("channel normalization failed")
	}
	if NormalizeBatch(" b-1 ") != "B-1" {
		t.Fatal("batch normalization failed")
	}
	message := ProtectedMessage{ChannelNumber: "A", BatchNumber: "B", Nonce: "N", Body: "body"}
	if AuditID("A", "B", "N") == "" || Fingerprint(message) == "" {
		t.Fatal("identifiers should not be empty")
	}
}

func TestDeliveryPlanQuotaAndAlertLifecycle(t *testing.T) {
	plan := DeliveryPlan{Channel: "A", Batch: "B", Operator: "op", MessageCount: 2, Priority: 3, MessageNonces: []string{"N1", "N2"}}
	if err := plan.Validate(); err != nil {
		t.Fatalf("plan validation: %v", err)
	}
	if err := plan.RecordAccepted(); err != nil || plan.Remaining() != 1 {
		t.Fatalf("plan accounting: %v", err)
	}
	quota := ChannelQuota{Channel: "A", Limit: 2}
	if err := quota.Consume(2); err != nil || quota.Available() != 0 {
		t.Fatalf("quota accounting: %v", err)
	}
	alert := AlertSummary{Accepted: 1, Rejected: 1, Duplicate: 1}
	if alert.Severity() != "high" || alert.Total() != 2 {
		t.Fatalf("alert classification failed")
	}
}
