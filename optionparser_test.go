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
