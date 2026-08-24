package service

import (
	"errors"
	"fmt"
	"strings"

	"emergencycomms/internal/codec"
	"emergencycomms/internal/model"
	"emergencycomms/internal/security"
	"emergencycomms/internal/store"
)

type Application struct {
	db       *store.Database
	guard    *security.ReplayGuard
	auth     *security.ChannelAuthenticator
	policy   security.Policy
	sequence int
}

func New(db *store.Database) *Application {
	return &Application{db: db, guard: security.NewReplayGuard(), auth: security.NewChannelAuthenticator(), policy: security.DefaultPolicy()}
}

func (a *Application) RegisterChannel(channel model.ChannelRecord) error {
	channel.Number = model.NormalizeChannel(channel.Number)
	if channel.Revision == 0 {
		channel.Revision = 1
	}
	if err := channel.Validate(); err != nil {
		return err
	}
	return a.db.SaveChannel(channel)
}

func (a *Application) RegisterBatch(batch model.BatchRecord) error {
	batch.Channel = model.NormalizeChannel(batch.Channel)
	batch.Number = model.NormalizeBatch(batch.Number)
	if _, err := a.db.GetChannel(batch.Channel); err != nil {
		return fmt.Errorf("register batch channel: %w", err)
	}
	return a.db.SaveBatch(batch)
}

func (a *Application) Authenticate(channel, secret string) error {
	record, err := a.db.GetChannel(model.NormalizeChannel(channel))
	if err != nil {
		return fmt.Errorf("%w: %s", security.ErrUnknownChannel, channel)
	}
	return a.auth.Authenticate(record, secret)
}

func (a *Application) ChannelStatus(channel string) (string, error) {
	record, err := a.db.GetChannel(model.NormalizeChannel(channel))
	if err != nil {
		return "", err
	}
	return security.ChannelStatus(record), nil
}

func (a *Application) Receive(wire codec.WireMessage) (model.ReceiveResult, error) {
	wire.ChannelNumber = model.NormalizeChannel(wire.ChannelNumber)
	wire.BatchNumber = model.NormalizeBatch(wire.BatchNumber)
	wire.Nonce = model.NormalizeNonce(wire.Nonce)
	message := codec.Decode(wire)
	if err := message.Validate(); err != nil {
		return a.reject(message, "malformed", err), nil
	}
	channel, err := a.db.GetChannel(message.ChannelNumber)
	if err != nil {
		return a.reject(message, "unknown-channel", security.ErrUnknownChannel), nil
	}
	if !a.auth.IsEligible(channel) {
		return a.reject(message, "inactive-channel", security.ErrInactiveChannel), nil
	}
	if !codec.VerifyTag(wire, channel.Secret) {
		return a.reject(message, "bad-tag", security.ErrBadTag), nil
	}
	batch, err := a.db.GetBatch(message.ChannelNumber, message.BatchNumber)
	if err != nil {
		return a.reject(message, "unknown-batch", err), nil
	}
	if err := a.policy.Check(message, batch); err != nil {
		return a.reject(message, "policy", err), nil
	}
	if err := a.guard.Reserve(message.ChannelNumber, message.Nonce); err != nil {
		if errors.Is(err, security.ErrDuplicateNonce) {
			return a.reject(message, "duplicate-nonce", err), nil
		}
		return model.ReceiveResult{}, err
	}
	a.sequence++
	message.Sequence = a.sequence
	if err := a.db.UpdateBatch(message.ChannelNumber, message.BatchNumber, func(current *model.BatchRecord) error {
		current.Accepted++
		current.LastSequence = message.Sequence
		return nil
	}); err != nil {
		return model.ReceiveResult{}, err
	}
	entry := model.AuditEntry{ID: fmt.Sprintf("%s-%03d", model.AuditID(message.ChannelNumber, message.BatchNumber, message.Nonce), message.Sequence), Channel: message.ChannelNumber, Batch: message.BatchNumber, Nonce: message.Nonce, Outcome: "accepted", Reason: "authenticated", Sequence: message.Sequence, Fingerprint: model.Fingerprint(message)}
	if err := a.db.SaveAudit(entry); err != nil {
		return model.ReceiveResult{}, err
	}
	return model.ReceiveResult{Accepted: true, Outcome: "accepted", Reason: "authenticated", Entry: entry}, nil
}

func (a *Application) reject(message model.ProtectedMessage, outcome string, reason error) model.ReceiveResult {
	a.sequence++
	entry := model.AuditEntry{ID: fmt.Sprintf("%s-%03d", model.AuditID(message.ChannelNumber, message.BatchNumber, message.Nonce), a.sequence), Channel: message.ChannelNumber, Batch: message.BatchNumber, Nonce: message.Nonce, Outcome: "rejected", Reason: outcome + ": " + reason.Error(), Sequence: a.sequence, Fingerprint: model.Fingerprint(message)}
	_ = a.db.SaveAudit(entry)
	if message.BatchNumber != "" && message.ChannelNumber != "" {
		_ = a.db.UpdateBatch(message.ChannelNumber, message.BatchNumber, func(batch *model.BatchRecord) error { batch.Rejected++; return nil })
	}
	return model.ReceiveResult{Accepted: false, Outcome: "rejected", Reason: entry.Reason, Entry: entry}
}

func (a *Application) Audit(channel string) ([]model.AuditEntry, error) {
	return a.db.ListChannelAudits(model.NormalizeChannel(channel))
}

func (a *Application) Batch(channel, number string) (model.BatchRecord, error) {
	return a.db.GetBatch(model.NormalizeChannel(channel), model.NormalizeBatch(number))
}

func (a *Application) GuardedNonces(channel string) []string {
	return a.guard.Snapshot(model.NormalizeChannel(channel))
}

func (a *Application) Describe(channel string) string {
	status, err := a.ChannelStatus(channel)
	if err != nil {
		return "unavailable"
	}
	return strings.ToLower(status)
}
