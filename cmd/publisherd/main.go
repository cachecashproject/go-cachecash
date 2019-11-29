package main

import (
	"context"
	"database/sql"
	"flag"
	_ "net/http/pprof"
	"time"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/dbtx"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/log"
	"github.com/cachecashproject/go-cachecash/publisher"
	"github.com/cachecashproject/go-cachecash/publisher/migrations"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

var (
	configPath  = flag.String("config", "publisher.config.json", "Path to configuration file")
	keypairPath = flag.String("keypair", "publisher.keypair.json", "Path to keypair file")
	traceAPI    = flag.String("trace", "", "Jaeger API for tracing")
)

func loadConfigFile(l *logrus.Logger, path string) (*publisher.ConfigFile, error) {
	conf := publisher.ConfigFile{}
	p, err := common.NewConfigParser(l, "publisher")
	if err != nil {
		return nil, err
	}
	err = p.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf.Origin = p.GetString("origin", "localhost")
	conf.PublisherAddr = p.GetString("publisher_addr", "")
	conf.GrpcAddr = p.GetString("grpc_addr", ":7070")
	conf.StatusAddr = p.GetString("status_addr", ":8100")
	conf.BootstrapAddr = p.GetString("bootstrap_addr", "bootstrapd:7777")
	conf.DefaultCacheDuration = time.Duration(p.GetInt64("default_cache_duration", 300)) * time.Second

	conf.UpstreamURL = p.GetString("upstream", "")
	conf.Database = p.GetString("database", "host=publisher-db port=5432 user=postgres dbname=publisher sslmode=disable")
	conf.Insecure = p.GetInsecure()

	return &conf, nil
}

func main() {
	common.Main(mainC)
}

func mainC() error {
	l := log.NewCLILogger("publisherd", log.CLIOpt{JSON: true})
	flag.Parse()

	cf, err := loadConfigFile(&l.Logger, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}

	kp, err := keypair.LoadOrGenerate(&l.Logger, *keypairPath)
	if err != nil {
		return errors.Wrap(err, "failed to get keypair")
	}

	if err := l.ConfigureLogger(); err != nil {
		return errors.Wrap(err, "failed to configure logger")
	}
	l.Info("Starting CacheCash publisherd ", cachecash.CurrentVersion)

	defer common.SetupTracing(*traceAPI, "cachecash-publisherd", &l.Logger).Flush()

	upstream, err := catalog.NewHTTPUpstream(&l.Logger, cf.UpstreamURL, cf.DefaultCacheDuration)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP upstream")
	}

	cat, err := catalog.NewCatalog(&l.Logger, upstream)
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

	var publisherAddr string
	if len(cf.PublisherAddr) == 0 {
		publisherAddr = cf.GrpcAddr
	} else {
		publisherAddr = cf.PublisherAddr
	}
	p, err := publisher.NewContentPublisher(&l.Logger, publisherAddr, cat, kp.PrivateKey)
	if err != nil {
		return errors.Wrap(err, "failed to create publisher")
	}

	ctx := dbtx.ContextWithExecutor(context.Background(), db)
	num, err := p.LoadFromDatabase(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load state from database")
	}
	l.Infof("loaded %d escrows from database", num)

	app, err := publisher.NewApplication(&l.Logger, p, db, cf)
	if err != nil {
		return errors.Wrap(err, "failed to create cache application")
	}

	if err := common.RunStarterShutdowner(&l.Logger, app); err != nil {
		return err
	}
	return nil
}
