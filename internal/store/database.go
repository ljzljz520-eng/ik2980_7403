package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"emergencycomms/internal/model"
	"go.etcd.io/bbolt"
)

var (
	channelsBucket = []byte("channels")
	batchesBucket  = []byte("batches")
	auditsBucket   = []byte("audits")
	metaBucket     = []byte("meta")
)

type Database struct {
	path string
	db   *bbolt.DB
}

func Open(path string) (*Database, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	opened, err := bbolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	d := &Database{path: path, db: opened}
	if err := d.ensureBuckets(); err != nil {
		_ = opened.Close()
		return nil, err
	}
	return d, nil
}

func (d *Database) ensureBuckets() error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range [][]byte{channelsBucket, batchesBucket, auditsBucket, metaBucket} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}
		return nil
	})
}

func (d *Database) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.Close()
}

func (d *Database) Path() string { return d.path }

func (d *Database) Health() error {
	if d == nil || d.db == nil {
		return fmt.Errorf("database is not open")
	}
	return d.db.View(func(tx *bbolt.Tx) error {
		for _, name := range [][]byte{channelsBucket, batchesBucket, auditsBucket, metaBucket} {
			if tx.Bucket(name) == nil {
				return fmt.Errorf("missing bucket %s", name)
			}
		}
		return nil
	})
}

func marshal(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal record: %w", err)
	}
	return data, nil
}

func unmarshal(data []byte, target any) error {
	if len(data) == 0 {
		return fmt.Errorf("empty record")
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("unmarshal record: %w", err)
	}
	return nil
}

func keys(bucket *bbolt.Bucket, prefix string) []string {
	values := make([]string, 0)
	if bucket == nil {
		return values
	}
	_ = bucket.ForEach(func(key, _ []byte) error {
		if strings.HasPrefix(string(key), prefix) {
			values = append(values, string(key))
		}
		return nil
	})
	sort.Strings(values)
	return values
}

func recordKey(parts ...string) string { return strings.Join(parts, "|") }

func auditKey(entry model.AuditEntry) string { return recordKey(entry.Channel, entry.Batch, entry.ID) }

func (d *Database) transaction(fn func(*bbolt.Tx) error) error {
	if d == nil || d.db == nil {
		return fmt.Errorf("database is not open")
	}
	return d.db.Update(fn)
}

func (d *Database) view(fn func(*bbolt.Tx) error) error {
	if d == nil || d.db == nil {
		return fmt.Errorf("database is not open")
	}
	return d.db.View(fn)
}

func (d *Database) SetMeta(name, value string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("metadata name is required")
	}
	return d.transaction(func(tx *bbolt.Tx) error {
		return tx.Bucket(metaBucket).Put([]byte(name), []byte(value))
	})
}

func (d *Database) Meta(name string) (string, error) {
	var value string
	err := d.view(func(tx *bbolt.Tx) error {
		data := tx.Bucket(metaBucket).Get([]byte(name))
		if data == nil {
			return fmt.Errorf("metadata %q not found", name)
		}
		value = string(data)
		return nil
	})
	return value, err
}
