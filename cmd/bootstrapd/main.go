package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/bootstrap"
	"github.com/cachecashproject/go-cachecash/bootstrap/migrations"
	"github.com/cachecashproject/go-cachecash/common"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

var (
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller   = flag.Bool("logCaller", false, "Enable method name logging")
	logFile     = flag.String("logFile", "", "Path where file should be logged")
	configPath  = flag.String("config", "bootstrapd.config.json", "Path to configuration file")
	traceAPI    = flag.String("trace", "", "Jaeger API for tracing")
)

func loadConfigFile(path string) (*bootstrap.ConfigFile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cf bootstrap.ConfigFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, err
	}

	return &cf, nil
}

func generateConfigFile(path string) error {
	cf := &bootstrap.ConfigFile{
		GrpcAddr: ":7777",
		Database: "./bootstrapd.db",
	}
	buf, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}

	err = ioutil.WriteFile(path, buf, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to write config")
	}

	return nil
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
	if err := common.ConfigureLogger(l, &common.LoggerConfig{
		LogLevelStr: *logLevelStr,
		LogCaller:   *logCaller,
		LogFile:     *logFile,
		Json:        true,
	}); err != nil {
		return errors.Wrap(err, "failed to configure logger")
	}
	l.Info("Starting CacheCash bootstrapd ", cachecash.CurrentVersion)

	defer common.SetupTracing(*traceAPI, "cachecash-bootstrapd", l).Flush()

	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		l.Info("config doesn't exist, generating")
		if err := generateConfigFile(*configPath); err != nil {
			return err
		}
	}

	cf, err := loadConfigFile(*configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
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

	b, err := bootstrap.NewBootstrapd(l, db)
	if err != nil {
		return nil
	}

	app, err := bootstrap.NewApplication(l, b, cf)
	if err != nil {
		return errors.Wrap(err, "failed to create cache application")
	}

	if err := common.RunStarterShutdowner(l, app); err != nil {
		return err
	}
	return nil
}
