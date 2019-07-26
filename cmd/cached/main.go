package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
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
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller   = flag.Bool("logCaller", false, "Enable method name logging")
	logFile     = flag.String("logFile", "", "Path where file should be logged")
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
	flag.Parse()
	log.SetFlags(0)

	l := logrus.New()
	if err := common.ConfigureLogger(l, &common.LoggerConfig{
		LogLevelStr: *logLevelStr,
		LogCaller:   *logCaller,
		LogFile:     *logFile,
		Json:        true,
	}); err != nil {
		return errors.Wrap(err, "failed to configure logger")
	}
	l.Info("Starting CacheCash cached ", cachecash.CurrentVersion)

	defer common.SetupTracing(*traceAPI, "cachecash-cached", l).Flush()

	cf, err := loadConfigFile(l, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}
	kp, err := keypair.LoadOrGenerate(l, *keypairPath)
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

	c, err := cache.NewCache(l, db, cf, kp)
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

	app, err := cache.NewApplication(l, c, cf)
	if err != nil {
		return errors.Wrap(err, "failed to create cache application")
	}

	if err := common.RunStarterShutdowner(l, app); err != nil {
		return err
	}
	return nil
}
