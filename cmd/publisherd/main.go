package main

import (
	_ "net/http/pprof"

	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/publisher"
	"github.com/sirupsen/logrus"
)

func main() {
	// TODO: Configure logger.
	l := logrus.New()

	// TODO: temporary
	prov, err := makePublisher()
	if err != nil {
		panic(err)
	}

	conf := &publisher.Config{}

	// Serve traffic!
	a, err := publisher.NewApplication(l, prov, conf)
	if err != nil {
		panic(err)
	}
	if err := common.RunStarterShutdowner(a); err != nil {
		panic(err)
	}
}
