package main

import (
	"database/sql"
	"flag"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/bootstrap"
	"github.com/cachecashproject/go-cachecash/bootstrap/migrations"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

var (
	configPath = flag.String("config", "bootstrapd.config.json", "Path to configuration file")
	traceAPI   = flag.String("trace", "", "Jaeger API for tracing")
	proxy      = flag.Bool("proxy-protocol", false, "Enable PROXY protocol support")
)

func loadConfigFile(l *logrus.Logger, path string) (*bootstrap.ConfigFile, error) {
	conf := bootstrap.ConfigFile{}
	p, err := common.NewConfigParser(l, "bootstrap")
	if err != nil {
		return nil, err
	}
	err = p.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf.GrpcAddr = p.GetString("grpc_addr", ":7777")
	conf.Database = p.GetString("database", "./bootstrapd.db")
	conf.StatusAddr = p.GetString("status_addr", ":8100")
	conf.ProxyProtocol = p.GetBool("proxy_protocol", *proxy)
	conf.Insecure = p.GetInsecure()

	return &conf, nil
}

func main() {
	common.Main(mainC)
}

func mainC() error {
	l := log.NewCLILogger("bootstrapd", log.CLIOpt{JSON: true})
	flag.Parse()

	cf, err := loadConfigFile(&l.Logger, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}

	if err := l.ConfigureLogger(cf.Insecure); err != nil {
		return errors.Wrap(err, "failed to configure logger")
	}
	l.Info("Starting CacheCash bootstrapd ", cachecash.CurrentVersion)

	defer common.SetupTracing(*traceAPI, "cachecash-bootstrapd", &l.Logger).Flush()

	db, err := sql.Open("sqlite3", cf.Database)
	if err != nil {
		return errors.Wrap(err, "failed to open database")
	}

	l.Info("applying migrations")
	n, err := migrate.Exec(db, "sqlite3", migrations.Migrations, migrate.Up)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}
	l.Infof("applied %d migrations", n)

	b, err := bootstrap.NewBootstrapd(&l.Logger, db)
	if err != nil {
		return nil
	}

	app, err := bootstrap.NewApplication(&l.Logger, b, cf)
	if err != nil {
		return errors.Wrap(err, "failed to create bootstrap application")
	}

	if err := common.RunStarterShutdowner(&l.Logger, app); err != nil {
		return err
	}
	return nil
}
