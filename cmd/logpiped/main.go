package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"net"
	"runtime"

	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/log"
	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	spoolDir        = flag.String("spooldir", "", "Base path to spool files to on disk")
	listenAddress   = flag.String("listenaddress", ":9005", "Listening address")
	maxLogSize      = flag.Int64("maxlogsize", log.DefaultLogSize, "Maximum size of log bundle to accept")
	maxConnections  = flag.Int("maxconnections", 0, "Maximum number of connections (default 0, unlimited)")
	maxSpoolDirSize = flag.Int64("maxspooldirsize", 0, "Maximum size in bytes of spool dir contents (default 0, unlimited)")
	processors      = flag.Int("processors", runtime.NumCPU()*4, "Total number of queue processors")
	esConfig        = flag.String("esconfig", "", "Path to elasticsearch configuration")
)

func main() {
	common.Main(mainC)
}

func mainC() error {
	flag.Parse()

	if len(flag.Args()) != 1 {
		return errors.New("Please provide an index name for elasticsearch")
	}

	if *spoolDir == "" {
		var err error
		*spoolDir, err = ioutil.TempDir("", "")
		if err != nil {
			return err
		}

		logrus.Warnf("No spool dir was picked: %q was created for the purpose", *spoolDir)
	}

	var realESConfig es.Config

	if *esConfig != "" {
		content, err := ioutil.ReadFile(*esConfig)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(content, &realESConfig); err != nil {
			return err
		}
	} else {
		realESConfig = es.Config{Addresses: []string{"http://127.0.0.1:9200"}}
	}

	ps, err := log.NewPipeServer(flag.Args()[0], *processors, *spoolDir, realESConfig)
	if err != nil {
		return err
	}

	ps.MaxLogSize = *maxLogSize
	ps.MaxConnections = *maxConnections
	ps.MaxSpoolDirSize = *maxSpoolDirSize

	s := grpc.NewServer()
	log.RegisterLogPipeServer(s, ps)

	l, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		return err
	}
	return s.Serve(l)
}
