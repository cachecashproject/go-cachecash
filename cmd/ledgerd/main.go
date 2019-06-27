package main

import (
	"database/sql"
	"flag"
	"log"
	_ "net/http/pprof"
	"os"
	"time"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/ledgerservice"
	"github.com/cachecashproject/go-cachecash/ledgerservice/migrations"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

var (
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller   = flag.Bool("logCaller", false, "Enable method name logging")
	logFile     = flag.String("logFile", "", "Path where file should be logged")
	configPath  = flag.String("config", "ledger.config.json", "Path to configuration file")
	// keypairPath = flag.String("keypair", "ledger.keypair.json", "Path to keypair file") // XXX: Not used yet.
	traceAPI = flag.String("trace", "", "Jaeger API for tracing")
)

func loadConfigFile(l *logrus.Logger, path string) (*ledgerservice.ConfigFile, error) {
	conf := ledgerservice.ConfigFile{}
	p := common.NewConfigParser(l, "ledger")
	err := p.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf.LedgerProtocolAddr = p.GetString("ledger_addr", ":8080")
	conf.StatusAddr = p.GetString("status_addr", ":8100")
	conf.Database = p.GetString("database", "host=publisher-db port=5432 user=postgres dbname=publisher sslmode=disable")

	return &conf, nil
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
	l.Info("Starting CacheCash ledgerd ", cachecash.CurrentVersion)

	defer common.SetupTracing(*traceAPI, "cachecash-ledgerd", l).Flush()

	cf, err := loadConfigFile(l, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}

	db, err := sql.Open("postgres", cf.Database)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	// Connect to the database.
	deadline := time.Now().Add(5 * time.Minute)
	for {
		err = db.Ping()

		if err == nil {
			// connected successfully
			break
		} else if time.Now().Before(deadline) {
			// connection failed, try again
			l.Info("Connection failed, trying again shortly")
			time.Sleep(250 * time.Millisecond)
		} else {
			// connection failed too many times, giving up
			return errors.Wrap(err, "database ping failed")
		}
	}
	l.Info("connected to database")

	l.Info("applying migrations")
	n, err := migrate.Exec(db, "postgres", migrations.Migrations, migrate.Up)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}
	l.Infof("applied %d migrations", n)

	ls, err := ledgerservice.NewLedgerService(l, db)
	if err != nil {
		return errors.Wrap(err, "failed to create publisher")
	}

	app, err := ledgerservice.NewApplication(l, ls, cf)
	if err != nil {
		return errors.Wrap(err, "failed to create cache application")
	}

	if err := common.RunStarterShutdowner(l, app); err != nil {
		return err
	}
	return nil
}
