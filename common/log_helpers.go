package common

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/client9/reopen"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type LoggerConfig struct {
	LogLevelStr string
	LogCaller   bool
	LogFile     string
	Json        bool
}

func ConfigureLogger(l *logrus.Logger, c *LoggerConfig) error {
	logLevel, err := logrus.ParseLevel(c.LogLevelStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse log level")
	}
	l.SetLevel(logLevel)
	l.SetReportCaller(c.LogCaller)

	if c.Json {
		l.SetFormatter(&logrus.JSONFormatter{})
	}

	if c.LogFile != "" {
		if c.LogFile == "-" {
			l.SetOutput(os.Stdout)
		} else {
			f, err := reopen.NewFileWriter(c.LogFile)
			if err != nil {
				return errors.Wrap(err, "unable to open log file")
			}
			l.SetOutput(f)

			sighupCh := make(chan os.Signal, 1)
			signal.Notify(sighupCh, syscall.SIGHUP)
			go func() {
				for {
					<-sighupCh
					if err := f.Reopen(); err != nil {
						l.WithError(err).Error("failed to reopen log file on SIGHUP")
					}
				}
			}()
		}
	}

	return nil
}
