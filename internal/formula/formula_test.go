package formula

import (
	"testing"
)

func TestValidate(t *testing.T) {
	f := &Formula{Name: "x", Version: "1", URL: "https://a/b", SHA256: "abcd"}
	if err := f.Validate(); err == nil {
		t.Fatal("expected sha256 length error")
	}
	good := &Formula{
		Name: "x", Version: "1", URL: "https://a/b",
		SHA256: "0000000000000000000000000000000000000000000000000000000000000000",
	}
	if err := good.Validate(); err != nil {
		t.Fatal(err)
	}
}
