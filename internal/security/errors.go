package security

import "errors"

var (
	ErrDuplicateNonce  = errors.New("duplicate nonce")
	ErrUnknownChannel  = errors.New("unknown channel")
	ErrInactiveChannel = errors.New("channel inactive")
	ErrBadTag          = errors.New("invalid authentication tag")
	ErrBatchClosed     = errors.New("batch closed")
)
