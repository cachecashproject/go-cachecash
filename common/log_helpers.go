package common

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/client9/reopen"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// LoggerConfig describes the configuration of logging for a program.
type LoggerConfig struct {
	logrus.Logger
	LogLevelStr string
	LogCaller   bool
	LogFile     string
	JSON        bool
}

// LogOpt is used to parameterise NewCLILogger in an extensible fashion.
// For instance, to make JSON logging the default:
//
// ```
// NewCLILogger(LogOpt{Json:true})
// ```
type LogOpt struct {
	JSON bool
}

// NewCLILogger creates a new CLI logger. Specifically this:
// - adds CLI flags for configuring the logging system
// - instantiates a Logrus logger
// - returns a struct with CLI flags registered and ready to be parsed by flag.Parse
func NewCLILogger(opts ...LogOpt) *LoggerConfig {
	result := LoggerConfig{Logger: *logrus.New()}
	// Accumulate any options into a single struct
	options := LogOpt{}
	for _, opt := range opts {
		if opt.JSON {
			options.JSON = opt.JSON
		}
	}
	// Use the accumulated options to override defaults for options where we have
	// variability.
	flag.StringVar(&result.LogLevelStr, "logLevel", "info", "Verbosity of log output")
	flag.BoolVar(&result.LogCaller, "logCaller", false, "Enable method name logging")
	flag.StringVar(&result.LogFile, "logFile", "", "Path where file should be logged")
	flag.BoolVar(&result.LogCaller, "logJSON", options.JSON, "Log in JSON")
	return &result
}

// ConfigureLogger configures logging from command line parameters.
func (c *LoggerConfig) ConfigureLogger() error {
	l := &c.Logger
	log.SetFlags(0)

	logLevel, err := logrus.ParseLevel(c.LogLevelStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse log level")
	}
	l.SetLevel(logLevel)
	l.SetReportCaller(c.LogCaller)

	if c.JSON {
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
