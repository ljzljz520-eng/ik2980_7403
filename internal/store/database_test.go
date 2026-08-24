package store

import (
	"path/filepath"
	"testing"

	"emergencycomms/internal/model"
)

func TestDatabaseStoresChannelAndMeta(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "relay.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	channel := model.ChannelRecord{Number: "A", Name: "Alpha", Secret: "s", Active: true}
	if err := db.SaveChannel(channel); err != nil {
		t.Fatalf("save channel: %v", err)
	}
	got, err := db.GetChannel("A")
	if err != nil || got.Name != "Alpha" {
		t.Fatalf("get channel: %v", err)
	}
	if err := db.SetMeta("mode", "test"); err != nil {
		t.Fatalf("set meta: %v", err)
	}
	if value, err := db.Meta("mode"); err != nil || value != "test" {
		t.Fatalf("meta mismatch: %q %v", value, err)
	}
}
