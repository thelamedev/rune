package storage

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func newTestBoltStore(t *testing.T) *BoltStore {
	t.Helper()

	// dbPath := t.TempDir()
	dbPath := "./"
	dbPath = filepath.Join(dbPath, "test.db")

	store, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create bolt store: %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("failed to close bolt store: %v", err)
		}

		if err := os.Remove(dbPath); err != nil {
			t.Fatalf("failed to remove bolt db file: %v", err)
		}
	})

	return store
}

func TestBoltStore(t *testing.T) {
	store := newTestBoltStore(t)
	ctx := context.Background()

	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("failed to initialize bolt store: %v", err)
	}

	key := "my-test-key"
	value := []byte("my-test-password")

	t.Run("Put and Get", func(t *testing.T) {
		if err := store.Put(ctx, key, value); err != nil {
			t.Fatalf("failed to put value: %v", err)
		}

		got, err := store.Get(ctx, key)
		if err != nil {
			t.Fatalf("failed to get value: %v", err)
		}

		if !reflect.DeepEqual(got, value) {
			t.Fatalf("expected value %q, got %q", value, got)
		}
	})

	t.Run("Got non-existent key", func(t *testing.T) {
		_, err := store.Get(ctx, "non-existent-key")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("List", func(t *testing.T) {
		testKeys := map[string][]byte{
			"secrets/db/pass": []byte("my-test-password"),
			"secrets/db/user": []byte("my-test-user"),
			"secrets/api/key": []byte("abc"),
			"config/feature":  []byte("true"),
		}
		for k, v := range testKeys {
			if err := store.Put(ctx, k, v); err != nil {
				t.Fatalf("failed to put value: %v", err)
			}
		}

		expected := []string{"secrets/db/pass", "secrets/db/user"}
		keys, err := store.List(ctx, "secrets/db/")
		if err != nil {
			t.Fatalf("failed to list keys: %v", err)
		}

		if !reflect.DeepEqual(keys, expected) {
			t.Fatalf("expected keys %q, got %q", expected, keys)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := store.Delete(ctx, key); err != nil {
			t.Fatalf("failed to delete value: %v", err)
		}

		_, err := store.Get(ctx, key)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		err = store.Delete(ctx, key)
		if err != nil {
			t.Fatalf("deleting a non-existent key should not produce an error, but got: %v", err)
		}
	})
}
