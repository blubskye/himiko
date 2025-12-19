// Himiko Discord Bot
// Copyright (C) 2025 Himiko Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package crypto provides field-level encryption utilities for sensitive database fields.
// Uses AES-256-GCM with PBKDF2 key derivation.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// PBKDF2 iterations for key derivation
	keyIterations = 100000
	// Key size for AES-256
	keySize = 32
	// Salt for key derivation (fixed for consistency)
	keySalt = "himiko-field-encryption-v1"
	// Minimum ciphertext length (nonce + at least 1 byte + auth tag)
	minCiphertextLen = 12 + 1 + 16
)

// FieldEncryptor handles encryption and decryption of sensitive database fields.
// It is safe for concurrent use.
type FieldEncryptor struct {
	gcm     cipher.AEAD
	enabled bool
}

// NewFieldEncryptor creates a new FieldEncryptor with the given passphrase.
// If passphrase is empty, encryption is disabled and Encrypt/Decrypt are no-ops.
func NewFieldEncryptor(passphrase string) (*FieldEncryptor, error) {
	if passphrase == "" {
		return &FieldEncryptor{enabled: false}, nil
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(passphrase), []byte(keySalt), keyIterations, keySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &FieldEncryptor{
		gcm:     gcm,
		enabled: true,
	}, nil
}

// IsEnabled returns whether encryption is enabled.
func (e *FieldEncryptor) IsEnabled() bool {
	return e.enabled
}

// Encrypt encrypts a plaintext string and returns a base64-encoded ciphertext.
// Returns the original string unchanged if encryption is disabled or input is empty.
func (e *FieldEncryptor) Encrypt(plaintext string) (string, error) {
	if !e.enabled || plaintext == "" {
		return plaintext, nil
	}

	// Generate random nonce
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and append auth tag
	ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Base64 encode
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded ciphertext and returns the plaintext.
// Returns the original string unchanged if:
// - Encryption is disabled
// - Input is empty
// - Input doesn't appear to be encrypted (too short or not valid base64)
// - Decryption fails (assumed to be legacy unencrypted data)
func (e *FieldEncryptor) Decrypt(ciphertext string) (string, error) {
	if !e.enabled || ciphertext == "" {
		return ciphertext, nil
	}

	// Try to decode base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		// Not valid base64, return as-is (legacy unencrypted data)
		return ciphertext, nil
	}

	// Check minimum length
	if len(data) < minCiphertextLen {
		// Too short to be encrypted, return as-is
		return ciphertext, nil
	}

	// Extract nonce and ciphertext
	nonce := data[:e.gcm.NonceSize()]
	encrypted := data[e.gcm.NonceSize():]

	// Decrypt
	plaintext, err := e.gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		// Decryption failed, return original (legacy unencrypted data)
		return ciphertext, nil
	}

	return string(plaintext), nil
}

// EncryptNullable encrypts a nullable string (pointer).
// Returns nil if input is nil.
func (e *FieldEncryptor) EncryptNullable(plaintext *string) (*string, error) {
	if plaintext == nil {
		return nil, nil
	}
	encrypted, err := e.Encrypt(*plaintext)
	if err != nil {
		return nil, err
	}
	return &encrypted, nil
}

// DecryptNullable decrypts a nullable string (pointer).
// Returns nil if input is nil.
func (e *FieldEncryptor) DecryptNullable(ciphertext *string) (*string, error) {
	if ciphertext == nil {
		return nil, nil
	}
	decrypted, err := e.Decrypt(*ciphertext)
	if err != nil {
		return nil, err
	}
	return &decrypted, nil
}

// IsEncrypted checks if a string appears to be encrypted.
// This is a heuristic check based on base64 encoding and minimum length.
func (e *FieldEncryptor) IsEncrypted(data string) bool {
	if data == "" {
		return false
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return false
	}

	return len(decoded) >= minCiphertextLen
}

// MustEncrypt encrypts and panics on error (for use in tests or where errors shouldn't occur)
func (e *FieldEncryptor) MustEncrypt(plaintext string) string {
	result, err := e.Encrypt(plaintext)
	if err != nil {
		panic(err)
	}
	return result
}

// MustDecrypt decrypts and panics on error (for use in tests)
func (e *FieldEncryptor) MustDecrypt(ciphertext string) string {
	result, err := e.Decrypt(ciphertext)
	if err != nil {
		panic(err)
	}
	return result
}

// ValidateKey checks if the encryption key can decrypt test data
func (e *FieldEncryptor) ValidateKey() error {
	if !e.enabled {
		return errors.New("encryption is not enabled")
	}

	testData := "himiko-encryption-test"
	encrypted, err := e.Encrypt(testData)
	if err != nil {
		return fmt.Errorf("encryption test failed: %w", err)
	}

	decrypted, err := e.Decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("decryption test failed: %w", err)
	}

	if decrypted != testData {
		return errors.New("encryption round-trip failed: data mismatch")
	}

	return nil
}
