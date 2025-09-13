package seal

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	"github.com/hashicorp/vault/shamir"
)

var (
	ErrSealInitialized   = errors.New("seal is already initialized")
	ErrSealUninitialized = errors.New("seal is not initialized")
	ErrSealThresholdMet  = errors.New("unseal threshold has been met")
	ErrInvalidShare      = errors.New("provided share is not valid")
)

type Seal struct {
	mu sync.Mutex

	shares    int
	threshold int

	secret       []byte
	masterKey    []byte
	unsealShares [][]byte
}

func New(shares, threshold int) *Seal {
	return &Seal{
		shares:    shares,
		threshold: threshold,
	}
}

// GenerateKeys creates a new master key and splites it into the congfiigured number of Shamir shares. This should only be called once when initializing Rune. It returns the key shares as base64-encoded strings.
func (s *Seal) GenerateKeys(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.secret != nil {
		return nil, ErrSealInitialized
	}

	// Generate a 32-byte master key (AES-256)
	s.secret = make([]byte, 32)
	if _, err := rand.Read(s.secret); err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	// Split the key into shamir shares
	shares, err := shamir.Split(s.secret, s.shares, s.threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to split master key: %w", err)
	}

	encodedShares := make([]string, len(shares))
	for i, share := range shares {
		encodedShares[i] = base64.StdEncoding.EncodeToString(share)
	}

	return encodedShares, nil
}

// Unseal accepts a single base64-encoded key share. If the number of shares meets the threshold, it attempts to reconstruct the master key. It returns true if the vault is now unsealed, and the progress (n/threshold)
func (s *Seal) Unseal(ctx context.Context, share string) (bool, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// if s.secret != nil {
	// 	return true, s.threshold, ErrSealInitialized
	// }

	if s.masterKey != nil {
		return true, s.threshold, ErrSealThresholdMet
	}

	keyBytes, err := base64.StdEncoding.DecodeString(share)
	if err != nil {
		return false, len(s.unsealShares), ErrInvalidShare
	}

	s.unsealShares = append(s.unsealShares, keyBytes)
	progress := len(s.unsealShares)

	if progress < s.threshold {
		return false, progress, nil
	}

	masterKey, err := shamir.Combine(s.unsealShares)
	if err != nil {
		s.unsealShares = nil
		return false, 0, fmt.Errorf("%w: %v", ErrInvalidShare, err)
	}

	// For security reasons, clear shares from memory
	s.unsealShares = nil
	s.masterKey = masterKey

	return true, progress, nil
}

func (s *Seal) IsUnsealed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.masterKey != nil
}

func (s *Seal) MasterKey() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.masterKey == nil {
		return nil, ErrSealUninitialized
	}

	keyCopy := make([]byte, len(s.masterKey))
	copy(keyCopy, s.masterKey)

	return keyCopy, nil
}
