package dbtx

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

// ContextWithExecutor injects a ContextExecutor into a context.
func ContextWithExecutor(ctx context.Context, db boil.ContextExecutor) context.Context {
	return context.WithValue(ctx, contextKey, db)
}

// BeginTx begins a transaction from the DB stored in ctx, or errors if either
// there is no DB, or a transaction is already in progress.
func BeginTx(ctx context.Context) (context.Context, *sql.Tx, error) {
	db, ok := ctx.Value(contextKey).(boil.ContextBeginner)
	if !ok {
		// Provide useful diagnostics
		_, ok := ctx.Value(contextKey).(boil.ContextTransactor)
		if ok {
			// IF we decide to support these, we can do so by storing the DB in
			// a second key only used to look up when starting transactions
			return nil, nil, errors.New("Attempt to start a nested transaction")
		}
		return nil, nil, errors.New("No database available")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to begin transaction")
	}
	return ContextWithExecutor(ctx, tx), tx, nil
}

// ExecutorFromContext retrieves the ContextExecutor from a context. If there is none set, nil is returned - boil will
// panic if that happens, but since this is a programming error (a failure to set a ContextExecutor on a
// BackgroundContext), it is a reasonable tradeoff given the frequency with which ExecutorFromContext is used.
func ExecutorFromContext(ctx context.Context) boil.ContextExecutor {
	db, _ := ctx.Value(contextKey).(boil.ContextExecutor)
	return db
}
