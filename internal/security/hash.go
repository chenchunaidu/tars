package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// FileSHA256 returns lowercase hex SHA256 of a file.
func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return ReaderSHA256(f)
}

// ReaderSHA256 hashes r to completion.
func ReaderSHA256(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifySHA256 compares expected hex (any case) to file hash.
func VerifySHA256(path, expectedHex string) error {
	got, err := FileSHA256(path)
	if err != nil {
		return err
	}
	if !EqualHex(got, expectedHex) {
		return fmt.Errorf("sha256 mismatch: expected %s, got %s", expectedHex, got)
	}
	return nil
}

// EqualHex compares two hex strings case-insensitively.
func EqualHex(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'F' {
			ca = ca - 'A' + 'a'
		}
		if cb >= 'A' && cb <= 'F' {
			cb = cb - 'A' + 'a'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
