package store

import (
	"fmt"
	"sort"
	"strings"

	"emergencycomms/internal/model"
	"go.etcd.io/bbolt"
)

func (d *Database) SaveAudit(entry model.AuditEntry) error {
	if entry.ID == "" || entry.Channel == "" || entry.Batch == "" {
		return fmt.Errorf("audit identity is required")
	}
	data, err := marshal(entry)
	if err != nil {
		return err
	}
	return d.transaction(func(tx *bbolt.Tx) error {
		return tx.Bucket(auditsBucket).Put([]byte(auditKey(entry)), data)
	})
}

func (d *Database) GetAudit(entryID, channel, batch string) (model.AuditEntry, error) {
	var entry model.AuditEntry
	err := d.view(func(tx *bbolt.Tx) error {
		value := tx.Bucket(auditsBucket).Get([]byte(recordKey(channel, batch, entryID)))
		if value == nil {
			return fmt.Errorf("audit %q not found", entryID)
		}
		return unmarshal(value, &entry)
	})
	return entry, err
}

func (d *Database) ListAudits(channel, batch string) ([]model.AuditEntry, error) {
	entries := make([]model.AuditEntry, 0)
	err := d.view(func(tx *bbolt.Tx) error {
		prefix := recordKey(channel, batch, "")
		return tx.Bucket(auditsBucket).ForEach(func(key, value []byte) error {
			if !strings.HasPrefix(string(key), prefix) {
				return nil
			}
			var entry model.AuditEntry
			if err := unmarshal(value, &entry); err != nil {
				return err
			}
			entries = append(entries, entry)
			return nil
		})
	})
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Sequence == entries[j].Sequence {
			return entries[i].ID < entries[j].ID
		}
		return entries[i].Sequence < entries[j].Sequence
	})
	return entries, err
}

func (d *Database) ListChannelAudits(channel string) ([]model.AuditEntry, error) {
	entries := make([]model.AuditEntry, 0)
	err := d.view(func(tx *bbolt.Tx) error {
		prefix := channel + "|"
		return tx.Bucket(auditsBucket).ForEach(func(key, value []byte) error {
			if !strings.HasPrefix(string(key), prefix) {
				return nil
			}
			var entry model.AuditEntry
			if err := unmarshal(value, &entry); err != nil {
				return err
			}
			entries = append(entries, entry)
			return nil
		})
	})
	sort.Slice(entries, func(i, j int) bool { return entries[i].Sequence < entries[j].Sequence })
	return entries, err
}

func (d *Database) AuditCount(channel, batch string) (int, error) {
	entries, err := d.ListAudits(channel, batch)
	return len(entries), err
}

func (d *Database) RemoveAudit(entry model.AuditEntry) error {
	return d.transaction(func(tx *bbolt.Tx) error {
		return tx.Bucket(auditsBucket).Delete([]byte(auditKey(entry)))
	})
}
