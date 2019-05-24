package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	_ "net/http/pprof"
	"os"
	"time"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/publisher"
	"github.com/cachecashproject/go-cachecash/publisher/migrations"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

var (
	logLevelStr   = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller     = flag.Bool("logCaller", false, "Enable method name logging")
	logFile       = flag.String("logFile", "", "Path where file should be logged")
	configPath    = flag.String("config", "publisher.config.json", "Path to configuration file")
	bootstrapAddr = flag.String("bootstrapd", "bootstrapd:7777", "Bootstrap service to use")
)

func loadConfigFile(path string) (*publisher.ConfigFile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cf publisher.ConfigFile
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
	publisherCacheAddr := os.Getenv("PUBLISHER_CACHE_ADDR")
	upstream := os.Getenv("PUBLISHER_UPSTREAM")
	bootstrapAddr := GetenvDefault("BOOTSTRAP_ADDR", *bootstrapAddr)

	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}

	cf := &publisher.ConfigFile{
		CacheProtocolAddr: publisherCacheAddr,
		BootstrapAddr:     bootstrapAddr,

		UpstreamURL: upstream,
		PrivateKey:  privateKey,
		Database:    "host=publisher-db port=5432 user=postgres dbname=publisher sslmode=disable",
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

	upstream, err := catalog.NewHTTPUpstream(l, cf.UpstreamURL, cf.DefaultCacheDuration)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP upstream")
	}

	cat, err := catalog.NewCatalog(l, upstream)
	if err != nil {
		return errors.Wrap(err, "failed to create catalog")
	}

	db, err := sql.Open("postgres", cf.Database)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

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

	p, err := publisher.NewContentPublisher(l, db, cf.CacheProtocolAddr, cat, cf.PrivateKey)
	if err != nil {
		return errors.Wrap(err, "failed to create publisher")
	}

	num, err := p.LoadFromDatabase(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to load state from database")
	}
	l.Infof("loaded %d escrows from database", num)

	app, err := publisher.NewApplication(l, p, cf)
	if err != nil {
		return errors.Wrap(err, "failed to create cache application")
	}

	if err := common.RunStarterShutdowner(l, app); err != nil {
		return err
	}
	return nil
}
