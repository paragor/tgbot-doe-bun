package badgerdb

import (
	"context"
	"errors"
	"github.com/dgraph-io/badger/v3"
)

type badgerDb struct {
	db *badger.DB
}

func NewBadgerDb(path string) (*badgerDb, func(), error) {
	db, err := badger.Open(badger.DefaultOptions(path).WithSyncWrites(true))
	if err != nil {
		return nil, nil, err
	}
	return &badgerDb{db: db}, func() { db.Close() }, nil
}

func (db *badgerDb) KeysWithPrefix(ctx context.Context, prefix string) ([]string, error) {
	keys := []string{}
	err := db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(prefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			keys = append(keys, string(item.Key()))
		}
		return nil

	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (db *badgerDb) Get(ctx context.Context, key string) (string, error) {
	result := ""
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil
			}
			return err
		}
		return item.Value(func(val []byte) error {
			result = string(val)
			return nil
		})
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

func (db *badgerDb) Set(ctx context.Context, key, value string) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
}

func (db *badgerDb) Delete(ctx context.Context, key string) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}
