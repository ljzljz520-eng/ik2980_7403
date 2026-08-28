package security

import (
	"fmt"
	"strings"

	"emergencycomms/internal/model"
)

type ChannelAuthenticator struct{}

func NewChannelAuthenticator() *ChannelAuthenticator { return &ChannelAuthenticator{} }

func (a *ChannelAuthenticator) Authenticate(channel model.ChannelRecord, provided string) error {
	if err := channel.Validate(); err != nil {
		return err
	}
	if !channel.Active {
		return ErrInactiveChannel
	}
	if strings.TrimSpace(provided) == "" {
		return fmt.Errorf("channel credential is required")
	}
	if provided != channel.Secret {
		return fmt.Errorf("channel credential rejected")
	}
	return nil
}

func (a *ChannelAuthenticator) IsEligible(channel model.ChannelRecord) bool {
	return channel.Number != "" && channel.Secret != "" && channel.Active
}

func ChannelStatus(channel model.ChannelRecord) string {
	if !channel.Active {
		return "inactive"
	}
	if channel.Revision == 0 {
		return "active-unversioned"
	}
	return fmt.Sprintf("active-revision-%d", channel.Revision)
}

func RotateSecret(channel *model.ChannelRecord, next string) error {
	if strings.TrimSpace(next) == "" {
		return fmt.Errorf("new secret is required")
	}
	channel.Secret = next
	channel.Revision++
	return nil
}

func Deactivate(channel *model.ChannelRecord) { channel.Active = false }

func Activate(channel *model.ChannelRecord) { channel.Active = true }
