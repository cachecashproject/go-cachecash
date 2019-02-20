// Code generated by SQLBoiler (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/strmangle"
	"golang.org/x/crypto/ed25519"
)

// Escrow is an object representing the database table.
type Escrow struct {
	ID         int                `boil:"id" json:"id" toml:"id" yaml:"id"`
	StartBlock int                `boil:"start_block" json:"start_block" toml:"start_block" yaml:"start_block"`
	EndBlock   int                `boil:"end_block" json:"end_block" toml:"end_block" yaml:"end_block"`
	State      string             `boil:"state" json:"state" toml:"state" yaml:"state"`
	PublicKey  ed25519.PublicKey  `boil:"public_key" json:"public_key" toml:"public_key" yaml:"public_key"`
	PrivateKey ed25519.PrivateKey `boil:"private_key" json:"private_key" toml:"private_key" yaml:"private_key"`
	Raw        []byte             `boil:"raw" json:"raw" toml:"raw" yaml:"raw"`

	R *escrowR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L escrowL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var EscrowColumns = struct {
	ID         string
	StartBlock string
	EndBlock   string
	State      string
	PublicKey  string
	PrivateKey string
	Raw        string
}{
	ID:         "id",
	StartBlock: "start_block",
	EndBlock:   "end_block",
	State:      "state",
	PublicKey:  "public_key",
	PrivateKey: "private_key",
	Raw:        "raw",
}

// Generated where

type whereHelpered25519_PrivateKey struct{ field string }

func (w whereHelpered25519_PrivateKey) EQ(x ed25519.PrivateKey) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpered25519_PrivateKey) NEQ(x ed25519.PrivateKey) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpered25519_PrivateKey) LT(x ed25519.PrivateKey) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpered25519_PrivateKey) LTE(x ed25519.PrivateKey) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpered25519_PrivateKey) GT(x ed25519.PrivateKey) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpered25519_PrivateKey) GTE(x ed25519.PrivateKey) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var EscrowWhere = struct {
	ID         whereHelperint
	StartBlock whereHelperint
	EndBlock   whereHelperint
	State      whereHelperstring
	PublicKey  whereHelpered25519_PublicKey
	PrivateKey whereHelpered25519_PrivateKey
	Raw        whereHelper__byte
}{
	ID:         whereHelperint{field: `id`},
	StartBlock: whereHelperint{field: `start_block`},
	EndBlock:   whereHelperint{field: `end_block`},
	State:      whereHelperstring{field: `state`},
	PublicKey:  whereHelpered25519_PublicKey{field: `public_key`},
	PrivateKey: whereHelpered25519_PrivateKey{field: `private_key`},
	Raw:        whereHelper__byte{field: `raw`},
}

// EscrowRels is where relationship names are stored.
var EscrowRels = struct {
	Bundles      string
	EscrowCaches string
}{
	Bundles:      "Bundles",
	EscrowCaches: "EscrowCaches",
}

// escrowR is where relationships are stored.
type escrowR struct {
	Bundles      BundleSlice
	EscrowCaches EscrowCacheSlice
}

// NewStruct creates a new relationship struct
func (*escrowR) NewStruct() *escrowR {
	return &escrowR{}
}

// escrowL is where Load methods for each relationship are stored.
type escrowL struct{}

var (
	escrowColumns               = []string{"id", "start_block", "end_block", "state", "public_key", "private_key", "raw"}
	escrowColumnsWithoutDefault = []string{"state", "public_key", "private_key", "raw"}
	escrowColumnsWithDefault    = []string{"id", "start_block", "end_block"}
	escrowPrimaryKeyColumns     = []string{"id"}
)

