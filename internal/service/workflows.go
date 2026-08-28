package service

import (
	"fmt"

	"emergencycomms/internal/codec"
	"emergencycomms/internal/model"
)

func (a *Application) PrepareChannel(channel, name, secret string) error {
	return a.RegisterChannel(model.ChannelRecord{Number: channel, Name: name, Secret: secret, Active: true, Revision: 1})
}

func (a *Application) PrepareBatch(channel, number string, expected int) error {
	return a.RegisterBatch(model.NewBatch(channel, number, expected))
}

func (a *Application) Send(channel, batch, nonce, body, secret string) (codec.WireMessage, error) {
	message := model.ProtectedMessage{ChannelNumber: model.NormalizeChannel(channel), BatchNumber: model.NormalizeBatch(batch), Nonce: nonce, Body: body}
	wire, err := codec.Encode(message, secret)
	if err != nil {
		return codec.WireMessage{}, err
	}
	if err := a.AcceptOutgoing(wire); err != nil {
		return codec.WireMessage{}, err
	}
	return wire, nil
}

func (a *Application) AcceptOutgoing(wire codec.WireMessage) error {
	message := codec.Decode(wire)
	if err := message.Validate(); err != nil {
		return err
	}
	if _, err := a.db.GetChannel(message.ChannelNumber); err != nil {
		return err
	}
	if _, err := a.db.GetBatch(message.ChannelNumber, message.BatchNumber); err != nil {
		return err
	}
	return nil
}

func (a *Application) Demo(channel, batch, nonce, body string) (string, error) {
	if err := a.PrepareChannel(channel, "Demo relay", "ALPHA-SECRET"); err != nil {
		return "", err
	}
	if err := a.PrepareBatch(channel, batch, 3); err != nil {
		return "", err
	}
	wire, err := a.Send(channel, batch, nonce, body, "ALPHA-SECRET")
	if err != nil {
		return "", err
	}
	result, err := a.Receive(wire)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s: %s", result.Outcome, result.Reason), nil
}

func (a *Application) CloseBatch(channel, batch string) error {
	return a.db.CloseBatch(model.NormalizeChannel(channel), model.NormalizeBatch(batch))
}

func (a *Application) RotateChannelSecret(channel, next string) error {
	return a.db.UpdateChannel(model.NormalizeChannel(channel), func(record *model.ChannelRecord) error {
		return a.rotate(record, next)
	})
}

func (a *Application) rotate(record *model.ChannelRecord, next string) error {
	if record.Secret == next {
		return fmt.Errorf("new channel secret must differ")
	}
	record.Secret = next
	record.Revision++
	return nil
}

func (a *Application) DeactivateChannel(channel string) error {
	return a.db.UpdateChannel(model.NormalizeChannel(channel), func(record *model.ChannelRecord) error {
		record.Active = false
		return nil
	})
}

func (a *Application) ReactivateChannel(channel string) error {
	return a.db.UpdateChannel(model.NormalizeChannel(channel), func(record *model.ChannelRecord) error {
		record.Active = true
		return nil
	})
}
