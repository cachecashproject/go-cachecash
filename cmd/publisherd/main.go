package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	_ "net/http/pprof"
	"os"

	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/publisher"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller   = flag.Bool("logCaller", false, "Enable method name logging")
	configPath  = flag.String("config", "publisher.config.json", "Path to configuration file")
)

func loadConfigFile(path string) (*publisher.ConfigFile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cf publisher.ConfigFile
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

	upstream, err := catalog.NewHTTPUpstream(l, cf.UpstreamURL)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP upstream")
	}

	cat, err := catalog.NewCatalog(l, upstream)
	if err != nil {
		return errors.Wrap(err, "failed to create catalog")
	}

	p, err := publisher.NewContentPublisher(l, cat, cf.PrivateKey)
	if err != nil {
		return errors.Wrap(err, "failed to create publisher")
	}

	for _, e := range cf.Escrows {
		if err := p.AddEscrow(e); err != nil {
			return errors.Wrap(err, "failed to add escrow to publisher")
		}
	}

	app, err := publisher.NewApplication(l, p, cf.Config)
	if err != nil {
		return errors.Wrap(err, "failed to create cache application")
	}

	if err := common.RunStarterShutdowner(app); err != nil {
		return err
	}
	return nil
}
