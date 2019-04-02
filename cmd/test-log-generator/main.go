package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/client9/reopen"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller   = flag.Bool("logCaller", false, "Enable method name logging")
	logFile     = flag.String("logFile", "", "Path where file should be logged")
)

func generateMessage() string {
	words := []string{"Alfa", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot", "Golf", "Hotel", "India", "Juliett",
		"Kilo", "Lima", "Mike", "November", "Oscar", "Papa", "Quebec", "Romeo", "Sierra", "Tango", "Uniform", "Victor",
		"Whiskey", "X-ray", "Yankee", "Zulu"}

	var parts []string
	for i := 0; i < 5; i++ {
		parts = append(parts, words[rand.Intn(len(words))])
	}
	return strings.Join(parts, " ")
}

func main() {
	if err := mainC(); err != nil {
		if _, err := os.Stderr.WriteString(err.Error() + "\n"); err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}

func mainC() error {
	flag.Parse()
	log.SetFlags(0)

	l := logrus.New()
	logLevel, err := logrus.ParseLevel(*logLevelStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse log level")
	}
	l.SetLevel(logLevel)
	l.SetReportCaller(*logCaller)

	l.SetFormatter(&logrus.JSONFormatter{})

	if *logFile != "" {
		if *logFile == "-" {
			l.SetOutput(os.Stdout)
		} else {
			f, err := reopen.NewFileWriter(*logFile)
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

	l.Info("ready to spew test log messages")

	// XXX: No way to trigger this right now; if we need a graceful shutdown we could hook this up.
	quit := make(chan struct{})
	var idx int
	for {
		select {
		case <-time.After(time.Duration(rand.Intn(5)) * time.Second):
			l.WithFields(logrus.Fields{"idx": idx}).Info(generateMessage())
			idx++
		case <-quit:
			return nil
		}
	}
}
