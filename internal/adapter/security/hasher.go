// Package security містить адаптери безпеки (хешування паролів).
package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

// SHA256Hasher — солоний ітеративний SHA-256, що не потребує зовнішніх залежностей.
// ПРИМІТКА: для production замініть на bcrypt/argon2 (реалізує service.PasswordHasher).
type SHA256Hasher struct{ iterations int }

func NewSHA256Hasher() *SHA256Hasher { return &SHA256Hasher{iterations: 100_000} }

func (h *SHA256Hasher) digest(salt []byte, raw string) []byte {
	sum := sha256.Sum256(append(append([]byte{}, salt...), raw...))
	for i := 1; i < h.iterations; i++ {
		sum = sha256.Sum256(sum[:])
	}
	return sum[:]
}

func (h *SHA256Hasher) Hash(raw string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return hex.EncodeToString(salt) + "$" + hex.EncodeToString(h.digest(salt, raw)), nil
}

func (h *SHA256Hasher) Verify(raw, hash string) bool {
	parts := strings.SplitN(hash, "$", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	want, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(want, h.digest(salt, raw)) == 1
}
