package main

import (
	_ "net/http/pprof"

	"github.com/cachecashproject/go-cachecash/cache"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/sirupsen/logrus"
)

func main() {
	// TODO: Configure logger.
	l := logrus.New()

	// TODO: temporary
	c, err := makeCache()
	if err != nil {
		panic(err)
	}

	conf := &cache.Config{
		// Any non-defaults should be specified here!
	}

	// Serve traffic!
	a, err := cache.NewApplication(l, c, conf)
	if err != nil {
		panic(err)
	}
	if err := common.RunStarterShutdowner(a); err != nil {
		panic(err)
	}
}
