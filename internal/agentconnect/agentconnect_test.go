package agentconnect

import "testing"

func TestReplaceMarkedBlock(t *testing.T) {
	old := "hello\n<!-- tars-connect:begin -->\nold\n<!-- tars-connect:end -->\nworld"
	newB := "<!-- tars-connect:begin -->\nnew\n<!-- tars-connect:end -->"
	got := replaceMarkedBlock(old, newB)
	want := "hello\n<!-- tars-connect:begin -->\nnew\n<!-- tars-connect:end -->\nworld"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
