// +build external_test

package server

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/kv/migrations"
	"github.com/cachecashproject/go-cachecash/kv/ratelimit"
	"github.com/cachecashproject/go-cachecash/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	for _, val := range []string{"PSQL_HOST", "PSQL_DBNAME"} {
		if os.Getenv(val) == "" {
			panic(fmt.Sprintf("please provide %q in the env to test this module", val))
		}
	}
}

var testDSN = fmt.Sprintf(
	"host=%s port=5432 user=postgres dbname=%s sslmode=disable",
	os.Getenv("PSQL_HOST"),
	os.Getenv("PSQL_DBNAME"),
)

type logRecorder struct {
	entries []*log.Entry
	mutex   sync.Mutex
}

// this is mostly to keep race detector happy in asserts we know are safe. if
// you want to check len and values at the same time, orchestrate it
// differently.
func (lr *logRecorder) Len() int {
	lr.mutex.Lock()
	defer lr.mutex.Unlock()

	return len(lr.entries)
}

func (lr *logRecorder) checkEntries(t *testing.T, fun func(t *testing.T, e *log.Entry)) {
	lr.mutex.Lock()
	defer lr.mutex.Unlock()

	for _, e := range lr.entries {
		fun(t, e)
	}
}

func (lr *logRecorder) recordLogs(fm *FileMeta) error {
	r, err := log.NewReader(fm.Name)
	if err != nil {
		return err
	}
	defer r.Close()

	for {
		e, err := r.NextProto()
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}

		lr.mutex.Lock()
		lr.entries = append(lr.entries, e)
		lr.mutex.Unlock()
	}
}

func testProcessorLog(fm *FileMeta) error {
	r, err := log.NewReader(fm.Name)
	if err != nil {
		return err
	}
	defer r.Close()

	for {
		e, err := r.NextProto()
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}

		logrus.Infof("Log received: %+v", e)
	}
}

func makeClient(lp *LogPipe, dir string) (*log.Client, *logrus.Logger, string, error) {
	if dir == "" {
		var err error
		dir, err = ioutil.TempDir("", "")
		if err != nil {
			return nil, nil, dir, err
		}
	}

	config := &log.Config{
		TickInterval:       100 * time.Millisecond,
		BackoffCap:         100 * time.Millisecond,
		BackoffGranularity: time.Millisecond,
		DeliverLogs:        true,
		ShowOurLogs:        os.Getenv("DEBUG") != "",
	}

	kp, err := keypair.Generate()
	if err != nil {
		return nil, nil, dir, err
	}

	c, err := log.NewClient(lp.ListenAddr(), "test", dir, os.Getenv("DEBUG") != "", true, config, kp)
	l := logrus.New()
	l.AddHook(log.NewHook(c))

	if os.Getenv("DEBUG") == "" {
		l.SetOutput(ioutil.Discard)
	}

	return c, l, dir, err
}

func makeConfig(processor func(fm *FileMeta) error) (Config, *sql.DB, error) {
	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		return Config{}, nil, err
	}

	_, err = migrate.Exec(db, "postgres", migrations.Migrations, migrate.Up)
	if err != nil {
		return Config{}, nil, err
	}

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return Config{}, nil, err
	}

	logger := logrus.New()

	if os.Getenv("DEBUG") == "" {
		logger.SetOutput(ioutil.Discard)
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	return Config{
		Logger:        logger,
		KVMember:      "member",
		IndexName:     "test",
		MaxLogSize:    1024,
		SpoolDir:      dir,
		ListenAddress: ":0",
		Processor:     processor,
		RateLimiting:  false,
		RateLimitConfig: ratelimit.Config{
			Cap:             2048,
			RefreshInterval: time.Second,
			RefreshAmount:   128,
		},
	}, db, nil
}

func wipeTables(db *sql.DB) error {
	_, err := db.Exec("truncate table kvstore")
	return err
}

func readyServer(t *testing.T, config Config, db *sql.DB) (*LogPipe, chan error) {
	errChan := make(chan error, 1)
	ready := make(chan struct{})
	lp, err := NewLogPipe(config)
	assert.Nil(t, err)

	go func() {
		errChan <- lp.Boot(ready, db)
	}()

	<-ready
	return lp, errChan
}
