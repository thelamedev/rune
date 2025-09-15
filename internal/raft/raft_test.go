package raft

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/require"
	"github.com/thelamedev/rune/internal/storage"
)

func newTestRaftNode(t *testing.T, bootstrap bool, nodeID, dataDir string) (*RaftNode, *storage.BoltStore) {
	t.Helper()

	storePath := filepath.Join(dataDir, "store.db")
	store, err := storage.NewBoltStore(storePath)
	require.NoError(t, err)

	err = store.Initialize(context.Background())
	require.NoError(t, err)

	fsm := NewFSM(store)
	cfg := &Config{
		NodeID:    nodeID,
		BindAddr:  "127.0.0.1:0",
		Bootstrap: bootstrap,
		DataDir:   dataDir,
	}

	node, err := NewRaftNode(cfg, fsm)
	require.NoError(t, err)

	t.Cleanup(func() {
		shutdownFuture := node.raft.Shutdown()
		require.NoError(t, shutdownFuture.Error(), "failed to shutdown raft node")
		require.NoError(t, store.Close(), "failed to close bolt store")
	})

	return node, store
}

func TestRaftNode_SingleNode_Apply(t *testing.T) {
	dataDir := t.TempDir()
	node, store := newTestRaftNode(t, true, "node-1", dataDir)

	// Wait for the node to become the leader.
	// In a single-node bootstrap cluster, this should happen almost instantly.
	require.Eventually(t, func() bool {
		return node.raft.State() == raft.Leader
	}, 3*time.Second, 100*time.Millisecond, "node never became leader")

	ctx := context.Background()

	// 1. Test a 'set' operation.
	t.Run("set operation", func(t *testing.T) {
		setCmd := command{
			Op:    "set",
			Key:   "hello",
			Value: []byte("world"),
		}
		cmdBytes, err := json.Marshal(setCmd)
		require.NoError(t, err)

		// Apply the command to the Raft log.
		applyFuture := node.raft.Apply(cmdBytes, 500*time.Millisecond)
		require.NoError(t, applyFuture.Error(), "failed to apply 'set' command")

		// Verify the value was written to the store.
		// We use `Eventually` because applying the log to the FSM is asynchronous.
		require.Eventually(t, func() bool {
			val, err := store.Get(ctx, "hello")
			if err != nil {
				return false
			}
			return string(val) == "world"
		}, 2*time.Second, 300*time.Millisecond, "value was not set in store")
	})

	// 2. Test a 'delete' operation.
	t.Run("delete operation", func(t *testing.T) {
		deleteCmd := command{
			Op:  "delete",
			Key: "hello",
		}
		cmdBytes, err := json.Marshal(deleteCmd)
		require.NoError(t, err)

		// Apply the delete command.
		applyFuture := node.raft.Apply(cmdBytes, 500*time.Millisecond)
		require.NoError(t, applyFuture.Error(), "failed to apply 'delete' command")

		// Verify the value was deleted from the store.
		require.Eventually(t, func() bool {
			_, err := store.Get(ctx, "hello")
			// We expect a "key not found" error from our BoltStore.
			return err != nil && strings.Contains(err.Error(), "key not found")
		}, 2*time.Second, 50*time.Millisecond, "value was not deleted from store")
	})
}
