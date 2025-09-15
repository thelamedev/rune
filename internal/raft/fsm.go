package raft

import (
	"context"
	"encoding/json"
	"io"

	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"github.com/thelamedev/rune/internal/storage"
)

type command struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value []byte `json:"value,omitempty"`
}

type fsm struct {
	store *storage.BoltStore
}

func NewFSM(store *storage.BoltStore) raft.FSM {
	return &fsm{
		store: store,
	}
}

func (f *fsm) Apply(log *raft.Log) any {
	var cmd command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		panic(errors.Wrap(err, "failed to unmarshal command"))
	}

	ctx := context.Background()

	switch cmd.Op {
	case "set":
		return f.store.Put(ctx, cmd.Key, cmd.Value)
	case "delete":
		return f.store.Delete(ctx, cmd.Key)
	default:
		panic(errors.Errorf("unrecognized command op: %s", cmd.Op))
	}
}

type fsmSnapshot struct {
	store *storage.BoltStore
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return &fsmSnapshot{store: f.store}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	return f.store.Restore(rc)
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		if err := f.store.Snapshot(sink); err != nil {
			return errors.Wrap(err, "failed to snapshot store")
		}

		return nil
	}()
	if err != nil {
		if err := sink.Cancel(); err != nil {
			return errors.Wrap(err, "failed to cancel snapshot")
		}
	}

	return err
}

func (f *fsmSnapshot) Release() {}
