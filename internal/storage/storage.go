package storage

import (
	"context"
	"io"
)

type Storage interface {
	Initialize(ctx context.Context) error
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]string, error)
	Snapshot(w io.Writer) error
	Restore(r io.Reader) error
	Close() error
}
