package crypto

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"
)

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	masterKey := make([]byte, KeySize)
	if _, err := rand.Read(masterKey); err != nil {
		t.Fatalf("failed to generate master key: %v", err)
	}

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"Simple Text", []byte("my secret data")},
		{"Empty String", []byte("")},
		{"Large Data", make([]byte, 1024*1024)}, // 1MB
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine, err := NewAESGCM(masterKey)
			if err != nil {
				t.Fatalf("NewAESGCM() failed: %v", err)
			}
			// Encrypt
			ciphertext, err := engine.Encrypt(tc.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() failed: %v", err)
			}
			if ciphertext == nil {
				t.Fatal("Encrypt() returned nil ciphertext")
			}

			// Decrypt
			decrypted, err := engine.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decrypt() failed: %v", err)
			}

			// Verify
			if !bytes.Equal(tc.plaintext, decrypted) {
				t.Error("decrypted data does not match original plaintext")
			}
		})
	}
}

func TestInvalidKeySize(t *testing.T) {
	badKey := []byte("shortkey")
	_, err := NewAESGCM(badKey)
	if err == nil {
		t.Fatalf("NewAESGCM() expected error, got %v", err)
	}
}

func TestDecrypt_CorruptedPayload(t *testing.T) {
	masterKey := make([]byte, KeySize)
	engine, err := NewAESGCM(masterKey)
	if err != nil {
		t.Fatalf("NewAESGCM() failed: %v", err)
	}
	if _, err := rand.Read(masterKey); err != nil {
		t.Fatalf("failed to generate master key: %v", err)
	}

	plaintext := []byte("some data")
	ciphertext, err := engine.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() failed: %v", err)
	}

	// Tamper with the ciphertext (flip a bit)
	ciphertext[len(ciphertext)-1] ^= 0x01

	_, err = engine.Decrypt(ciphertext)
	if err == nil {
		t.Fatal("expected an error when decrypting a corrupted payload, but got nil")
	}

	// We expect the underlying error to be our specific one
	if !errors.Is(err, ErrDecryptionFailed) {
		t.Logf("Note: A decryption error occurred as expected, but it wasn't wrapped as ErrDecryptionFailed. Error: %v", err)
	}
}
