package log

import (
	"github.com/sirupsen/logrus"
)

// Hook is a log emitter compatible with logrus; speaks to a client to send its data over the wire
type Hook struct {
	client *Client
}

// NewHook initializes a logrus Hook.
func NewHook(c *Client) logrus.Hook {
	return &Hook{client: c}
}

// Levels returns the levels appropriate for this hook.
func (l *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends the event to our client.
func (l *Hook) Fire(e *logrus.Entry) error {
	return l.client.Write(e)
}
