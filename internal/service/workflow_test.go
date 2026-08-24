package service

import (
	"errors"
	"path/filepath"
	"testing"

	"emergencycomms/internal/codec"
	"emergencycomms/internal/security"
	"emergencycomms/internal/store"
)

func testApplication(t *testing.T) (*Application, func()) {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "relay.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return New(db), func() { _ = db.Close() }
}

func seedWorkflow(t *testing.T) (*Application, codec.WireMessage, func()) {
	app, closeApp := testApplication(t)
	if err := app.PrepareChannel("ALPHA-01", "Field Alpha", "secret"); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if err := app.PrepareBatch("ALPHA-01", "BATCH-1", 4); err != nil {
		t.Fatalf("batch: %v", err)
	}
	wire, err := app.Send("ALPHA-01", "BATCH-1", "N-1", "status-green", "secret")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	return app, wire, closeApp
}

func TestWorkflowRegisterSendReceiveAndAudit(t *testing.T) {
	app, wire, closeApp := seedWorkflow(t)
	defer closeApp()
	result, err := app.Receive(wire)
	if err != nil || !result.Accepted {
		t.Fatalf("receive: %v %+v", err, result)
	}
	entries, err := app.Audit("alpha-01")
	if err != nil || len(entries) != 1 || entries[0].Outcome != "accepted" {
		t.Fatalf("audit: %v %+v", err, entries)
	}
}

func TestWorkflowRejectsCorruptTagAndWrongChannel(t *testing.T) {
	app, wire, closeApp := seedWorkflow(t)
	defer closeApp()
	wire.Tag = "bad"
	result, err := app.Receive(wire)
	if err != nil || result.Accepted || result.Outcome != "rejected" {
		t.Fatalf("bad tag result: %v %+v", err, result)
	}
	wire.ChannelNumber = "UNKNOWN"
	result, err = app.Receive(wire)
	if err != nil || result.Accepted || !errors.Is(security.ErrUnknownChannel, security.ErrUnknownChannel) {
		t.Fatalf("unknown channel result: %v %+v", err, result)
	}
}

func TestWorkflowCloseBatchAndInspectSummary(t *testing.T) {
	app, wire, closeApp := seedWorkflow(t)
	defer closeApp()
	if _, err := app.Receive(wire); err != nil {
		t.Fatalf("receive: %v", err)
	}
	if err := app.CloseBatch("ALPHA-01", "BATCH-1"); err != nil {
		t.Fatalf("close batch: %v", err)
	}
	if status, err := app.ChannelStatus("ALPHA-01"); err != nil || status == "" {
		t.Fatalf("status: %s %v", status, err)
	}
	summaries, err := app.Summaries()
	if err != nil || len(summaries) != 1 || summaries[0].Batches != 1 {
		t.Fatalf("summaries: %v %+v", err, summaries)
	}
}

func TestWorkflowDispatchesDeterministicPlan(t *testing.T) {
	app, _, closeApp := seedWorkflow(t)
	defer closeApp()
	plan, err := app.BuildPlan("ALPHA-01", "BATCH-1", "operator", 4, []string{"one", "two"})
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	result, err := app.Dispatch(plan, []string{"one", "two"}, "secret")
	if err != nil || result.Accepted != 2 || result.Completion() != "complete" {
		t.Fatalf("dispatch: %v %+v", err, result)
	}
}

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "relay.db")
	db, err := store.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	app := New(db)
	if err := app.PrepareChannel("BRAVO-02", "Field Bravo", "bravo-secret"); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if err := app.PrepareBatch("BRAVO-02", "BATCH-9", 2); err != nil {
		t.Fatalf("batch: %v", err)
	}
	wire, err := app.Send("BRAVO-02", "BATCH-9", "N-9", "status-yellow", "bravo-secret")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if result, err := app.Receive(wire); err != nil || !result.Accepted {
		t.Fatalf("receive: %v %+v", err, result)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	reopened, err := store.Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	if _, err := reopened.GetChannel("BRAVO-02"); err != nil {
		t.Fatalf("channel missing after reopen: %v", err)
	}
	batch, err := reopened.GetBatch("BRAVO-02", "BATCH-9")
	if err != nil || batch.Accepted != 1 {
		t.Fatalf("batch missing after reopen: %v %+v", err, batch)
	}
	entries, err := reopened.ListAudits("BRAVO-02", "BATCH-9")
	if err != nil || len(entries) != 1 {
		t.Fatalf("audit missing after reopen: %v %+v", err, entries)
	}
}

func TestRepeatedNonceMessageIsRejected(t *testing.T) {
	app, wire, closeApp := seedWorkflow(t)
	defer closeApp()
	first, err := app.Receive(wire)
	if err != nil || !first.Accepted {
		t.Fatalf("first receive: %v %+v", err, first)
	}
	second, err := app.Receive(wire)
	if err != nil {
		t.Fatalf("second receive returned transport error: %v", err)
	}
	if second.Accepted {
		t.Fatalf("repeated nonce was accepted: %+v", second)
	}
	if second.Reason == "" {
		t.Fatal("duplicate rejection should include a reason")
	}
}
