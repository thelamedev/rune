package server

import (
	"context"
	"errors"
	"testing"

	apiv1 "github.com/thelamedev/rune/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- Mock Implementations ---

// mockStorer is a mock of the Storer interface.
type mockStorer struct {
	data    map[string][]byte
	putErr  error
	getErr  error
	listErr error
}

func (m *mockStorer) Get(ctx context.Context, key string) ([]byte, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	val, ok := m.data[key]
	if !ok {
		// Simulate a not-found error
		return nil, errors.New("not found")
	}
	return val, nil
}

func (m *mockStorer) Put(ctx context.Context, key string, value []byte) error {
	if m.putErr != nil {
		return m.putErr
	}
	if m.data == nil {
		m.data = make(map[string][]byte)
	}
	m.data[key] = value
	return nil
}

// mockSealer is a mock of the Sealer interface.
type mockSealer struct {
	unsealed bool
	key      []byte
	err      error
}

func (m *mockSealer) IsUnsealed() bool {
	return m.unsealed
}

func (m *mockSealer) MasterKey() ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.key, nil
}

// mockCryptoEngine is a mock of the CryptoEngine interface.
// It performs a fake "encryption" by prepending a string.
type mockCryptoEngine struct {
	encryptErr error
	decryptErr error
}

func (m *mockCryptoEngine) Encrypt(plaintext []byte) ([]byte, error) {
	if m.encryptErr != nil {
		return nil, m.encryptErr
	}
	return append([]byte("encrypted:"), plaintext...), nil
}

func (m *mockCryptoEngine) Decrypt(payload []byte) ([]byte, error) {
	if m.decryptErr != nil {
		return nil, m.decryptErr
	}
	prefix := []byte("encrypted:")
	if len(payload) < len(prefix) {
		return nil, errors.New("invalid payload")
	}
	return payload[len(prefix):], nil
}

// --- Test Cases ---

func TestGRPCServer_Put(t *testing.T) {
	ctx := context.Background()
	req := &apiv1.PutRequest{Path: "test/secret", Value: []byte("my-value")}

	t.Run("success", func(t *testing.T) {
		server := &GRPCServer{
			Config: &Config{
				Storage: &mockStorer{},
				Seal:    &mockSealer{unsealed: true},
				Crypto:  &mockCryptoEngine{},
			},
		}
		_, err := server.Put(ctx, req)
		if err != nil {
			t.Fatalf("Put() returned an unexpected error: %v", err)
		}
	})

	t.Run("failure when sealed", func(t *testing.T) {
		server := &GRPCServer{
			Config: &Config{
				Seal: &mockSealer{unsealed: false}, // Vault is sealed
			},
		}
		_, err := server.Put(ctx, req)
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.FailedPrecondition {
			t.Fatalf("expected FailedPrecondition, got: %v", err)
		}
	})

	t.Run("failure on encryption", func(t *testing.T) {
		server := &GRPCServer{
			Config: &Config{
				Seal:   &mockSealer{unsealed: true},
				Crypto: &mockCryptoEngine{encryptErr: errors.New("crypto boom")}, // Encryption fails
			},
		}
		_, err := server.Put(ctx, req)
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Fatalf("expected Internal error, got: %v", err)
		}
	})

	t.Run("failure on storage", func(t *testing.T) {
		server := &GRPCServer{
			Config: &Config{
				Seal:    &mockSealer{unsealed: true},
				Crypto:  &mockCryptoEngine{},
				Storage: &mockStorer{putErr: errors.New("db boom")}, // Storage fails
			},
		}
		_, err := server.Put(ctx, req)
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Fatalf("expected Internal error, got: %v", err)
		}
	})
}

func TestGRPCServer_Get(t *testing.T) {
	ctx := context.Background()
	path := "test/secret"
	value := []byte("my-value")
	encryptedValue := append([]byte("encrypted:"), value...)
	req := &apiv1.GetRequest{Path: path}

	t.Run("success", func(t *testing.T) {
		server := &GRPCServer{
			Config: &Config{
				Storage: &mockStorer{data: map[string][]byte{path: encryptedValue}},
				Seal:    &mockSealer{unsealed: true},
				Crypto:  &mockCryptoEngine{},
			},
		}
		res, err := server.Get(ctx, req)
		if err != nil {
			t.Fatalf("Get() returned an unexpected error: %v", err)
		}
		if string(res.Value) != string(value) {
			t.Errorf("expected value %q, got %q", string(value), string(res.Value))
		}
	})

	t.Run("failure when sealed", func(t *testing.T) {
		server := &GRPCServer{
			Config: &Config{
				Seal: &mockSealer{unsealed: false}, // Vault is sealed
			},
		}
		_, err := server.Get(ctx, req)
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.FailedPrecondition {
			t.Fatalf("expected FailedPrecondition, got: %v", err)
		}
	})

	t.Run("failure on not found", func(t *testing.T) {
		server := &GRPCServer{
			Config: &Config{
				Storage: &mockStorer{}, // Empty storage
				Seal:    &mockSealer{unsealed: true},
				Crypto:  &mockCryptoEngine{},
			},
		}
		_, err := server.Get(ctx, req)
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.NotFound {
			t.Fatalf("expected NotFound, got: %v", err)
		}
	})

	t.Run("failure on decryption", func(t *testing.T) {
		server := &GRPCServer{
			Config: &Config{
				Storage: &mockStorer{data: map[string][]byte{path: encryptedValue}},
				Seal:    &mockSealer{unsealed: true},
				Crypto:  &mockCryptoEngine{decryptErr: errors.New("crypto boom")}, // Decryption fails
			},
		}
		_, err := server.Get(ctx, req)
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Fatalf("expected Internal, got: %v", err)
		}
	})
}
