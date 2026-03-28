package formula

import (
	"strings"
	"testing"
)

const hex64 = "0000000000000000000000000000000000000000000000000000000000000000"

func TestValidate(t *testing.T) {
	f := &Formula{Name: "x", Version: "1", URL: "https://a/b", SHA256: "abcd"}
	if err := f.Validate(); err == nil {
		t.Fatal("expected sha256 length error")
	}
	good := &Formula{
		Name: "x", Version: "1", URL: "https://a/b",
		SHA256: hex64,
	}
	if err := good.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePlatforms(t *testing.T) {
	raw := `{
  "name": "rg",
  "version": "1",
  "platforms": {
    "Linux_AMD64": {"url": "https://x/linux.tgz", "sha256": "` + hex64 + `"},
    "default": {"url": "https://x/fallback.tgz", "sha256": "` + hex64 + `"}
  }
}`
	f, err := Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	u, h, key, err := f.ResolveArtifact("linux", "amd64")
	if err != nil || u != "https://x/linux.tgz" || h != hex64 || key != "linux_amd64" {
		t.Fatalf("ResolveArtifact linux_amd64: %v %q %q %q", err, u, h, key)
	}
	u, h, key, err = f.ResolveArtifact("darwin", "arm64")
	if err != nil || u != "https://x/fallback.tgz" || key != PlatformDefault {
		t.Fatalf("ResolveArtifact default: %v %q key=%q", err, u, key)
	}
}

func TestResolveArtifactMissingPlatform(t *testing.T) {
	raw := `{
  "name": "x",
  "version": "1",
  "platforms": {
    "linux_amd64": {"url": "https://a", "sha256": "` + hex64 + `"}
  }
}`
	f, err := Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, err = f.ResolveArtifact("darwin", "arm64")
	if err == nil || !strings.Contains(err.Error(), PlatformDefault) {
		t.Fatalf("expected error mentioning default, got %v", err)
	}
}

func TestLegacyURLStillWorks(t *testing.T) {
	raw := `{"name":"x","version":"1","url":"https://a/b.zip","sha256":"` + hex64 + `"}` 
	f, err := Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	u, h, key, err := f.ResolveArtifact("windows", "386")
	if err != nil || u != "https://a/b.zip" || h != hex64 || key != "" {
		t.Fatalf("legacy: %v url=%q key=%q", err, u, key)
	}
}

func TestAgentUsageText(t *testing.T) {
	empty := &Formula{
		Name: "x", Version: "1", URL: "https://a/b",
		SHA256: hex64,
	}
	if got := empty.AgentUsageText(); got != "" {
		t.Fatalf("empty: got %q want empty", got)
	}
	descOnly := &Formula{
		Name: "x", Version: "1", URL: "https://a/b",
		SHA256:      hex64,
		Description: "Primary for agents",
	}
	if got := descOnly.AgentUsageText(); got != "Primary for agents" {
		t.Fatalf("description only: got %q", got)
	}
	descAndSummary := &Formula{
		Name: "x", Version: "1", URL: "https://a/b",
		SHA256:      hex64,
		Description: "Short desc",
		Model:       &ModelMeta{Summary: "Model summary"},
	}
	if got := descAndSummary.AgentUsageText(); got != "Short desc\n\nModel summary" {
		t.Fatalf("got %q", got)
	}
}
