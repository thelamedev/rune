package seal

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
)

func TestSeal_GenerateKeys(t *testing.T) {
	ctx := context.Background()
	shares, threshold := 5, 3
	s := New(shares, threshold)

	generatedShares, err := s.GenerateKeys(ctx)
	if err != nil {
		t.Fatalf("expected no error from GenerateKeys, but got %v", err)
	}

	if len(generatedShares) != shares {
		t.Errorf("expected %d shares, but got %d", shares, len(generatedShares))
	}

	for _, share := range generatedShares {
		if _, err := base64.StdEncoding.DecodeString(share); err != nil {
			t.Errorf("expected share to be valid base64, but got error: %v", err)
		}
	}

	// Calling a second time should fail
	_, err = s.GenerateKeys(ctx)
	if !errors.Is(err, ErrSealInitialized) {
		t.Errorf("expected ErrSealInitialized on second call, but got %v", err)
	}
}

func TestSeal_Lifecycle(t *testing.T) {
	shares, threshold := 5, 3
	s := New(shares, threshold)
	ctx := context.Background()

	// Should be sealed initially
	if s.IsUnsealed() {
		t.Fatal("seal should be sealed initially")
	}

	// Getting master key should fail when sealed
	_, err := s.MasterKey()
	if !errors.Is(err, ErrSealUninitialized) {
		t.Fatalf("expected ErrSealUninitialized when getting master key while sealed, got %v", err)
	}

	// Generate keys
	keyShares, err := s.GenerateKeys(ctx)
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	// Unseal step-by-step
	for i := 0; i < threshold; i++ {
		isUnsealed, progress, err := s.Unseal(ctx, keyShares[i])
		if err != nil {
			t.Fatalf("failed during unseal: %v", err)
		}
		if i < threshold-1 {
			if isUnsealed {
				t.Fatal("should not be unsealed before threshold is met")
			}
			if progress != i+1 {
				t.Errorf("expected progress %d, got %d", i+1, progress)
			}
		} else {
			if !isUnsealed {
				t.Fatal("should be unsealed when threshold is met")
			}
		}
	}

	// Should be unsealed now
	if !s.IsUnsealed() {
		t.Fatal("seal should be unsealed after providing enough shares")
	}

	// Getting master key should now succeed
	masterKey, err := s.MasterKey()
	if err != nil {
		t.Fatalf("expected no error from MasterKey after unsealing, got %v", err)
	}
	if len(masterKey) != 32 { // AES-256 key size
		t.Errorf("expected master key of size 32, got %d", len(masterKey))
	}

	// Submitting another share after unsealing should have no effect and return error
	_, _, err = s.Unseal(ctx, keyShares[threshold])
	if !errors.Is(err, ErrSealThresholdMet) {
		t.Errorf("expected ErrSealThresholdMet after vault is unsealed, got %v", err)
	}
}

func TestSeal_Unseal_InvalidShare(t *testing.T) {
	t.Run("invalid base64 format", func(t *testing.T) {
		ctx := context.Background()
		s := New(5, 3)
		_, _ = s.GenerateKeys(ctx)

		_, _, err := s.Unseal(ctx, "this-is-not-base64-!")
		if !errors.Is(err, ErrInvalidShare) {
			t.Fatalf("expected ErrInvalidShare for an invalid base64 share, got %v", err)
		}
	})

	t.Run("invalid shamir share content", func(t *testing.T) {
		ctx := context.Background()
		shares, threshold := 5, 3
		s := New(shares, threshold)
		validShares, err := s.GenerateKeys(ctx)
		if err != nil {
			t.Fatalf("key generation failed: %v", err)
		}

		// Submit two valid shares
		for i := 0; i < threshold-1; i++ {
			_, _, err = s.Unseal(ctx, validShares[i])
			if err != nil {
				t.Fatalf("expected no error when submitting valid share, but got %v", err)
			}
		}

		// Now submit a final, invalid share to trigger the combine
		invalidShare := base64.StdEncoding.EncodeToString([]byte("this is a valid base64 string but not a real shamir share"))
		isUnsealed, progress, err := s.Unseal(ctx, invalidShare)

		if isUnsealed {
			t.Fatal("vault should not be unsealed with an invalid share")
		}
		if !errors.Is(err, ErrInvalidShare) {
			t.Fatalf("expected ErrInvalidShare when combining, but got: %v", err)
		}
		// Progress should be reset to 0 after a failed combination attempt
		if progress != 0 {
			t.Errorf("expected progress to be reset to 0 after failed combine, got %d", progress)
		}
	})
}
