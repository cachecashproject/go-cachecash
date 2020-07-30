package main

import (
	"flag"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/blockexplorer"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/log"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	configPath = flag.String("config", "blockexplorer.config.json", "Path to configuration file")
	traceAPI   = flag.String("trace", "", "Jaeger API for tracing")
)

func loadConfigFile(l *logrus.Logger, path string) (*blockexplorer.ConfigFile, error) {
	conf := blockexplorer.ConfigFile{}
	p, err := common.NewConfigParser(l, "blockexplorer")
	if err != nil {
		return nil, err
	}
	err = p.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf.LedgerAddr = p.GetString("ledger_addr", "")
	conf.StatusAddr = p.GetString("status_addr", ":8100")
	conf.HTTPAddr = p.GetString("web_addr", ":8080")
	conf.Insecure = p.GetInsecure()
	conf.Root = p.GetString("root", "")

	return &conf, nil
}

func main() {
	common.Main(mainC)
}

func mainC() error {
	l := log.NewCLILogger("blockexplorerd", log.CLIOpt{JSON: true})
	flag.Parse()

	cf, err := loadConfigFile(&l.Logger, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}

	if err := l.ConfigureLogger(); err != nil {
		return errors.Wrap(err, "failed to configure logger")
	}
	l.Info("Starting CacheCash blockexplorerd ", cachecash.CurrentVersion)

	if len(cf.LedgerAddr) == 0 {
		return errors.New("Must supply a cachecash ledger endpoint via the ledger_addr setting")
	}

	if len(cf.Root) == 0 {
		return errors.New("Must supply an HTTP url root via the root setting")
	}

	defer common.SetupTracing(*traceAPI, "cachecash-blockexplorerd", &l.Logger).Flush()

	gin.SetMode(gin.ReleaseMode)
	app, err := blockexplorer.NewDaemon(&l.Logger, cf)
	if err != nil {
		return errors.Wrap(err, "failed to create blockexplorer server")
	}

	if err := common.RunStarterShutdowner(&l.Logger, app); err != nil {
		return err
	}
	return nil
}
