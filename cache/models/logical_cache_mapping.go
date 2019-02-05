// Code generated by SQLBoiler (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cachecashproject/go-cachecash/common"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/strmangle"
)

// LogicalCacheMapping is an object representing the database table.
type LogicalCacheMapping struct {
	EscrowID      common.EscrowID `boil:"escrow_id" json:"escrow_id" toml:"escrow_id" yaml:"escrow_id"`
	SlotIdx       uint64          `boil:"slot_idx" json:"slot_idx" toml:"slot_idx" yaml:"slot_idx"`
	BlockEscrowID string          `boil:"block_escrow_id" json:"block_escrow_id" toml:"block_escrow_id" yaml:"block_escrow_id"`
	BlockID       common.BlockID  `boil:"block_id" json:"block_id" toml:"block_id" yaml:"block_id"`

	R *logicalCacheMappingR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L logicalCacheMappingL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var LogicalCacheMappingColumns = struct {
	EscrowID      string
	SlotIdx       string
	BlockEscrowID string
	BlockID       string
}{
	EscrowID:      "escrow_id",
	SlotIdx:       "slot_idx",
	BlockEscrowID: "block_escrow_id",
	BlockID:       "block_id",
}

// Generated where

type whereHelpercommon_EscrowID struct{ field string }

func (w whereHelpercommon_EscrowID) EQ(x common.EscrowID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpercommon_EscrowID) NEQ(x common.EscrowID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpercommon_EscrowID) LT(x common.EscrowID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpercommon_EscrowID) LTE(x common.EscrowID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpercommon_EscrowID) GT(x common.EscrowID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpercommon_EscrowID) GTE(x common.EscrowID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

type whereHelperuint64 struct{ field string }

func (w whereHelperuint64) EQ(x uint64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperuint64) NEQ(x uint64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperuint64) LT(x uint64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperuint64) LTE(x uint64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperuint64) GT(x uint64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperuint64) GTE(x uint64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }

type whereHelperstring struct{ field string }

func (w whereHelperstring) EQ(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperstring) NEQ(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperstring) LT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperstring) LTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperstring) GT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperstring) GTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }

type whereHelpercommon_BlockID struct{ field string }

func (w whereHelpercommon_BlockID) EQ(x common.BlockID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpercommon_BlockID) NEQ(x common.BlockID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpercommon_BlockID) LT(x common.BlockID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpercommon_BlockID) LTE(x common.BlockID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpercommon_BlockID) GT(x common.BlockID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpercommon_BlockID) GTE(x common.BlockID) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var LogicalCacheMappingWhere = struct {
	EscrowID      whereHelpercommon_EscrowID
	SlotIdx       whereHelperuint64
	BlockEscrowID whereHelperstring
	BlockID       whereHelpercommon_BlockID
}{
	EscrowID:      whereHelpercommon_EscrowID{field: `escrow_id`},
	SlotIdx:       whereHelperuint64{field: `slot_idx`},
	BlockEscrowID: whereHelperstring{field: `block_escrow_id`},
	BlockID:       whereHelpercommon_BlockID{field: `block_id`},
}

// LogicalCacheMappingRels is where relationship names are stored.
var LogicalCacheMappingRels = struct {
}{}

// logicalCacheMappingR is where relationships are stored.
type logicalCacheMappingR struct {
}

// NewStruct creates a new relationship struct
func (*logicalCacheMappingR) NewStruct() *logicalCacheMappingR {
	return &logicalCacheMappingR{}
}

// logicalCacheMappingL is where Load methods for each relationship are stored.
type logicalCacheMappingL struct{}

var (
	logicalCacheMappingColumns               = []string{"escrow_id", "slot_idx", "block_escrow_id", "block_id"}
	logicalCacheMappingColumnsWithoutDefault = []string{"escrow_id", "slot_idx", "block_escrow_id", "block_id"}
	logicalCacheMappingColumnsWithDefault    = []string{}
	logicalCacheMappingPrimaryKeyColumns     = []string{"escrow_id", "slot_idx"}
)

