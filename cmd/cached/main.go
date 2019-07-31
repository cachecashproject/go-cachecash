package main

import (
	"context"
	"database/sql"
	"flag"
	_ "net/http/pprof"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/cache"
	"github.com/cachecashproject/go-cachecash/cache/migrations"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/keypair"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

var (
	configPath  = flag.String("config", "cache.config.json", "Path to configuration file")
	keypairPath = flag.String("keypair", "cache.keypair.json", "Path to keypair file")
	traceAPI    = flag.String("trace", "", "Jaeger API for tracing")
)

func loadConfigFile(l *logrus.Logger, path string) (*cache.ConfigFile, error) {
	conf := cache.ConfigFile{}
	p := common.NewConfigParser(l, "cache")
	err := p.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf.ClientProtocolGrpcAddr = p.GetString("grpc_addr", ":9000")
	conf.ClientProtocolHttpAddr = p.GetString("http_addr", ":9443")
	conf.StatusAddr = p.GetString("status_addr", ":9100")
	conf.BootstrapAddr = p.GetString("bootstrap_addr", "bootstrapd:7777")

	conf.BadgerDirectory = p.GetString("badger_directory", "./chunks/")
	conf.Database = p.GetString("database", "cache.db")
	conf.ContactUrl = p.GetString("contact_url", "")

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
	l.Info("Starting CacheCash cached ", cachecash.CurrentVersion)

	defer common.SetupTracing(*traceAPI, "cachecash-cached", &l.Logger).Flush()

	cf, err := loadConfigFile(&l.Logger, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}
	kp, err := keypair.LoadOrGenerate(&l.Logger, *keypairPath)
	if err != nil {
		return errors.Wrap(err, "failed to get keypair")
	}

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

	c, err := cache.NewCache(&l.Logger, db, cf, kp)
	if err != nil {
		return err
	}
	defer c.Close()

	num, err := c.LoadFromDatabase(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to load state from database")
	}
	l.WithFields(logrus.Fields{
		"len(escrows)": num,
	}).Info("loaded escrows from database")

	app, err := cache.NewApplication(&l.Logger, c, cf)
	if err != nil {
		return errors.Wrap(err, "failed to create cache application")
	}

	if err := common.RunStarterShutdowner(&l.Logger, app); err != nil {
		return err
	}
	return nil
}
