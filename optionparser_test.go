package optionparser

import (
	"testing"
)

func TestSimple(t *testing.T) {
	op := NewOptionParser()
	var str string
	op.On("-x", "--x ARG", "helptext", &str)
	args := []string{"foo", "-x", "abc"}
	err := op.ParseFrom(args)
	if err != nil {
		t.Error(err)
	}
	expected := "abc"
	if got := str; got != expected {
		t.Errorf("op.Parse() str = %s, want %s", got, expected)
	}
}

func TestDash(t *testing.T) {
	op := NewOptionParser()
	options := make(map[string]string)
	op.On("-d", "--dash DASH", "check for dash error", options)
	err := op.ParseFrom([]string{"", "-d", "-"})
	if err != nil {
		t.Error(err)
	}
	if options["dash"] != "-" {
		t.Errorf(`options["dash"] = %s want %s`, options["dash"], "-")
	}
}

func ExampleOptionParser_Help() {
	op := NewOptionParser()
	var str string
	var options = make(map[string]string)
	var truefalse bool
	var stringslice []string
	myfunc := func() { return }
	op.On("-a", "--func", "call myfunc", myfunc)
	op.On("--bstring FOO", "set string to FOO", &str)
	op.On("-c", "set boolean option (try -no-c)", options)
	op.On("-d", "--dlong VAL", "set option", options)
	op.On("-e", "--elong [VAL]", "set option with optional parameter", options)
	op.On("-f", "boolean option", &truefalse)
	op.On("-g VALUES", "give multiple values", &stringslice)
	op.Help()

	//Output:
	// Usage: [parameter] command
	// -h, --help                   Show this help
	// -a, --func                   call myfunc
	//     --bstring=FOO            set string to FOO
	// -c                           set boolean option (try -no-c)
	// -d, --dlong=VAL              set option
	// -e, --elong[=VAL]            set option with optional parameter
	// -f                           boolean option
	// -g=VALUES                    give multiple values
}

func TestParse(t *testing.T) {
	op := NewOptionParser()
	var str string
	var options = make(map[string]string)
	var truefalse bool
	var stringslice []string
	myfunc := func() { options["func"] = "ok" }
	op.On("-a", "--func", "call myfunc", myfunc)
	op.On("--bstring FOO", "set string to FOO", &str)
	op.On("-c", "set boolean option (try -no-c)", options)
	op.On("-d", "--dlong VAL", "set option", options)
	op.On("-e", "--elong [VAL]", "set option with optional parameter", options)
	op.On("-f", "boolean option", &truefalse)
	op.On("-g VALUES", "give multiple values", &stringslice)
	op.ParseFrom([]string{"", "--func", "--bstring barg", "-no-c", "-d darg", "-e", "eopt", "-g", "a,b,c"})
	for k, v := range map[string]string{"c": "false", "dlong": "darg", "elong": "eopt", "func": "ok"} {
		if options[k] != v {
			t.Errorf("op.ParseFrom() = %s, want %s", options[k], v)
		}
	}
	if expected := "barg"; str != expected {
		t.Errorf("op.ParseFrom() --bstring = %s, want %s", str, expected)
	}
}
