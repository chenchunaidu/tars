package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifySHA256(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(p, []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	want := "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03"
	if err := VerifySHA256(p, want); err != nil {
		t.Fatal(err)
	}
	bad := "0000000000000000000000000000000000000000000000000000000000000000"
	if err := VerifySHA256(p, bad); err == nil {
		t.Fatal("expected mismatch")
	}
}

func TestEqualHex(t *testing.T) {
	if !EqualHex("Ab", "ab") {
		t.Fatal()
	}
}
