package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

func ReadBytes(n int) ([]byte, error) {
	b := make([]byte, n)

	rn, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to read bytes: %w", err)
	}

	if rn != n {
		return nil, fmt.Errorf("failed to read correct number of bytes: %d", rn)
	}

	return b, nil
}

func ReadHex(n int) (string, error) {
	b := make([]byte, n)

	rn, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to read bytes: %w", err)
	}

	if rn != n {
		return "", fmt.Errorf("failed to read correct number of bytes: %d", rn)
	}

	return hex.EncodeToString(b), nil
}

func Sha256(key, seed string) (string, error) {
	h := sha256.New()

	if _, err := fmt.Fprint(h, key, seed); err != nil {
		return "", fmt.Errorf("failed to write args: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func CompareSha256(key, seed string, oldHash string) (bool, error) {
	nh, err := Sha256(key, seed)
	if err != nil {
		return false, fmt.Errorf("failed to calculate new hash: %w", err)
	}

	return subtle.ConstantTimeCompare([]byte(nh), []byte(oldHash)) == 1, nil
}
