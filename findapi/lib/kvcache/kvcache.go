package kvcache

import (
	"errors"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/schollz/golock"
)

// Database manages file access through bolt.DB connection and a file lock
type Database struct {
	db    *bolt.DB
	glock *golock.Lock
}

// Open opens(or creates) bolt database file
func (self *Database) Open(db_file string) error {
	if nil != self.db {
		self.Close()
	}

	if !strings.HasSuffix(db_file, ".bolt") {
		db_file += ".bolt"
	}

	// first initiate lockfile
	lock_file := strings.Replace(db_file, ".db", ".lock", -1)
	self.glock = golock.New(
		golock.OptionSetName(lock_file),
		golock.OptionSetInterval(1*time.Millisecond),
		golock.OptionSetTimeout(60*time.Second),
	)

	err := self.glock.Lock()
	if err != nil {
		return err
	}

	db, err := bolt.Open(db_file, 0600, &bolt.Options{Timeout: 1 * time.Second})
	self.db = db
	return err
}

// Close close database connection and remove file lock
func (self *Database) Close() error {
	self.db.Close()
	return self.glock.Unlock()
}

// Get retrieves a key from a bucket.
// Decrypts the value using the supplied passphrase.
func (self *Database) Get(table, key string) (string, error) {
	if nil == self.db {
		return "", errors.New("Database not opened")
	}
	var result string
	var err error
	return result, self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if nil == b {
			return errors.New("Bucket does not exist")
		}
		result = string(b.Get([]byte(key)))
		return err
	})
}

// Set saves a key value to a bucket.
// Encrypts the value using the supplied passphrase.
func (self *Database) Set(table, key, value string) error {
	if nil == self.db {
		return errors.New("Database not opened")
	}

	return self.db.Update(func(tx *bolt.Tx) error {

		_, err := tx.CreateBucketIfNotExists([]byte(table))
		if nil != err {
			return err
		}

		b := tx.Bucket([]byte(table))
		if nil == b {
			return errors.New("Bucket does not exist")
		}

		return b.Put([]byte(key), []byte(value))
	})
}

// // Keys lists all keys with in a bucket
// func (self *Database) Keys(table string) ([]string, error) {
// 	var result []string
// 	if nil == self.db {
// 		return result, errors.New("Database not opened")
// 	}
// 	return result, self.db.View(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte(table))
// 		if nil == b {
// 			return errors.New("Bucket does not exist")
// 		}
// 		return b.ForEach(func(k, v []byte) error {
// 			result = append(result, string(k))
// 			return nil
// 		})
// 	})
// }

// // Remove deletes a key from a bucket
// func (self *Database) Remove(table string, key string, passphrase string) error {
// 	return self.db.Update(func(tx *bolt.Tx) error {
// 		bucket := tx.Bucket([]byte(table))
// 		if bucket == nil {
// 			return fmt.Errorf("Bucket %q not found!", table)
// 		}
//
// 		err := bucket.Delete([]byte(key))
// 		if err != nil {
// 			return fmt.Errorf("Could not delete key: %s", err)
// 		}
// 		return err
// 	})
// }
//
// // Tables returns list of buckets
// func (self *Database) Tables() ([]string, error) {
// 	var result []string
// 	if nil == self.db {
// 		return result, errors.New("Database not opened")
// 	}
// 	return result, self.db.View(func(tx *bolt.Tx) error {
// 		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
// 			result = append(result, string(name))
// 			return nil
// 		})
// 	})
// }

// OpenDb opens bolt file and returns Database
func OpenDb(db_file string) (*Database, error) {
	db := Database{}
	db.Open(db_file)
	return &db, db.Open(db_file)
}
