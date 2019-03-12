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

	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/publisher"
	"github.com/cachecashproject/go-cachecash/publisher/migrations"
	"github.com/cachecashproject/go-cachecash/publisher/models"
	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
)

var (
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller   = flag.Bool("logCaller", false, "Enable method name logging")
	configPath  = flag.String("config", "publisher.config.json", "Path to configuration file")
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
	logLevel, err := logrus.ParseLevel(*logLevelStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse log level")
	}
	l.SetLevel(logLevel)
	l.SetReportCaller(*logCaller)

	cf, err := loadConfigFile(*configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}

	upstream, err := catalog.NewHTTPUpstream(l, cf.UpstreamURL, cf.Config.DefaultCacheDuration)
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

	p, err := publisher.NewContentPublisher(l, db, cat, cf.PrivateKey)
	if err != nil {
		return errors.Wrap(err, "failed to create publisher")
	}

	for _, e := range cf.Escrows {
		l.Infof("Adding escrow from config to database: %v", e)
		if err = p.AddEscrow(e); err != nil {
			return errors.Wrap(err, "failed to add escrow to publisher")
		}
		err = e.Inner.Insert(context.TODO(), db, boil.Infer())
		if err != nil {
			return errors.Wrap(err, "failed to add escrow to database")
		}

		for _, c := range e.Caches {
			l.Infof("Adding cache from config to database: %v", c)
			err = c.Cache.Upsert(context.TODO(), db, true, []string{"public_key"}, boil.Whitelist("inetaddr", "port"), boil.Infer())
			if err != nil {
				return errors.Wrap(err, "failed to add cache to database")
			}

			ec := models.EscrowCache{
				EscrowID:       e.Inner.ID,
				CacheID:        c.Cache.ID,
				InnerMasterKey: c.InnerMasterKey,
			}
			err = ec.Upsert(context.TODO(), db, false, []string{"escrow_id", "cache_id"}, boil.Whitelist("inner_master_key"), boil.Infer())
			if err != nil {
				return errors.Wrap(err, "failed to link cache to escrow")
			}
		}
	}

	app, err := publisher.NewApplication(l, p, cf.Config)
	if err != nil {
		return errors.Wrap(err, "failed to create cache application")
	}

	if err := common.RunStarterShutdowner(app); err != nil {
		return err
	}
	return nil
}
