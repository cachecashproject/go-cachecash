package common

import (
	"time"

	"github.com/sirupsen/logrus"
)

type TelemetryTimer struct {
	l         *logrus.Logger
	startTime time.Time
	label     string
}

func StartTelemetryTimer(l *logrus.Logger, label string) *TelemetryTimer {
	return &TelemetryTimer{
		l:         l,
		startTime: time.Now(),
		label:     label,
	}
}

func (tt *TelemetryTimer) Stop() {
	stopTime := time.Now()

	tt.l.WithFields(logrus.Fields{
		"kind": "timing",
		"when": []time.Time{tt.startTime, stopTime},
	}).Info(tt.label)
}
