package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/kv/migrations"
	"github.com/cachecashproject/go-cachecash/log"
	"github.com/cachecashproject/go-cachecash/log/server"
	es "github.com/elastic/go-elasticsearch/v7"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

var (
	spoolDir      = flag.String("spooldir", "", "Base path to spool files to on disk")
	listenAddress = flag.String("listenaddress", ":9005", "Listening address")
	maxLogSize    = flag.Uint64("maxlogsize", server.DefaultLogSize, "Maximum size of log bundle to accept")
	esConfig      = flag.String("esconfig", "", "Path to elasticsearch configuration")
	dbDSN         = flag.String("dsn", "host=kvstore-db dbname=kvstore port=5432 user=postgres sslmode=disable", "DSN to connect to postgres")
)

func main() {
	common.Main(mainC)
}

func mainC() error {
	cl := log.NewCLILogger("logpiped", log.CLIOpt{JSON: true})
	flag.Parse()

	if err := cl.ConfigureLogger(); err != nil {
		return err
	}

	if len(flag.Args()) != 1 {
		return errors.New("Please provide an index name for elasticsearch")
	}

	if *spoolDir == "" {
		var err error
		*spoolDir, err = ioutil.TempDir("", "")
		if err != nil {
			return err
		}

		logrus.Warnf("No spool dir was picked: %q was created for the purpose", *spoolDir)
	}

	var realESConfig es.Config

	if *esConfig != "" {
		content, err := ioutil.ReadFile(*esConfig)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(content, &realESConfig); err != nil {
			return err
		}
	} else {
		realESConfig = es.Config{Addresses: []string{"http://127.0.0.1:9200"}}
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

retry:
	db, err := sql.Open("postgres", *dbDSN)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		cl.Warnf("Database is down or unavailable, retrying connection in 1s")
		time.Sleep(time.Second)
		db.Close()
		goto retry
	}

	_, err = migrate.Exec(db, "postgres", migrations.Migrations, migrate.Up)
	if err != nil {
		return err
	}

	config := server.Config{
		KVStoreDB:     db,
		Logger:        &cl.Logger,
		KVMember:      hostname,
		IndexName:     flag.Args()[0],
		MaxLogSize:    *maxLogSize,
		SpoolDir:      *spoolDir,
		ListenAddress: *listenAddress,
		ElasticSearch: realESConfig,
	}

	ps, err := server.NewLogPipe(config)
	if err != nil {
		return err
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signalChan
		if err := ps.Close(30 * time.Second); err != nil {
			logrus.Errorf("Error during shutdown: %v", err)
			os.Exit(1)
		}

		os.Exit(0)
	}()

	return ps.Boot(make(chan struct{}))
}
