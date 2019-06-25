package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"io/ioutil"
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
	"golang.org/x/crypto/ed25519"
)

var (
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller   = flag.Bool("logCaller", false, "Enable method name logging")
	logFile     = flag.String("logFile", "", "Path where file should be logged")
	configPath  = flag.String("config", "ledger.config.json", "Path to configuration file")
)

func loadConfigFile(path string) (*ledgerservice.ConfigFile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cf ledgerservice.ConfigFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, err
	}

	return &cf, nil
}

func GetenvDefault(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	} else {
		return defaultValue
	}
}

func generateConfigFile(path string) error {
	ledgerProtocolAddr := os.Getenv("LEDGER_PROTCOOL_ADDR")

	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}

	cf := &ledgerservice.ConfigFile{
		LedgerProtocolAddr: ledgerProtocolAddr,

		PrivateKey: privateKey,
		Database:   "host=ledger-db port=5432 user=postgres dbname=ledger sslmode=disable",
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
	l.Info("Starting CacheCash publisherd ", cachecash.CurrentVersion)

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
