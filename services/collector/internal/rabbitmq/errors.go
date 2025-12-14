package rabbitmq

import "errors"

var (
	ErrChannelClosed = errors.New("rabbitmq channel closed")
)
