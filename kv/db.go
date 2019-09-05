package kv

import (
	"context"
	"crypto/rand"
	"database/sql"

	"github.com/cachecashproject/go-cachecash/kv/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
)

// DBDriver is a database driver.
type DBDriver struct {
	db         *sql.DB
	log        *logrus.Logger
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewDBDriver creates a new DBDriver from a db handle.
func NewDBDriver(db *sql.DB, log *logrus.Logger) Driver {
	ctx, cancel := context.WithCancel(context.Background())
	return &DBDriver{log: log, db: db, ctx: ctx, cancelFunc: cancel}
}

func randomBuf() ([]byte, error) {
	buf := make([]byte, 32)

	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}

	return buf, nil
}

// Close closes the database driver; not the handle
func (db *DBDriver) Close() {
	db.cancelFunc()
}

func (db *DBDriver) reapTx(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		db.log.Errorf("Rollback could not complete: %v", err)
	}
}

// tx free get
func (db *DBDriver) get(tx *sql.Tx, member, key string) ([]byte, []byte, error) {
	record, err := models.Kvstores(
		models.KvstoreWhere.Member.EQ(member),
		models.KvstoreWhere.Key.EQ(key),
	).One(db.ctx, tx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, ErrUnsetValue
		}

		return nil, nil, err
	}

	return []byte(record.Value), record.Nonce, nil
}

// Delete removes a key. If a nonce is provided, it will be checked.
func (db *DBDriver) Delete(member, key string, nonce []byte) error {
	tx, err := db.db.BeginTx(db.ctx, nil)
	if err != nil {
		return err
	}
	defer db.reapTx(tx)

	var kv models.Kvstore

	if nonce != nil {
		kv = models.Kvstore{
			Member: member,
			Key:    key,
			Nonce:  nonce,
		}
	} else {
		kv = models.Kvstore{
			Member: member,
			Key:    key,
		}
	}

	res, err := kv.Delete(db.ctx, tx)
	if err != nil {
		return err
	}

	if res == 0 {
		if nonce != nil {
			// return not equal here because we're not really sure if we're unset or not.
			// Instead of checking twice, this is safe and moderately sane.
			return ErrNotEqual
		}
		return ErrUnsetValue
	}

	return tx.Commit()
}

// Create creates a key from scratch.
func (db *DBDriver) Create(member, key string, value []byte) ([]byte, error) {
	buf, err := randomBuf()
	if err != nil {
		return nil, err
	}

	tx, err := db.db.BeginTx(db.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer db.reapTx(tx)

	kv := models.Kvstore{
		Member: member,
		Key:    key,
		Value:  value,
		Nonce:  buf,
	}

	if err := kv.Insert(db.ctx, tx, boil.Infer()); err != nil {
		return nil, errors.Wrap(ErrAlreadySet, err.Error())
	}

	return buf, tx.Commit()
}

// Get retrieves an item from the store. Users must pass a pointer to the out
// argument so it can be filled by json.Marshal.
func (db *DBDriver) Get(member, key string) ([]byte, []byte, error) {
	tx, err := db.db.BeginTx(db.ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer db.reapTx(tx)

	return db.get(tx, member, key)
}

// tx free set
func (db *DBDriver) set(tx *sql.Tx, member, key string, value []byte) ([]byte, error) {
	buf, err := randomBuf()
	if err != nil {
		return nil, err
	}

	kv := models.Kvstore{
		Member: member,
		Key:    key,
		Value:  value,
		Nonce:  buf,
	}

	return buf, kv.Upsert(db.ctx, tx, true, []string{"member", "key"}, boil.Whitelist("value"), boil.Infer())
}

// Set sets a value in the store
func (db *DBDriver) Set(member, key string, value []byte) ([]byte, error) {
	tx, err := db.db.BeginTx(db.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer db.reapTx(tx)

	buf, err := db.set(tx, member, key, value)
	if err != nil {
		return nil, err
	}

	return buf, tx.Commit()
}

// CAS implements a compare-and-swap operation.
func (db *DBDriver) CAS(member, key string, nonce, origValue, value []byte) ([]byte, error) {
	tx, err := db.db.BeginTx(db.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer db.reapTx(tx)

	buf, err := randomBuf()
	if err != nil {
		return nil, err
	}

	count, err := models.Kvstores(
		models.KvstoreWhere.Member.EQ(member),
		models.KvstoreWhere.Key.EQ(key),
		models.KvstoreWhere.Value.EQ(origValue),
		models.KvstoreWhere.Nonce.EQ(nonce),
	).UpdateAll(db.ctx, tx, models.M{"value": value, "nonce": buf})
	if err != nil {
		return nil, err
	}

	if count == 0 {
		if _, _, err := db.get(tx, member, key); err != nil {
			return nil, err
		}
		return nil, ErrNotEqual
	}

	return buf, tx.Commit()
}
