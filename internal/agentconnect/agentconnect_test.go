package agentconnect

import "testing"

func TestOptionsFromConnectArgs(t *testing.T) {
	o, err := OptionsFromConnectArgs([]string{"all"})
	if err != nil || o.SkipCursor || o.SkipClaude || o.SkipGemini || o.SkipPi {
		t.Fatalf("all: err=%v opts=%+v", err, o)
	}
	o, err = OptionsFromConnectArgs([]string{"cursor"})
	if err != nil || o.SkipCursor || !o.SkipClaude || !o.SkipGemini || !o.SkipPi {
		t.Fatalf("cursor: err=%v opts=%+v", err, o)
	}
	o, err = OptionsFromConnectArgs([]string{"cursor", "gemini"})
	if err != nil || o.SkipCursor || !o.SkipClaude || o.SkipGemini || !o.SkipPi {
		t.Fatalf("cursor+gemini: err=%v opts=%+v", err, o)
	}
	o, err = OptionsFromConnectArgs([]string{"cursor", "cursor"})
	if err != nil || o.SkipCursor || !o.SkipClaude || !o.SkipGemini || !o.SkipPi {
		t.Fatalf("dup: err=%v opts=%+v", err, o)
	}
	if _, err := OptionsFromConnectArgs([]string{}); err == nil {
		t.Fatal("empty args: want error")
	}
	if _, err := OptionsFromConnectArgs([]string{"nope"}); err == nil {
		t.Fatal("bad name: want error")
	}
	if _, err := OptionsFromConnectArgs([]string{"all", "cursor"}); err == nil {
		t.Fatal("all+agent: want error")
	}
}

func TestReplaceMarkedBlock(t *testing.T) {
	old := "hello\n<!-- tars-connect:begin -->\nold\n<!-- tars-connect:end -->\nworld"
	newB := "<!-- tars-connect:begin -->\nnew\n<!-- tars-connect:end -->"
	got := replaceMarkedBlock(old, newB)
	want := "hello\n<!-- tars-connect:begin -->\nnew\n<!-- tars-connect:end -->\nworld"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
