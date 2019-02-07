package main

import (
	"github.com/cachecashproject/go-cachecash/publisher"
	"github.com/cachecashproject/go-cachecash/testdatagen"
	"github.com/sirupsen/logrus"
)

// TEMP: Cribbed from `integration_test.go`.
func makePublisher() (*publisher.ContentPublisher, error) {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)

	scen, err := testdatagen.GenerateTestScenario(l, &testdatagen.TestScenarioParams{
		BlockSize:  128 * 1024,
		ObjectSize: 128 * 1024 * 16,
	})
	if err != nil {
		return nil, err
	}

	return scen.Publisher, nil
}
