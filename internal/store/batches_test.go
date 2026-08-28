package store

import (
	"path/filepath"
	"testing"

	"emergencycomms/internal/model"
)

func TestBatchAndAuditQueries(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "relay.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	if err := db.SaveBatch(model.NewBatch("A", "B", 2)); err != nil {
		t.Fatalf("save batch: %v", err)
	}
	entry := model.AuditEntry{ID: "1", Channel: "A", Batch: "B", Nonce: "N", Outcome: "accepted", Sequence: 1}
	if err := db.SaveAudit(entry); err != nil {
		t.Fatalf("save audit: %v", err)
	}
	if count, err := db.AuditCount("A", "B"); err != nil || count != 1 {
		t.Fatalf("audit count: %d %v", count, err)
	}
	if err := db.UpdateBatch("A", "B", func(batch *model.BatchRecord) error { batch.Accepted = 1; return nil }); err != nil {
		t.Fatalf("update batch: %v", err)
	}
	batch, err := db.GetBatch("A", "B")
	if err != nil || batch.Accepted != 1 {
		t.Fatalf("updated batch: %v", err)
	}
}

func TestAuditQueriesAndPurge(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "relay.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	for _, entry := range []model.AuditEntry{
		{ID: "1", Channel: "A", Batch: "B", Nonce: "N1", Outcome: "rejected", Reason: "bad-tag", Sequence: 1},
		{ID: "2", Channel: "A", Batch: "B", Nonce: "N2", Outcome: "accepted", Reason: "authenticated", Sequence: 2},
	} {
		if err := db.SaveAudit(entry); err != nil {
			t.Fatalf("save audit: %v", err)
		}
	}
	counts, err := db.CountOutcomes("A", "B")
	if err != nil || counts["rejected"] != 1 {
		t.Fatalf("counts: %v %+v", err, counts)
	}
	match, found, err := db.FindNonce("A", "N1")
	if err != nil || !found || !match.IsRejected() {
		t.Fatalf("find nonce: %v %+v", err, match)
	}
	removed, err := db.PurgeBatchAudits("A", "B")
	if err != nil || removed != 2 {
		t.Fatalf("purge: %d %v", removed, err)
	}
}