type (
	// EscrowSlice is an alias for a slice of pointers to Escrow.
	// This should generally be used opposed to []Escrow.
	EscrowSlice []*Escrow
	// EscrowHook is the signature for custom Escrow hook methods
	EscrowHook func(context.Context, boil.ContextExecutor, *Escrow) error

	escrowQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	escrowType                 = reflect.TypeOf(&Escrow{})
	escrowMapping              = queries.MakeStructMapping(escrowType)
	escrowPrimaryKeyMapping, _ = queries.BindMapping(escrowType, escrowMapping, escrowPrimaryKeyColumns)
	escrowInsertCacheMut       sync.RWMutex
	escrowInsertCache          = make(map[string]insertCache)
	escrowUpdateCacheMut       sync.RWMutex
	escrowUpdateCache          = make(map[string]updateCache)
	escrowUpsertCacheMut       sync.RWMutex
	escrowUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var escrowBeforeInsertHooks []EscrowHook
var escrowBeforeUpdateHooks []EscrowHook
var escrowBeforeDeleteHooks []EscrowHook
var escrowBeforeUpsertHooks []EscrowHook

var escrowAfterInsertHooks []EscrowHook
var escrowAfterSelectHooks []EscrowHook
var escrowAfterUpdateHooks []EscrowHook
var escrowAfterDeleteHooks []EscrowHook
var escrowAfterUpsertHooks []EscrowHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Escrow) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range escrowBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Escrow) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range escrowBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Escrow) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range escrowBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Escrow) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range escrowBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Escrow) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range escrowAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Escrow) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range escrowAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Escrow) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range escrowAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Escrow) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range escrowAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Escrow) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range escrowAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddEscrowHook registers your hook function for all future operations.
func AddEscrowHook(hookPoint boil.HookPoint, escrowHook EscrowHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		escrowBeforeInsertHooks = append(escrowBeforeInsertHooks, escrowHook)
	case boil.BeforeUpdateHook:
		escrowBeforeUpdateHooks = append(escrowBeforeUpdateHooks, escrowHook)
	case boil.BeforeDeleteHook:
		escrowBeforeDeleteHooks = append(escrowBeforeDeleteHooks, escrowHook)
	case boil.BeforeUpsertHook:
		escrowBeforeUpsertHooks = append(escrowBeforeUpsertHooks, escrowHook)
	case boil.AfterInsertHook:
		escrowAfterInsertHooks = append(escrowAfterInsertHooks, escrowHook)
	case boil.AfterSelectHook:
		escrowAfterSelectHooks = append(escrowAfterSelectHooks, escrowHook)
	case boil.AfterUpdateHook:
		escrowAfterUpdateHooks = append(escrowAfterUpdateHooks, escrowHook)
	case boil.AfterDeleteHook:
		escrowAfterDeleteHooks = append(escrowAfterDeleteHooks, escrowHook)
	case boil.AfterUpsertHook:
		escrowAfterUpsertHooks = append(escrowAfterUpsertHooks, escrowHook)
	}
}

// One returns a single escrow record from the query.
func (q escrowQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Escrow, error) {
	o := &Escrow{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for escrow")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Escrow records from the query.
func (q escrowQuery) All(ctx context.Context, exec boil.ContextExecutor) (EscrowSlice, error) {
	var o []*Escrow

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Escrow slice")
	}

	if len(escrowAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Escrow records in the query.
func (q escrowQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count escrow rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q escrowQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if escrow exists")
	}

	return count > 0, nil
}

// Bundles retrieves all the bundle's Bundles with an executor.
func (o *Escrow) Bundles(mods ...qm.QueryMod) bundleQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"bundle\".\"escrow_id\"=?", o.ID),
	)

	query := Bundles(queryMods...)
	queries.SetFrom(query.Query, "\"bundle\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"bundle\".*"})
	}

	return query
}

// EscrowCaches retrieves all the escrow_cach's EscrowCaches with an executor.
func (o *Escrow) EscrowCaches(mods ...qm.QueryMod) escrowCacheQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"escrow_caches\".\"escrow_id\"=?", o.ID),
	)

	query := EscrowCaches(queryMods...)
	queries.SetFrom(query.Query, "\"escrow_caches\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"escrow_caches\".*"})
	}

	return query
}

// LoadBundles allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (escrowL) LoadBundles(ctx context.Context, e boil.ContextExecutor, singular bool, maybeEscrow interface{}, mods queries.Applicator) error {
	var slice []*Escrow
	var object *Escrow

	if singular {
		object = maybeEscrow.(*Escrow)
	} else {
		slice = *maybeEscrow.(*[]*Escrow)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &escrowR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &escrowR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(qm.From(`bundle`), qm.WhereIn(`escrow_id in ?`, args...))
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load bundle")
	}

	var resultSlice []*Bundle
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice bundle")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on bundle")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for bundle")
	}

	if len(bundleAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.Bundles = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &bundleR{}
			}
			foreign.R.Escrow = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.EscrowID {
				local.R.Bundles = append(local.R.Bundles, foreign)
				if foreign.R == nil {
					foreign.R = &bundleR{}
				}
				foreign.R.Escrow = local
				break
			}
		}
	}

	return nil
}

// LoadEscrowCaches allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (escrowL) LoadEscrowCaches(ctx context.Context, e boil.ContextExecutor, singular bool, maybeEscrow interface{}, mods queries.Applicator) error {
	var slice []*Escrow
	var object *Escrow

	if singular {
		object = maybeEscrow.(*Escrow)
	} else {
		slice = *maybeEscrow.(*[]*Escrow)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &escrowR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &escrowR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(qm.From(`escrow_caches`), qm.WhereIn(`escrow_id in ?`, args...))
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load escrow_caches")
	}

	var resultSlice []*EscrowCache
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice escrow_caches")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on escrow_caches")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for escrow_caches")
	}

	if len(escrowCacheAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.EscrowCaches = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &escrowCacheR{}
			}
			foreign.R.Escrow = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.EscrowID {
				local.R.EscrowCaches = append(local.R.EscrowCaches, foreign)
				if foreign.R == nil {
					foreign.R = &escrowCacheR{}
				}
				foreign.R.Escrow = local
				break
			}
		}
	}

	return nil
}

