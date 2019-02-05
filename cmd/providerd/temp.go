package main

import (
	"github.com/cachecashproject/go-cachecash/provider"
	"github.com/cachecashproject/go-cachecash/testdatagen"
	"github.com/sirupsen/logrus"
)

// TEMP: Cribbed from `integration_test.go`.
func makeProvider() (*provider.ContentProvider, error) {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)

	scen, err := testdatagen.GenerateTestScenario(l, &testdatagen.TestScenarioParams{
		BlockSize:  128 * 1024,
		ObjectSize: 128 * 1024 * 16,
	})
	if err != nil {
		return nil, err
	}

	return scen.Provider, nil
}
