// +build external_test

package kv

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cachecashproject/go-cachecash/dbtx"
	"github.com/cachecashproject/go-cachecash/kv/migrations"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
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

func wipeTables(db *sql.DB) error {
	_, err := db.Exec("truncate table kvstore")
	return err
}

func createDBConn() (*sql.DB, error) {
	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		return nil, err
	}

	_, err = migrate.Exec(db, "postgres", migrations.Migrations, migrate.Up)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Test01DBConn(t *testing.T) {
	db, err := createDBConn()
	assert.Nil(t, err)
	assert.Nil(t, db.Close())
}

func TestDBCreateDelete(t *testing.T) {
	db, err := createDBConn()
	assert.Nil(t, err)
	wipeTables(db)
	defer db.Close()

	c := NewClient("basic", NewDBDriver(logrus.New()))

	ctx := dbtx.ContextWithExecutor(context.Background(), db)

	nonce, err := c.CreateUint64(ctx, "uint", 0)
	assert.Nil(t, err)

	_, err = c.CreateUint64(ctx, "uint", 0)
	assert.Equal(t, errors.Cause(err), ErrAlreadySet)

	assert.Nil(t, c.Delete(ctx, "uint", nonce))
	assert.Equal(t, c.Delete(ctx, "uint", nonce), ErrNotEqual)
	assert.Equal(t, c.Delete(ctx, "uint", nil), ErrUnsetValue)

	_, err = c.CreateUint64(ctx, "uint", 0)
	assert.Nil(t, err)
	assert.Nil(t, c.Delete(ctx, "uint", nil))
}

func TestDBBasic(t *testing.T) {
	db, err := createDBConn()
	assert.NoError(t, err)
	if err != nil {
		return // no point doing anything without a data base
	}
	defer db.Close()
	err = wipeTables(db)
	assert.NoError(t, err)
	if err != nil {
		return // no point doing anything in a dirty db
	}
	ctx := dbtx.ContextWithExecutor(context.Background(), db)

	c := NewClient("basic", NewDBDriver(logrus.New()))
	basicTest(ctx, t, c)
}

type nonceTrack struct {
	nonceMutex sync.RWMutex
	nonce      []byte
}

func TestDBCASConcurrentCounter(t *testing.T) {
	db, err := createDBConn()
	assert.NoError(t, err)
	if err != nil {
		return // no point doing anything without a data base
	}
	defer db.Close()
	err = wipeTables(db)
	assert.NoError(t, err)
	if err != nil {
		return // no point doing anything in a dirty db
	}
	ctx := dbtx.ContextWithExecutor(context.Background(), db)
	c := NewClient("basic", NewDBDriver(logrus.New()))

	var routines uint64 = 100 // type matters here for the checks at the bottom
	done := make(chan struct{}, routines)
	timeout := time.Minute

	nonce, err := c.CreateUint64(ctx, "check", 0)
	assert.Nil(t, err)

	st := &nonceTrack{nonce: nonce}

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		time.Sleep(timeout)
		cancel()
	}()

	ready := make(chan struct{})

	// test plan: everything should finish within the timeout (typical runtime is
	// about 2s on my machine) and set its number properly. At the end, the
	// number should be the routine count, and the count of iterations on the
	// channel that's signaled when the goroutine is closing should also equal
	// the routine count.
	for i := uint64(0); i < routines; i++ {
		go func(i uint64, st *nonceTrack) {
			<-ready
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				st.nonceMutex.RLock()
				s := st.nonce
				st.nonceMutex.RUnlock()

				s2, err := c.CASUint64(ctx, "check", s, i, i+1)
				if err != nil {
					continue
				}

				st.nonceMutex.Lock()
				st.nonce = s2
				st.nonceMutex.Unlock()

				done <- struct{}{}
				return
			}
		}(i, st)
	}

	close(ready)
	var i uint64

	for i = 0; i < routines; i++ {
		select {
		case <-ctx.Done():
			goto end
		case <-done:
		}
	}

end:
	out, _, err := c.GetUint64(ctx, "check")
	assert.Nil(t, err)
	assert.Equal(t, out, routines)
	assert.Equal(t, routines, i)
}

func TestDBMember(t *testing.T) {
	db, err := createDBConn()
	assert.Nil(t, err)
	wipeTables(db)
	defer db.Close()
	ctx := dbtx.ContextWithExecutor(context.Background(), db)

	c := NewClient("member1", NewDBDriver(logrus.New()))

	_, _, err = c.GetUint64(ctx, "one")
	assert.Equal(t, err, ErrUnsetValue)
	_, err = c.SetUint64(ctx, "one", 1)
	assert.Nil(t, err)

	out, _, err := c.GetUint64(ctx, "one")
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), out)

	c2 := NewClient("member2", NewDBDriver(logrus.New()))
	_, _, err = c2.GetUint64(ctx, "one")
	assert.Equal(t, err, ErrUnsetValue)
	_, err = c2.SetUint64(ctx, "one", 2)
	assert.Nil(t, err)

	out, _, err = c2.GetUint64(ctx, "one")
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), out)

	out, _, err = c.GetUint64(ctx, "one")
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), out)
}
