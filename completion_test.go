package optionparser

import (
	"bytes"
	"strings"
	"testing"
)

func sampleParser() *OptionParser {
	op := NewOptionParser()
	var s string
	options := map[string]string{}
	op.On("-a", "--alpha", "alpha description", options)
	op.On("-c", "--config NAME", "config file name", &s)
	op.On("--no-cutmarks", "toggle cutmarks", options)
	op.On("--port PORT", "tcp port", options)
	op.On("--mode [NAME]", "optional mode", options)
	op.Command("run", "Run the thing")
	op.Command("clean", "Clean files")
	return op
}

func TestGenerateCompletionUnsupported(t *testing.T) {
	op := sampleParser()
	var buf bytes.Buffer
	if err := op.GenerateCompletion("tcsh", "sp", &buf); err == nil {
		t.Errorf("expected error for unsupported shell")
	}
}

func TestGenerateCompletionEmptyProgram(t *testing.T) {
	op := sampleParser()
	var buf bytes.Buffer
	if err := op.GenerateCompletion("bash", "", &buf); err == nil {
		t.Errorf("expected error for empty programName")
	}
}

func TestBashCompletionContents(t *testing.T) {
	op := sampleParser()
	var buf bytes.Buffer
	if err := op.GenerateCompletion("bash", "sp", &buf); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, want := range []string{
		"complete -F _sp_completions sp",
		"--alpha",
		"--config",
		"-c",
		"--no-cutmarks",
		"--cutmarks",
		"--port",
		"run",
		"clean",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("bash output missing %q\n---\n%s", want, s)
		}
	}
	// value-taking flags must appear in the case statement
	if !strings.Contains(s, "--config") || !strings.Contains(s, "--port") {
		t.Errorf("expected value-taking flags in case statement")
	}
}

func TestZshCompletionContents(t *testing.T) {
	op := sampleParser()
	var buf bytes.Buffer
	if err := op.GenerateCompletion("zsh", "sp", &buf); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, want := range []string{
		"#compdef sp",
		"_sp()",
		"--alpha",
		"--config",
		"PORT:_files",
		"'run:Run the thing'",
		"'clean:Clean files'",
		`"${funcstack[1]-}" = "_sp"`,
		"compdef _sp sp",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("zsh output missing %q\n---\n%s", want, s)
		}
	}
	// --no-foo form must not take a value
	if strings.Contains(s, "--no-cutmarks[toggle cutmarks]:") {
		t.Errorf("--no- form should not declare a value spec\n%s", s)
	}
}

func TestFishCompletionContents(t *testing.T) {
	op := sampleParser()
	var buf bytes.Buffer
	if err := op.GenerateCompletion("fish", "sp", &buf); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, want := range []string{
		"function __sp_needs_command",
		"complete -c sp -s a -l alpha",
		"complete -c sp -s c -l config -r",
		"complete -c sp -l cutmarks",
		"complete -c sp -l no-cutmarks",
		"complete -c sp -l port -r",
		"-a 'run'",
		"-a 'clean'",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("fish output missing %q\n---\n%s", want, s)
		}
	}
}

func TestZshQuotingEscapesSpecials(t *testing.T) {
	op := NewOptionParser()
	options := map[string]string{}
	op.On("--quoted FOO", "weird 'help' [text]: with colon", options)
	var buf bytes.Buffer
	if err := op.GenerateCompletion("zsh", "demo", &buf); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if strings.Contains(s, "[text]") {
		t.Errorf("unescaped brackets in zsh output:\n%s", s)
	}
	if !strings.Contains(s, `\[text\]`) {
		t.Errorf("expected escaped brackets in zsh output:\n%s", s)
	}
}

func TestFishQuotingEscapesQuotes(t *testing.T) {
	op := NewOptionParser()
	options := map[string]string{}
	op.On("--quoted", "help with 'apostrophes'", options)
	var buf bytes.Buffer
	if err := op.GenerateCompletion("fish", "demo", &buf); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, `\'apostrophes\'`) {
		t.Errorf("expected escaped apostrophes in fish output:\n%s", s)
	}
}
