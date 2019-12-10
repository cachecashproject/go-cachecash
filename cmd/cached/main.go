package main

import (
	"context"
	"database/sql"
	"flag"
	_ "net/http/pprof"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/cache"
	"github.com/cachecashproject/go-cachecash/cache/migrations"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/dbtx"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledgerclient"
	"github.com/cachecashproject/go-cachecash/log"
)

var (
	configPath  = flag.String("config", "cache.config.json", "Path to configuration file")
	keypairPath = flag.String("keypair", "cache.keypair.json", "Path to keypair file")
	traceAPI    = flag.String("trace", "", "Jaeger API for tracing")
)

func loadConfigFile(l *logrus.Logger, path string) (*cache.ConfigFile, error) {
	conf := cache.ConfigFile{}
	p, err := common.NewConfigParser(l, "cache")
	if err != nil {
		return nil, err
	}
	err = p.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf.ClientProtocolGrpcAddr = p.GetString("grpc_addr", ":9000")
	conf.ClientProtocolHttpAddr = p.GetString("http_addr", ":9443")
	conf.StatusAddr = p.GetString("status_addr", ":9100")
	conf.BootstrapAddr = p.GetString("bootstrap_addr", "bootstrapd:7777")
	conf.LedgerAddr = p.GetString("ledger_addr", "ledger:7778")

	conf.BadgerDirectory = p.GetString("badger_directory", "./chunks/")
	conf.Database = p.GetString("database", "cache.db")
	conf.ContactUrl = p.GetString("contact_url", "")
	conf.MetricsEndpoint = p.GetString("metrics_endpoint", "")
	conf.SyncInterval = p.GetSeconds("sync-interval", ledgerclient.DEFAULT_SYNC_INTERVAL)
	conf.Insecure = p.GetInsecure()

	return &conf, nil
}

func main() {
	common.Main(mainC)
}

func mainC() error {
	l := log.NewCLILogger("cached", log.CLIOpt{JSON: true})
	flag.Parse()

	cf, err := loadConfigFile(&l.Logger, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}

	if err := l.ConfigureLogger(); err != nil {
		return errors.Wrap(err, "failed to configure logger")
	}
	l.Info("Starting CacheCash cached ", cachecash.CurrentVersion)

	kp, err := keypair.LoadOrGenerate(&l.Logger, *keypairPath)
	if err != nil {
		return errors.Wrap(err, "failed to get keypair")
	}

	if err := l.Connect(cf.Insecure, kp); err != nil {
		return errors.Wrap(err, "failed to connect to logpipe")
	}

	defer common.SetupTracing(*traceAPI, "cachecash-cached", &l.Logger).Flush()

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

	persistence := ledger.NewChainStorageSQL(&l.Logger, ledger.NewChainStorageSqlite())
	if err := persistence.RunMigrations(db); err != nil {
		return err
	}
	storage := ledger.NewDatabase(persistence)
	r, err := ledgerclient.NewReplicator(&l.Logger, storage, cf.LedgerAddr, cf.Insecure)
	if err != nil {
		return errors.Wrap(err, "failed to create replicator")
	}

	c, err := cache.NewCache(&l.Logger, cf, kp)
	if err != nil {
		return err
	}
	defer c.Close()

	ctx := dbtx.ContextWithExecutor(context.Background(), db)
	num, err := c.LoadFromDatabase(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load state from database")
	}
	l.WithFields(logrus.Fields{
		"len(escrows)": num,
	}).Info("loaded escrows from database")

	app, err := cache.NewApplication(&l.Logger, c, db, cf, kp, r)
	if err != nil {
		return errors.Wrap(err, "failed to create cache application")
	}

	if err := common.RunStarterShutdowner(&l.Logger, app); err != nil {
		return err
	}
	return nil
}
