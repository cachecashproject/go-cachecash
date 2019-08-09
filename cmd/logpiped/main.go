package main

import (
	"flag"
	"io/ioutil"
	"net"

	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/log"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	spoolDir      = flag.String("spooldir", "", "Base path to spool files to on disk")
	listenAddress = flag.String("listenaddress", ":1234", "Listening address")
	maxLogSize    = flag.Int64("maxlogsize", 1024*1024*1024, "Maximum size of log bundle to accept")
)

func main() {
	common.Main(mainC)
}

func mainC() error {
	flag.Parse()

	if *spoolDir == "" {
		var err error
		*spoolDir, err = ioutil.TempDir("", "")
		if err != nil {
			return err
		}

		logrus.Warnf("No spool dir was picked: %q was created for the purpose", *spoolDir)
	}

	ps, err := log.NewPipeServer(*spoolDir, *maxLogSize)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	log.RegisterLogPipeServer(s, ps)

	l, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		return err
	}
	return s.Serve(l)
}
