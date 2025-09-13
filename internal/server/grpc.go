package server

import (
	"context"
	"errors"

	apiv1 "github.com/thelamedev/rune/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrStorageNotConfigured = errors.New("storage is not configured")
	ErrSealNotConfigured    = errors.New("seal is not configured")
	ErrCryptoNotConfigured  = errors.New("crypto is not configured")
)

type Storer interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, value []byte) error
}

type Sealer interface {
	IsUnsealed() bool
	MasterKey() ([]byte, error)
}

type CryptoEngine interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(payload []byte) ([]byte, error)
}

type Config struct {
	Storage Storer
	Seal    Sealer
	Crypto  CryptoEngine
}

type GRPCServer struct {
	apiv1.UnimplementedRuneServiceServer
	*Config
}

func NewGRPCServer(cfg *Config) (*grpc.Server, error) {
	gsrv := grpc.NewServer()
	srv, err := newRuneServiceServer(cfg)
	if err != nil {
		return nil, err
	}

	apiv1.RegisterRuneServiceServer(gsrv, srv)
	return gsrv, nil
}

func newRuneServiceServer(cfg *Config) (*GRPCServer, error) {
	if cfg.Storage == nil {
		return nil, ErrStorageNotConfigured
	}
	if cfg.Seal == nil {
		return nil, ErrSealNotConfigured
	}
	if cfg.Crypto == nil {
		return nil, ErrCryptoNotConfigured
	}

	return &GRPCServer{
		Config: cfg,
	}, nil
}

func (s *GRPCServer) Get(ctx context.Context, req *apiv1.GetRequest) (*apiv1.GetResponse, error) {
	if !s.Seal.IsUnsealed() {
		return nil, status.Error(codes.FailedPrecondition, "vault is sealed")
	}

	encryptedPayload, err := s.Storage.Get(ctx, req.Path)
	if err != nil {
		return nil, status.Error(codes.NotFound, "secret not found")
	}

	decryptedPayload, err := s.Crypto.Decrypt(encryptedPayload)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to decrypt secret")
	}

	return &apiv1.GetResponse{Value: decryptedPayload}, nil
}

func (s *GRPCServer) Put(ctx context.Context, req *apiv1.PutRequest) (*apiv1.PutResponse, error) {
	if !s.Seal.IsUnsealed() {
		return nil, status.Error(codes.FailedPrecondition, "vault is sealed")
	}

	encryptedPayload, err := s.Crypto.Encrypt(req.Value)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to encrypt secret")
	}

	err = s.Storage.Put(ctx, req.Path, encryptedPayload)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to store secret")
	}

	return &apiv1.PutResponse{Success: true}, nil
}
