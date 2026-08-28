package store

import (
	"fmt"
	"sort"

	"emergencycomms/internal/model"
	"go.etcd.io/bbolt"
)

func (d *Database) SaveBatch(batch model.BatchRecord) error {
	if err := batch.Validate(); err != nil {
		return err
	}
	data, err := marshal(batch)
	if err != nil {
		return err
	}
	return d.transaction(func(tx *bbolt.Tx) error {
		return tx.Bucket(batchesBucket).Put([]byte(recordKey(batch.Channel, batch.Number)), data)
	})
}

func (d *Database) GetBatch(channel, number string) (model.BatchRecord, error) {
	var batch model.BatchRecord
	err := d.view(func(tx *bbolt.Tx) error {
		value := tx.Bucket(batchesBucket).Get([]byte(recordKey(channel, number)))
		if value == nil {
			return fmt.Errorf("batch %q/%q not found", channel, number)
		}
		return unmarshal(value, &batch)
	})
	return batch, err
}

func (d *Database) ListBatches(channel string) ([]model.BatchRecord, error) {
	batches := make([]model.BatchRecord, 0)
	err := d.view(func(tx *bbolt.Tx) error {
		prefix := channel + "|"
		return tx.Bucket(batchesBucket).ForEach(func(key, value []byte) error {
			if len(key) < len(prefix) || string(key[:len(prefix)]) != prefix {
				return nil
			}
			var batch model.BatchRecord
			if err := unmarshal(value, &batch); err != nil {
				return err
			}
			batches = append(batches, batch)
			return nil
		})
	})
	sort.Slice(batches, func(i, j int) bool { return batches[i].Number < batches[j].Number })
	return batches, err
}

func (d *Database) UpdateBatch(channel, number string, update func(*model.BatchRecord) error) error {
	return d.transaction(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(batchesBucket)
		key := []byte(recordKey(channel, number))
		value := bucket.Get(key)
		if value == nil {
			return fmt.Errorf("batch %q/%q not found", channel, number)
		}
		var batch model.BatchRecord
		if err := unmarshal(value, &batch); err != nil {
			return err
		}
		if err := update(&batch); err != nil {
			return err
		}
		if err := batch.Validate(); err != nil {
			return err
		}
		data, err := marshal(batch)
		if err != nil {
			return err
		}
		return bucket.Put(key, data)
	})
}

func (d *Database) CloseBatch(channel, number string) error {
	return d.UpdateBatch(channel, number, func(batch *model.BatchRecord) error {
		batch.Closed = true
		return nil
	})
}

func (d *Database) BatchCount(channel string) (int, error) {
	batches, err := d.ListBatches(channel)
	return len(batches), err
}
