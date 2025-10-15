package optionparser

import (
	"strings"
	"testing"
)

// TestIssue8_FunctionCallbackPanic tests the scenario from issue #8
// where a user defines a callback that would panic without proper error handling.
// The parser should now catch the panic and return an error.
func TestIssue8_FunctionCallbackPanic(t *testing.T) {
	options := make(map[string]string)

	setOption := func(str string) {
		// User's callback that would panic without bounds checking
		a := strings.Split(str, "=")
		// This line would panic if len(a) < 2
		options[a[0]] = a[1]
	}

	op := NewOptionParser()
	op.On("--option=OPTION", "Set a specific option", setOption)

	// This is what the user ran: sp --option aa bb
	// The parser should catch the panic and return an error
	err := op.ParseFrom([]string{"prog", "--option", "aa", "bb"})
	if err == nil {
		t.Error("Expected error from callback panic, got nil")
	} else {
		t.Logf("Got expected error: %v", err)
	}
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

// TestIssue8_WithProperErrorHandling shows the recommended pattern
func TestIssue8_WithProperErrorHandling(t *testing.T) {
	options := make(map[string]string)

	setOption := func(str string) {
		// Proper error handling in callback
		a := strings.Split(str, "=")
		if len(a) < 2 {
			// Log or handle the error gracefully
			t.Logf("Warning: expected format 'key=value', got %q", str)
			return
		}
		options[a[0]] = a[1]
	}

	op := NewOptionParser()
	op.On("--option=OPTION", "Set a specific option", setOption)

	// User provides: sp --option aa bb
	// This should not panic or error - the callback handles it gracefully
	err := op.ParseFrom([]string{"prog", "--option", "aa", "bb"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// The option wasn't set because it didn't have the right format
	if len(options) != 0 {
		t.Errorf("Expected no options to be set, got %v", options)
	}

	if len(op.Extra) != 1 || op.Extra[0] != "bb" {
		t.Errorf("Expected Extra=[bb], got %v", op.Extra)
	}
}

// TestIssue8_FunctionNoArgsPanic tests that panics in functionNoArgs callbacks are also caught
func TestIssue8_FunctionNoArgsPanic(t *testing.T) {
	panicFunc := func() {
		// This will panic
		var slice []string
		_ = slice[10] // index out of range
	}

	op := NewOptionParser()
	op.On("-x", "trigger panic", panicFunc)

	// The parser should catch the panic and return an error
	err := op.ParseFrom([]string{"prog", "-x"})
	if err == nil {
		t.Error("Expected error from callback panic, got nil")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}
