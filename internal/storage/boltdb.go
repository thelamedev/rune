package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

var bucketName = []byte("rune_bucket")

type BoltStore struct {
	db     *bbolt.DB
	dbPath string
}

func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bbolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, err
	}
	return &BoltStore{db: db, dbPath: path}, nil
}

func (s *BoltStore) Initialize(ctx context.Context) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		return nil
	})
}

func (s *BoltStore) Get(ctx context.Context, key string) ([]byte, error) {
	var value []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		val := bucket.Get([]byte(key))
		if val == nil {
			return fmt.Errorf("key not found: %s", key)
		}

		value = make([]byte, len(val))
		copy(value, val)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *BoltStore) Put(ctx context.Context, key string, value []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf("failed to get bucket: %s", bucketName)
		}
		return bucket.Put([]byte(key), value)
	})
}

func (s *BoltStore) Delete(ctx context.Context, key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf("failed to get bucket: %s", bucketName)
		}
		return bucket.Delete([]byte(key))
	})
}

func (s *BoltStore) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		c := bucket.Cursor()
		prefixBytes := []byte(prefix)

		for k, _ := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, _ = c.Next() {
			keyCopy := make([]byte, len(k))
			copy(keyCopy, k)
			keys = append(keys, string(keyCopy))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	return keys, nil
}

func (s *BoltStore) Close() error {
	return s.db.Close()
}

func (b *BoltStore) Snapshot(w io.Writer) error {
	return b.db.View(func(tx *bbolt.Tx) error {
		// Set a large timeout for the snapshot to complete.
		_, err := tx.WriteTo(w)
		return err
	})
}

func (b *BoltStore) Restore(r io.Reader) error {
	if err := b.db.Close(); err != nil {
		return errors.Wrap(err, "failed to close database before restoring")
	}

	snapshotData, err := io.ReadAll(r)
	if err != nil {
		return errors.Wrap(err, "failed to read snapshot data")
	}

	if err := os.WriteFile(b.dbPath, snapshotData, 0o600); err != nil {
		return errors.Wrap(err, "failed to write snapshot data to database")
	}

	db, err := bbolt.Open(b.dbPath, 0o600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return errors.Wrap(err, "failed to open database after restoring")
	}

	b.db = db
	return b.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		return nil
	})
}
