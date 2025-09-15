package raft

import (
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	"github.com/pkg/errors"

	raftboltdb "github.com/hashicorp/raft-boltdb"
)

type Config struct {
	NodeID    string
	BindAddr  string
	Bootstrap bool
	DataDir   string
}

type RaftNode struct {
	config *Config
	raft   *raft.Raft
}

func NewRaftNode(cfg *Config, fsm raft.FSM) (*RaftNode, error) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(cfg.NodeID)

	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return nil, errors.Wrap(err, "failed to create data directory")
	}

	addr, err := net.ResolveTCPAddr("tcp", cfg.BindAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve bind address")
	}

	transport, err := raft.NewTCPTransport(cfg.BindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transport")
	}

	snapshots, err := raft.NewFileSnapshotStore(cfg.DataDir, 2, os.Stderr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create snapshot store")
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft-log.db"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create log store")
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft-stable.db"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create stable store")
	}

	r, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create raft node")
	}

	if cfg.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(cfg.NodeID),
					Address: raft.ServerAddress(cfg.BindAddr),
				},
			},
		}

		bootstrapFuture := r.BootstrapCluster(configuration)
		if err := bootstrapFuture.Error(); err != nil {
			return nil, errors.Wrap(err, "failed to bootstrap cluster")
		}
	}

	return &RaftNode{
		config: cfg,
		raft:   r,
	}, nil
}
