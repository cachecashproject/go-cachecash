package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	_ "net/http/pprof"
	"os"

	"github.com/cachecashproject/go-cachecash/cache"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller   = flag.Bool("logCaller", false, "Enable method name logging")
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
	l.SetReportCaller(*logCaller)

	cf, err := loadConfigFile(*configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}

	c, err := cache.NewCache(l, cf.BadgerDirectory)
	if err != nil {
		return nil
	}
	defer c.Close()

	c.Escrows = cf.Escrows

	app, err := cache.NewApplication(l, c, cf.Config)
	if err != nil {
		return errors.Wrap(err, "failed to create cache application")
	}

	if err := common.RunStarterShutdowner(app); err != nil {
		return err
	}
	return nil
}
