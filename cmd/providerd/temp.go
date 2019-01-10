package main

import (
	"github.com/kelleyk/go-cachecash/provider"
	"github.com/kelleyk/go-cachecash/testutil"
	"github.com/sirupsen/logrus"
)

// TEMP: Cribbed from `integration_test.go`.
func makeProvider() (*provider.ContentProvider, error) {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)

	scen, err := testutil.GenerateTestScenario(l, &testutil.TestScenarioParams{
		BlockSize:  128 * 1024,
		ObjectSize: 128 * 1024 * 16,
	})
	if err != nil {
		return nil, err
	}

	return scen.Provider, nil
}