// AddBundles adds the given related objects to the existing relationships
// of the escrow, optionally inserting them as new records.
// Appends related to o.R.Bundles.
// Sets related.R.Escrow appropriately.
func (o *Escrow) AddBundles(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*Bundle) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.EscrowID = o.ID
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"bundle\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"escrow_id"}),
				strmangle.WhereClause("\"", "\"", 2, bundlePrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}

			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.EscrowID = o.ID
		}
	}

	if o.R == nil {
		o.R = &escrowR{
			Bundles: related,
		}
	} else {
		o.R.Bundles = append(o.R.Bundles, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &bundleR{
				Escrow: o,
			}
		} else {
			rel.R.Escrow = o
		}
	}
	return nil
}

// AddEscrowCaches adds the given related objects to the existing relationships
// of the escrow, optionally inserting them as new records.
// Appends related to o.R.EscrowCaches.
// Sets related.R.Escrow appropriately.
func (o *Escrow) AddEscrowCaches(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*EscrowCache) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.EscrowID = o.ID
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"escrow_caches\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"escrow_id"}),
				strmangle.WhereClause("\"", "\"", 2, escrowCachePrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}

			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.EscrowID = o.ID
		}
	}

	if o.R == nil {
		o.R = &escrowR{
			EscrowCaches: related,
		}
	} else {
		o.R.EscrowCaches = append(o.R.EscrowCaches, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &escrowCacheR{
				Escrow: o,
			}
		} else {
			rel.R.Escrow = o
		}
	}
	return nil
}

// Escrows retrieves all the records using an executor.
func Escrows(mods ...qm.QueryMod) escrowQuery {
	mods = append(mods, qm.From("\"escrow\""))
	return escrowQuery{NewQuery(mods...)}
}

// FindEscrow retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindEscrow(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*Escrow, error) {
	escrowObj := &Escrow{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"escrow\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, escrowObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from escrow")
	}

	return escrowObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Escrow) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no escrow provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(escrowColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	escrowInsertCacheMut.RLock()
	cache, cached := escrowInsertCache[key]
	escrowInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			escrowColumns,
			escrowColumnsWithDefault,
			escrowColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(escrowType, escrowMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(escrowType, escrowMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"escrow\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"escrow\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into escrow")
	}

	if !cached {
		escrowInsertCacheMut.Lock()
		escrowInsertCache[key] = cache
		escrowInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Escrow.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Escrow) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	escrowUpdateCacheMut.RLock()
	cache, cached := escrowUpdateCache[key]
	escrowUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			escrowColumns,
			escrowPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update escrow, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"escrow\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, escrowPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(escrowType, escrowMapping, append(wl, escrowPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, values)
	}

	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update escrow row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for escrow")
	}

	if !cached {
		escrowUpdateCacheMut.Lock()
		escrowUpdateCache[key] = cache
		escrowUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q escrowQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for escrow")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for escrow")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o EscrowSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), escrowPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"escrow\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, escrowPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in escrow slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all escrow")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Escrow) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no escrow provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(escrowColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	escrowUpsertCacheMut.RLock()
	cache, cached := escrowUpsertCache[key]
	escrowUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			escrowColumns,
			escrowColumnsWithDefault,
			escrowColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			escrowColumns,
			escrowPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert escrow, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(escrowPrimaryKeyColumns))
			copy(conflict, escrowPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"escrow\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(escrowType, escrowMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(escrowType, escrowMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert escrow")
	}

	if !cached {
		escrowUpsertCacheMut.Lock()
		escrowUpsertCache[key] = cache
		escrowUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Escrow record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Escrow) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Escrow provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), escrowPrimaryKeyMapping)
	sql := "DELETE FROM \"escrow\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from escrow")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for escrow")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q escrowQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no escrowQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from escrow")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for escrow")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o EscrowSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Escrow slice provided for delete all")
	}

	if len(o) == 0 {
		return 0, nil
	}

	if len(escrowBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), escrowPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"escrow\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, escrowPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from escrow slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for escrow")
	}

	if len(escrowAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *Escrow) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindEscrow(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *EscrowSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := EscrowSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), escrowPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"escrow\".* FROM \"escrow\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, escrowPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in EscrowSlice")
	}

	*o = slice

	return nil
}

// EscrowExists checks if the Escrow row exists.
func EscrowExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"escrow\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}

	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if escrow exists")
	}

	return exists, nil
}
