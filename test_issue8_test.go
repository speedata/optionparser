package optionparser

import (
"strings"
"testing"
)

// TestIssue8_FunctionCallback tests the scenario from issue #8
// where a user defines a callback that expects "key=value" format
func TestIssue8_FunctionCallback(t *testing.T) {
options := make(map[string]string)

setOption := func(str string) {
// User's callback expects format "key=value"
a := strings.Split(str, "=")
if len(a) < 2 {
t.Errorf("Expected format 'key=value', got %q (would panic with index out of range)", str)
return
}
options[a[0]] = a[1]
}

op := NewOptionParser()
op.On("--option=OPTION", "Set a specific option", setOption)

// This is what the user ran: sp --option aa bb
err := op.ParseFrom([]string{"prog", "--option", "aa", "bb"})
if err != nil {
t.Errorf("Unexpected error: %v", err)
}

// The parser passes "aa" to the callback, but the callback expects "key=value"
// This would cause a panic in real code when accessing a[1]
}

// TestIssue8_WithProperFormat tests that when the user provides the right format, it works
func TestIssue8_WithProperFormat(t *testing.T) {
options := make(map[string]string)

setOption := func(str string) {
a := strings.Split(str, "=")
if len(a) < 2 {
t.Errorf("Expected format 'key=value', got %q", str)
return
}
options[a[0]] = a[1]
}

op := NewOptionParser()
op.On("--option=OPTION", "Set a specific option", setOption)

// User provides: sp --option key=value bb
err := op.ParseFrom([]string{"prog", "--option", "key=value", "bb"})
if err != nil {
t.Errorf("Unexpected error: %v", err)
}

// Check that it was parsed correctly
if options["key"] != "value" {
t.Errorf("Expected options[key]=value, got %q", options["key"])
}

if len(op.Extra) != 1 || op.Extra[0] != "bb" {
t.Errorf("Expected Extra=[bb], got %v", op.Extra)
}
}
