package store

import (
	"fmt"
	"sort"

	"emergencycomms/internal/model"
	"go.etcd.io/bbolt"
)

func (d *Database) SaveChannel(channel model.ChannelRecord) error {
	if err := channel.Validate(); err != nil {
		return err
	}
	data, err := marshal(channel)
	if err != nil {
		return err
	}
	return d.transaction(func(tx *bbolt.Tx) error {
		return tx.Bucket(channelsBucket).Put([]byte(channel.Number), data)
	})
}

func (d *Database) GetChannel(number string) (model.ChannelRecord, error) {
	var channel model.ChannelRecord
	err := d.view(func(tx *bbolt.Tx) error {
		value := tx.Bucket(channelsBucket).Get([]byte(number))
		if value == nil {
			return fmt.Errorf("channel %q not found", number)
		}
		return unmarshal(value, &channel)
	})
	return channel, err
}

func (d *Database) ListChannels() ([]model.ChannelRecord, error) {
	channels := make([]model.ChannelRecord, 0)
	err := d.view(func(tx *bbolt.Tx) error {
		return tx.Bucket(channelsBucket).ForEach(func(_, value []byte) error {
			var channel model.ChannelRecord
			if err := unmarshal(value, &channel); err != nil {
				return err
			}
			channels = append(channels, channel)
			return nil
		})
	})
	sort.Slice(channels, func(i, j int) bool { return channels[i].Number < channels[j].Number })
	return channels, err
}

func (d *Database) DeleteChannel(number string) error {
	return d.transaction(func(tx *bbolt.Tx) error {
		if tx.Bucket(channelsBucket).Get([]byte(number)) == nil {
			return fmt.Errorf("channel %q not found", number)
		}
		return tx.Bucket(channelsBucket).Delete([]byte(number))
	})
}

func (d *Database) UpdateChannel(number string, update func(*model.ChannelRecord) error) error {
	return d.transaction(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(channelsBucket)
		value := bucket.Get([]byte(number))
		if value == nil {
			return fmt.Errorf("channel %q not found", number)
		}
		var channel model.ChannelRecord
		if err := unmarshal(value, &channel); err != nil {
			return err
		}
		if err := update(&channel); err != nil {
			return err
		}
		if err := channel.Validate(); err != nil {
			return err
		}
		data, err := marshal(channel)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(number), data)
	})
}

func (d *Database) ChannelCount() (int, error) {
	count := 0
	err := d.view(func(tx *bbolt.Tx) error {
		count = tx.Bucket(channelsBucket).Stats().KeyN
		return nil
	})
	return count, err
}
