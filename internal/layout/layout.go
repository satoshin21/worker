// Package layout renders the embedded zellij KDL layout used by `worker create`.
package layout

import (
	_ "embed"
	"strings"
)

//go:embed default.kdl
var defaultKDL string

const claudeArgsPlaceholder = "__WORKER_CLAUDE_ARGS__"

// Render returns the layout KDL with the claude pane arguments populated.
// When instruction is empty the args line is omitted so claude starts without arguments.
func Render(instruction string) string {
	var replacement string
	if instruction != "" {
		replacement = "      args " + kdlQuote(instruction)
	}
	out := strings.Replace(defaultKDL, claudeArgsPlaceholder, replacement, 1)
	if replacement == "" {
		out = strings.Replace(out, "\n\n", "\n", 1)
	}
	return out
}

// kdlQuote escapes a string for use inside a KDL double-quoted literal.
func kdlQuote(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
