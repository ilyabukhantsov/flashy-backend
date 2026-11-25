// Package idgen генерує ідентифікатори та випадкові токени.
package idgen

import (
	"crypto/rand"
	"encoding/hex"
)

// New повертає 128-бітний ідентифікатор у hex.
func New() string { return randomHex(16) }

// Token повертає 256-бітний токен сесії у hex.
func Token() string { return randomHex(32) }

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
