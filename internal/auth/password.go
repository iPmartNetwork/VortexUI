// Package auth holds the pure security primitives the panel depends on: password
// hashing, JWT issuing/verification, and TOTP. It has no knowledge of HTTP,
// storage, or the domain services, so each primitive is small and unit-tested.
package auth

import "golang.org/x/crypto/bcrypt"

// DefaultCost is the bcrypt work factor. 12 is a sensible 2020s default: strong
// against offline cracking while keeping login latency acceptable.
const DefaultCost = 12

// HashPassword returns the bcrypt hash of a plaintext password.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), DefaultCost)
	return string(b), err
}

// CheckPassword reports whether plain matches the stored bcrypt hash. It runs in
// constant time relative to the hash, so it does not leak match progress.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
