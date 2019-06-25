package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	_ "net/http/pprof"
	"os"
	"time"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/publisher"
	"github.com/cachecashproject/go-cachecash/publisher/migrations"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

var (
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller   = flag.Bool("logCaller", false, "Enable method name logging")
	logFile     = flag.String("logFile", "", "Path where file should be logged")
	configPath  = flag.String("config", "publisher.config.json", "Path to configuration file")
	keypairPath = flag.String("keypair", "publisher.keypair.json", "Path to keypair file")
	traceAPI    = flag.String("trace", "", "Jaeger API for tracing")
)

func loadConfigFile(l *logrus.Logger, path string) (*publisher.ConfigFile, error) {
	conf := publisher.ConfigFile{}
	p := common.NewConfigParser(l, "publisher")

	conf.ClientProtocolAddr = p.GetString("client_grpc_addr", ":8080")
	conf.CacheProtocolAddr = p.GetString("cache_grpc_addr", ":8082")
	conf.StatusAddr = p.GetString("status_addr", ":8100")
	conf.BootstrapAddr = p.GetString("bootstrap_addr", "bootstrapd:7777")
	conf.DefaultCacheDuration = time.Duration(p.GetInt64("default_cache_duration", 300)) * time.Second

	conf.UpstreamURL = p.GetString("upstream", "")
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
	l.Info("Starting CacheCash publisherd ", cachecash.CurrentVersion)

	defer common.SetupTracing(*traceAPI, "cachecash-publisherd", l).Flush()

	cf, err := loadConfigFile(l, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}
	kp, err := keypair.LoadOrGenerate(l, *keypairPath)
	if err != nil {
		return errors.Wrap(err, "failed to get keypair")
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

	p, err := publisher.NewContentPublisher(l, db, cf.CacheProtocolAddr, cat, kp.PrivateKey)
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
