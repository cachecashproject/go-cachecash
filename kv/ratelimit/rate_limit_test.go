// +build external_test

package ratelimit

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cachecashproject/go-cachecash/dbtx"
	"github.com/cachecashproject/go-cachecash/kv"
	"github.com/cachecashproject/go-cachecash/kv/migrations"
	_ "github.com/lib/pq"
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

func doAssert(ctx context.Context, t *testing.T, rl *RateLimiter) {
	assert.Nil(t, rl.RateLimit(ctx, "my-key", 1337))
	assert.Nil(t, rl.RateLimit(ctx, "my-key", 1337))
	assert.Equal(t, rl.RateLimit(ctx, "my-key", 1337), ErrTooMuchData)

	time.Sleep(3 * time.Second)

	assert.Nil(t, rl.RateLimit(ctx, "my-key", 1337))
}

func TestRateLimitDB(t *testing.T) {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	db, err := createDBConn()
	assert.Nil(t, err)
	wipeTables(db)
	kv := kv.NewClient("ratelimiter", kv.NewDBDriver(l))
	ctx := dbtx.ContextWithExecutor(context.Background(), db)

	rl := NewRateLimiter(Config{
		Logger:          l,
		Cap:             3000,
		RefreshInterval: 1 * time.Second,
		RefreshAmount:   1000,
	}, kv)

	doAssert(ctx, t, rl)
}

func TestRateLimitMem(t *testing.T) {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	kv := kv.NewClient("ratelimiter", kv.NewMemoryDriver(nil))

	rl := NewRateLimiter(Config{
		Logger:          l,
		Cap:             3000,
		RefreshInterval: 1 * time.Second,
		RefreshAmount:   1000,
	}, kv)

	doAssert(ctx, t, rl)
}
