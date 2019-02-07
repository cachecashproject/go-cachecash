package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	_ "net/http/pprof"
	"os"

	"github.com/cachecashproject/go-cachecash/cache"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
	configPath  = flag.String("config", "cache.config.json", "Path to configuration file")
)

func loadConfigFile(path string) (*cache.ConfigFile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cf cache.ConfigFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, err
	}

	return &cf, nil
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

	cf, err := loadConfigFile(*configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}

	/*
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
	*/
	_ = cf
	return nil
}