type (
	// LogicalCacheMappingSlice is an alias for a slice of pointers to LogicalCacheMapping.
	// This should generally be used opposed to []LogicalCacheMapping.
	LogicalCacheMappingSlice []*LogicalCacheMapping
	// LogicalCacheMappingHook is the signature for custom LogicalCacheMapping hook methods
	LogicalCacheMappingHook func(context.Context, boil.ContextExecutor, *LogicalCacheMapping) error

	logicalCacheMappingQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	logicalCacheMappingType                 = reflect.TypeOf(&LogicalCacheMapping{})
	logicalCacheMappingMapping              = queries.MakeStructMapping(logicalCacheMappingType)
	logicalCacheMappingPrimaryKeyMapping, _ = queries.BindMapping(logicalCacheMappingType, logicalCacheMappingMapping, logicalCacheMappingPrimaryKeyColumns)
	logicalCacheMappingInsertCacheMut       sync.RWMutex
	logicalCacheMappingInsertCache          = make(map[string]insertCache)
	logicalCacheMappingUpdateCacheMut       sync.RWMutex
	logicalCacheMappingUpdateCache          = make(map[string]updateCache)
	logicalCacheMappingUpsertCacheMut       sync.RWMutex
	logicalCacheMappingUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var logicalCacheMappingBeforeInsertHooks []LogicalCacheMappingHook
var logicalCacheMappingBeforeUpdateHooks []LogicalCacheMappingHook
var logicalCacheMappingBeforeDeleteHooks []LogicalCacheMappingHook
var logicalCacheMappingBeforeUpsertHooks []LogicalCacheMappingHook

var logicalCacheMappingAfterInsertHooks []LogicalCacheMappingHook
var logicalCacheMappingAfterSelectHooks []LogicalCacheMappingHook
var logicalCacheMappingAfterUpdateHooks []LogicalCacheMappingHook
var logicalCacheMappingAfterDeleteHooks []LogicalCacheMappingHook
var logicalCacheMappingAfterUpsertHooks []LogicalCacheMappingHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *LogicalCacheMapping) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range logicalCacheMappingBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *LogicalCacheMapping) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range logicalCacheMappingBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *LogicalCacheMapping) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range logicalCacheMappingBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *LogicalCacheMapping) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range logicalCacheMappingBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *LogicalCacheMapping) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range logicalCacheMappingAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *LogicalCacheMapping) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range logicalCacheMappingAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *LogicalCacheMapping) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range logicalCacheMappingAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *LogicalCacheMapping) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range logicalCacheMappingAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *LogicalCacheMapping) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range logicalCacheMappingAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddLogicalCacheMappingHook registers your hook function for all future operations.
func AddLogicalCacheMappingHook(hookPoint boil.HookPoint, logicalCacheMappingHook LogicalCacheMappingHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		logicalCacheMappingBeforeInsertHooks = append(logicalCacheMappingBeforeInsertHooks, logicalCacheMappingHook)
	case boil.BeforeUpdateHook:
		logicalCacheMappingBeforeUpdateHooks = append(logicalCacheMappingBeforeUpdateHooks, logicalCacheMappingHook)
	case boil.BeforeDeleteHook:
		logicalCacheMappingBeforeDeleteHooks = append(logicalCacheMappingBeforeDeleteHooks, logicalCacheMappingHook)
	case boil.BeforeUpsertHook:
		logicalCacheMappingBeforeUpsertHooks = append(logicalCacheMappingBeforeUpsertHooks, logicalCacheMappingHook)
	case boil.AfterInsertHook:
		logicalCacheMappingAfterInsertHooks = append(logicalCacheMappingAfterInsertHooks, logicalCacheMappingHook)
	case boil.AfterSelectHook:
		logicalCacheMappingAfterSelectHooks = append(logicalCacheMappingAfterSelectHooks, logicalCacheMappingHook)
	case boil.AfterUpdateHook:
		logicalCacheMappingAfterUpdateHooks = append(logicalCacheMappingAfterUpdateHooks, logicalCacheMappingHook)
	case boil.AfterDeleteHook:
		logicalCacheMappingAfterDeleteHooks = append(logicalCacheMappingAfterDeleteHooks, logicalCacheMappingHook)
	case boil.AfterUpsertHook:
		logicalCacheMappingAfterUpsertHooks = append(logicalCacheMappingAfterUpsertHooks, logicalCacheMappingHook)
	}
}

