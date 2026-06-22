package layout

import (
	"strings"
	"testing"
)

func TestRenderWithoutInstruction(t *testing.T) {
	out := Render("")
	if strings.Contains(out, "__WORKER_CLAUDE_ARGS__") {
		t.Fatalf("placeholder not replaced: %s", out)
	}
	if strings.Contains(out, "args ") {
		t.Fatalf("args line should be omitted when instruction is empty: %s", out)
	}
}

func TestRenderWithInstruction(t *testing.T) {
	out := Render(`refactor "auth"`)
	if !strings.Contains(out, `args "refactor \"auth\""`) {
		t.Fatalf("expected escaped args line, got: %s", out)
	}
}
