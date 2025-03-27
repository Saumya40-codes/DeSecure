package storage

import (
	"log"

	"github.com/dgraph-io/badger/v4"
)

type DB struct {
	conn *badger.DB
}

// OpenDB initializes BadgerDB
func OpenDB(path string) *DB {
	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal("Failed to open BadgerDB:", err)
	}
	return &DB{conn: db}
}

// CloseDB closes the database connection
func (db *DB) CloseDB() {
	db.conn.Close()
}

// Save sets a key-value pair in the database
func (db *DB) Save(key string, value []byte) error {
	return db.conn.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

// Load retrieves a value for a given key
func (db *DB) Load(key string) ([]byte, error) {
	var value []byte
	err := db.conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	return value, err
}
