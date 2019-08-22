package main

import (
	"flag"
	"math/rand"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/log"
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
	common.Main(mainC)
}

func mainC() error {
	l := log.NewCLILogger("test-log-generator", log.CLIOpt{JSON: true})
	flag.Parse()

	if err := l.ConfigureLogger(true); err != nil {
		return errors.Wrap(err, "failed to configure logger")
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
