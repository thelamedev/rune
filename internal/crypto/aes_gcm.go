package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKeySize     = errors.New("invalid AES key size")
	ErrEncryptionFailed   = errors.New("failed to encrypt data")
	ErrDecryptionFailed   = errors.New("failed to decrypt data")
	ErrCiphertextTooShort = errors.New("ciphertext is too short")
)

const (
	// AES-256
	KeySize = 32
)

type AESGCMEngine struct {
	masterKey []byte
}

func NewAESGCM(masterKey []byte) (*AESGCMEngine, error) {
	if len(masterKey) != KeySize {
		return nil, ErrInvalidKeySize
	}

	return &AESGCMEngine{
		masterKey: masterKey,
	}, nil
}

// Encrypt performs envelope encryption on a given plaintext.
// It returns a single ciphertext blob containing the encrypted DEK and the encrypted data.
func (e *AESGCMEngine) Encrypt(plaintext []byte) ([]byte, error) {
	if len(e.masterKey) != KeySize {
		return nil, ErrInvalidKeySize
	}

	// 1. Generate a new, random Data Encryption Key (DEK).
	dek := make([]byte, KeySize)
	if _, err := rand.Read(dek); err != nil {
		return nil, fmt.Errorf("failed to generate DEK: %w", err)
	}

	// 2. Encrypt the DEK with the Master Key.
	encryptedDEK, err := e.aesGCMEncrypt(dek, e.masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt DEK: %w", err)
	}

	// 3. Encrypt the plaintext with the DEK.
	encryptedValue, err := e.aesGCMEncrypt(plaintext, dek)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt value: %w", err)
	}

	// 4. Construct the final payload: len(encryptedDEK) | encryptedDEK | encryptedValue
	// We use 2 bytes for the length, allowing DEKs up to 65535 bytes, which is plenty.
	payload := make([]byte, 2+len(encryptedDEK)+len(encryptedValue))
	binary.BigEndian.PutUint16(payload[:2], uint16(len(encryptedDEK)))
	copy(payload[2:], encryptedDEK)
	copy(payload[2+len(encryptedDEK):], encryptedValue)

	return payload, nil
}

// Decrypt reverses the envelope encryption process.
func (e *AESGCMEngine) Decrypt(payload []byte) ([]byte, error) {
	if len(e.masterKey) != KeySize {
		return nil, ErrInvalidKeySize
	}
	if len(payload) < 2 {
		return nil, ErrCiphertextTooShort
	}

	// 1. Deconstruct the payload: len(encryptedDEK) | encryptedDEK | encryptedValue
	dekLen := int(binary.BigEndian.Uint16(payload[:2]))
	if len(payload) < 2+dekLen {
		return nil, ErrCiphertextTooShort
	}
	encryptedDEK := payload[2 : 2+dekLen]
	encryptedValue := payload[2+dekLen:]

	// 2. Decrypt the DEK with the Master Key.
	dek, err := e.aesGCMDecrypt(encryptedDEK, e.masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data encryption key: %w", err)
	}

	// 3. Decrypt the value with the DEK.
	plaintext, err := e.aesGCMDecrypt(encryptedValue, dek)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt value: %w", err)
	}

	return plaintext, nil
}

// aesGCMEncrypt is a helper for AES-GCM encryption.
func (e *AESGCMEngine) aesGCMEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// aesGCMDecrypt is a helper for AES-GCM decryption.
func (e *AESGCMEngine) aesGCMDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrCiphertextTooShort
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}
	return plaintext, nil
}
