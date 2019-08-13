package main

import (
	"flag"

	"github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/metricsproxy"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	configPath = flag.String("config", "metrics-proxy.config.json", "Path to configuration file")
	traceAPI   = flag.String("trace", "", "Jaeger API for tracing")
)

func loadConfigFile(l *logrus.Logger, path string) (*metricsproxy.ConfigFile, error) {
	conf := metricsproxy.ConfigFile{}
	p := common.NewConfigParser(l, "metrics-proxy")
	err := p.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf.MetricsGRPCAddr = p.GetString("grpc_addr", ":8000")
	conf.StatusAddr = p.GetString("status_addr", ":8100")

	return &conf, nil
}

func main() {
	common.Main(mainC)
}

func mainC() error {
	l := common.NewCLILogger(common.LogOpt{JSON: true})
	flag.Parse()

	if err := l.ConfigureLogger(); err != nil {
		return errors.Wrap(err, "failed to configure logger")
	}
	l.Info("Starting CacheCash metrics proxy ", cachecash.CurrentVersion)

	defer common.SetupTracing(*traceAPI, "cachecash-metrics proxy", &l.Logger).Flush()

	cf, err := loadConfigFile(&l.Logger, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}

	app, err := metricsproxy.NewApplication(&l.Logger, cf)
	if err != nil {
		return errors.Wrap(err, "failed to create metrics application")
	}

	if err := common.RunStarterShutdowner(&l.Logger, app); err != nil {
		return err
	}
	return nil
}
