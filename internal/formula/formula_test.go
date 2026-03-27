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

func TestAgentUsageText(t *testing.T) {
	explicit := &Formula{
		Name: "x", Version: "1", URL: "https://a/b",
		SHA256: "0000000000000000000000000000000000000000000000000000000000000000",
		Usage:  "Run with --help.\nSee docs.",
	}
	if got := explicit.AgentUsageText(); got != explicit.Usage {
		t.Fatalf("got %q want %q", got, explicit.Usage)
	}
	fallback := &Formula{
		Name: "x", Version: "1", URL: "https://a/b",
		SHA256:      "0000000000000000000000000000000000000000000000000000000000000000",
		Description: "Short desc",
		Model:       &ModelMeta{Summary: "Model summary"},
	}
	if got := fallback.AgentUsageText(); got != "Short desc\n\nModel summary" {
		t.Fatalf("got %q", got)
	}
}