// One returns a single logicalCacheMapping record from the query.
func (q logicalCacheMappingQuery) One(ctx context.Context, exec boil.ContextExecutor) (*LogicalCacheMapping, error) {
	o := &LogicalCacheMapping{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for logical_cache_mapping")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all LogicalCacheMapping records from the query.
func (q logicalCacheMappingQuery) All(ctx context.Context, exec boil.ContextExecutor) (LogicalCacheMappingSlice, error) {
	var o []*LogicalCacheMapping

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to LogicalCacheMapping slice")
	}

	if len(logicalCacheMappingAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all LogicalCacheMapping records in the query.
func (q logicalCacheMappingQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count logical_cache_mapping rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q logicalCacheMappingQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if logical_cache_mapping exists")
	}

	return count > 0, nil
}

// LogicalCacheMappings retrieves all the records using an executor.
func LogicalCacheMappings(mods ...qm.QueryMod) logicalCacheMappingQuery {
	mods = append(mods, qm.From("\"logical_cache_mapping\""))
	return logicalCacheMappingQuery{NewQuery(mods...)}
}

// FindLogicalCacheMapping retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindLogicalCacheMapping(ctx context.Context, exec boil.ContextExecutor, escrowID common.EscrowID, slotIdx uint64, selectCols ...string) (*LogicalCacheMapping, error) {
	logicalCacheMappingObj := &LogicalCacheMapping{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"logical_cache_mapping\" where \"escrow_id\"=? AND \"slot_idx\"=?", sel,
	)

	q := queries.Raw(query, escrowID, slotIdx)

	err := q.Bind(ctx, exec, logicalCacheMappingObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from logical_cache_mapping")
	}

	return logicalCacheMappingObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *LogicalCacheMapping) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no logical_cache_mapping provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(logicalCacheMappingColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	logicalCacheMappingInsertCacheMut.RLock()
	cache, cached := logicalCacheMappingInsertCache[key]
	logicalCacheMappingInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			logicalCacheMappingColumns,
			logicalCacheMappingColumnsWithDefault,
			logicalCacheMappingColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(logicalCacheMappingType, logicalCacheMappingMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(logicalCacheMappingType, logicalCacheMappingMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"logical_cache_mapping\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"logical_cache_mapping\" () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT \"%s\" FROM \"logical_cache_mapping\" WHERE %s", strings.Join(returnColumns, "\",\""), strmangle.WhereClause("\"", "\"", 0, logicalCacheMappingPrimaryKeyColumns))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	_, err = exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into logical_cache_mapping")
	}

	var identifierCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.EscrowID,
		o.SlotIdx,
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.retQuery)
		fmt.Fprintln(boil.DebugWriter, identifierCols...)
	}

	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "models: unable to populate default values for logical_cache_mapping")
	}

CacheNoHooks:
	if !cached {
		logicalCacheMappingInsertCacheMut.Lock()
		logicalCacheMappingInsertCache[key] = cache
		logicalCacheMappingInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the LogicalCacheMapping.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *LogicalCacheMapping) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	logicalCacheMappingUpdateCacheMut.RLock()
	cache, cached := logicalCacheMappingUpdateCache[key]
	logicalCacheMappingUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			logicalCacheMappingColumns,
			logicalCacheMappingPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update logical_cache_mapping, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"logical_cache_mapping\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 0, wl),
			strmangle.WhereClause("\"", "\"", 0, logicalCacheMappingPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(logicalCacheMappingType, logicalCacheMappingMapping, append(wl, logicalCacheMappingPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update logical_cache_mapping row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for logical_cache_mapping")
	}

	if !cached {
		logicalCacheMappingUpdateCacheMut.Lock()
		logicalCacheMappingUpdateCache[key] = cache
		logicalCacheMappingUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q logicalCacheMappingQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for logical_cache_mapping")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for logical_cache_mapping")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o LogicalCacheMappingSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), logicalCacheMappingPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"logical_cache_mapping\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, logicalCacheMappingPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in logicalCacheMapping slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all logicalCacheMapping")
	}
	return rowsAff, nil
}

// Delete deletes a single LogicalCacheMapping record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *LogicalCacheMapping) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no LogicalCacheMapping provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), logicalCacheMappingPrimaryKeyMapping)
	sql := "DELETE FROM \"logical_cache_mapping\" WHERE \"escrow_id\"=? AND \"slot_idx\"=?"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from logical_cache_mapping")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for logical_cache_mapping")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q logicalCacheMappingQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no logicalCacheMappingQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from logical_cache_mapping")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for logical_cache_mapping")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o LogicalCacheMappingSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no LogicalCacheMapping slice provided for delete all")
	}

	if len(o) == 0 {
		return 0, nil
	}

	if len(logicalCacheMappingBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), logicalCacheMappingPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"logical_cache_mapping\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, logicalCacheMappingPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from logicalCacheMapping slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for logical_cache_mapping")
	}

	if len(logicalCacheMappingAfterDeleteHooks) != 0 {
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
func (o *LogicalCacheMapping) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindLogicalCacheMapping(ctx, exec, o.EscrowID, o.SlotIdx)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *LogicalCacheMappingSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := LogicalCacheMappingSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), logicalCacheMappingPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"logical_cache_mapping\".* FROM \"logical_cache_mapping\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, logicalCacheMappingPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in LogicalCacheMappingSlice")
	}

	*o = slice

	return nil
}

// LogicalCacheMappingExists checks if the LogicalCacheMapping row exists.
func LogicalCacheMappingExists(ctx context.Context, exec boil.ContextExecutor, escrowID common.EscrowID, slotIdx uint64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"logical_cache_mapping\" where \"escrow_id\"=? AND \"slot_idx\"=? limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, escrowID, slotIdx)
	}

	row := exec.QueryRowContext(ctx, sql, escrowID, slotIdx)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if logical_cache_mapping exists")
	}

	return exists, nil
}
