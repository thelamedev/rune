package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func newTestBoltStore(t *testing.T) *BoltStore {
	t.Helper()

	// dbPath := t.TempDir()
	dbPath := "."
	dbPath = filepath.Join(dbPath, "test.db")

	store, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create bolt store: %v", err)
	}

	if err := store.Initialize(t.Context()); err != nil {
		t.Fatalf("failed to initialize bolt store: %v", err)
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
	ctx := t.Context()

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

func TestBoltStore_Snapshot(t *testing.T) {
	ctx := t.Context()
	store := newTestBoltStore(t)

	// 1. Populate the store with some data.
	keysToSet := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
	}
	for k, v := range keysToSet {
		if err := store.Put(ctx, k, v); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	// 2. Create a snapshot into a buffer.
	var buf bytes.Buffer
	if err := store.Snapshot(&buf); err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("Snapshot produced an empty file")
	}

	// 3. Restore the snapshot into a new database to verify it.
	restoreDir := "."
	restorePath := filepath.Join(restoreDir, "restore.db")

	if err := os.WriteFile(restorePath, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("failed to write snapshot to restore path: %v", err)
	}

	restoredStore, err := NewBoltStore(restorePath)
	if err != nil {
		t.Fatalf("failed to create new bolt store from restored snapshot: %v", err)
	}

	if err := restoredStore.Initialize(t.Context()); err != nil {
		t.Fatalf("failed to initialize bolt store: %v", err)
	}

	t.Cleanup(func() {
		if err := restoredStore.Close(); err != nil {
			t.Fatalf("failed to close restored store: %v", err)
		}
	})

	// 4. Verify the data in the restored database.
	for k, v := range keysToSet {
		retrievedValue, err := restoredStore.Get(ctx, k)
		if err != nil {
			t.Fatalf("Get from restored store failed for key %q: %v", k, err)
		}
		if !reflect.DeepEqual(v, retrievedValue) {
			t.Errorf("Get from restored store returned incorrect value for key %q: got %q, want %q", k, retrievedValue, v)
		}
	}
}
