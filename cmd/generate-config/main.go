package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cachecashproject/go-cachecash/cache"
	"github.com/cachecashproject/go-cachecash/publisher"
	"github.com/cachecashproject/go-cachecash/testdatagen"
)

var (
	logLevelStr               = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller                 = flag.Bool("logCaller", false, "Enable method name logging")
	outputPath                = flag.String("outputPath", ".", "Directory where configuration files will be written")
	upstream                  = flag.String("upstream", "http://localhost:8081", "Upstream url")
	publisherCacheServiceAddr = flag.String("publisherCacheServiceAddr", "", "Publisher cache service address")
)

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

	scen, err := testdatagen.GenerateTestScenario(l, &testdatagen.TestScenarioParams{
		L:                         l,
		BlockSize:                 128 * 1024,
		ObjectSize:                128 * 1024 * 16,
		PublisherCacheServiceAddr: *publisherCacheServiceAddr,
	})
	if err != nil {
		return err
	}

	// Generate cache-side configuration files.
	for i, c := range scen.Caches {
		cf := cache.ConfigFile{
			Config: &cache.Config{
				// XXX: This must match what is set up in the Escrow struct on the publisher side so that the publisher
				// sends clients to the right place.
				ClientProtocolAddr: fmt.Sprintf(":%v", 9000+i),
				StatusAddr:         fmt.Sprintf(":%v", 9100+i),
			},
			Escrows: c.Escrows,
		}

		buf, err := json.MarshalIndent(cf, "", "  ")
		if err != nil {
			return err
		}

		path := filepath.Join(*outputPath, fmt.Sprintf("cache-%v.config.json", i))
		l.Debugf("writing cache configuration: %v", path)
		if err := ioutil.WriteFile(path, buf, 0644); err != nil {
			return err
		}
	}

	// Generate publisher-side configuration file.
	cf := &publisher.ConfigFile{
		Config: &publisher.Config{},
		Escrows: []*publisher.Escrow{
			scen.Escrow,
		},
		UpstreamURL: *upstream,
		PrivateKey:  scen.PublisherPrivateKey,
	}

	buf, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(*outputPath, "publisher.config.json")
	l.Debugf("writing publisher configuration: %v", path)
	if err := ioutil.WriteFile(path, buf, 0644); err != nil {
		return err
	}

	return nil
}
