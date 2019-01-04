package main

import (
	_ "net/http/pprof"

	"github.com/kelleyk/go-cachecash/common"
	"github.com/kelleyk/go-cachecash/provider"
	"github.com/sirupsen/logrus"
)

func main() {
	// TODO: Configure logger.
	l := logrus.New()

	// TODO: temporary
	prov, err := makeProvider()
	if err != nil {
		panic(err)
	}

	conf := &provider.Config{}

	// Serve traffic!
	a, err := provider.NewApplication(l, prov, conf)
	if err != nil {
		panic(err)
	}
	if err := common.RunStarterShutdowner(a); err != nil {
		panic(err)
	}
}
