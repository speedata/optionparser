// Package optionparser is a library for defining and parsing command line
// options. It aims to provide a natural language interface for defining short
// and long parameters and mandatory and optional arguments. It provides the
// user with nice output formatting on the built-in method '--help'.
package optionparser

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// ErrHelp is returned by Parse/ParseFrom when the user requested help.
// Callers can check for this sentinel error and decide whether to os.Exit(0).
var ErrHelp = errors.New("help requested")

// A command is a non-dash option (with a helptext)
type command struct {
	name     string
	helptext string
}

// OptionParser contains the methods to parse options and the settings to
// influence the output of --help. Set the Banner and Coda for usage info, set
// Start and Stop for output of the long description text.
type OptionParser struct {
	Extra    []string
	Banner   string
	Coda     string
	Start    int
	Stop     int
	options  []*allowedOptions
	short    map[string]*allowedOptions
	long     map[string]*allowedOptions
	commands []command

	// Out is the writer Help() and related formatting functions write to.
	// Defaults to os.Stdout.
	Out io.Writer

	// internal flag toggled by the default --help/-h handler
	showHelp bool
}

type argumentDescription struct {
	argument string
	param    string
	optional bool
	short    bool
	negate   bool
}

type allowedOptions struct {
	optional       bool
	param          string
	short          string
	long           string
	boolParameter  bool
	function       func(string)
	functionNoArgs func()
	boolvalue      *bool
	stringvalue    *string
	stringmap      map[string]string
	stringslice    *[]string
	helptext       string
}

var (
	// Example matches: "--", "-x", "-x=1", "-name", but not just "-" and not negative numbers once parsed as extras.
	isOptionRe     = regexp.MustCompile(`^(?:--.*|-[^-].*)$`)
	doubleDashRe   = regexp.MustCompile(`^--`)
	singleDashRe   = regexp.MustCompile(`^-[^-]`)
	spaceOrEqualRe = regexp.MustCompile(`[ \t=]`)
)

// Return true if s starts with a dash in a way that looks like an option.
func isOption(s string) bool {
	return isOptionRe.MatchString(s)
}

func wordwrap(s string, wd int) []string {
	// defensive
	if wd <= 0 {
		return []string{s}
	}
	// if the string is shorter than the width, we can just return it
	if len(s) <= wd {
		return []string{s}
	}

	// split at the last occurrence of space before wd
	stop := strings.LastIndex(s[0:wd], " ")
	// no space found in the next wd characters
	if stop < 0 {
		// try the first space after wd
		stop = strings.Index(s, " ")
		if stop < 0 { // no space at all
			return []string{s}
		}
	}
	a := []string{s[0:stop]}
	j := wordwrap(s[stop+1:], wd)
	return append(a, j...)
}

// Analyze the given argument such as '-s' or '--foo=bar' and
// return an argumentDescription
func splitOn(arg string) *argumentDescription {
	var (
		argument string
		param    string
		optional bool
		short    bool
		negate   bool
	)

	switch {
	case doubleDashRe.MatchString(arg):
		short = false
	case singleDashRe.MatchString(arg):
		short = true
	default:
		panic("unexpected argument shape")
	}

	init := 1
	if !short {
		init = 2
	}

	// Allow negation for long and short options (e.g., --no-foo, -no-f)
	if len(arg) >= init+3 && arg[init:init+3] == "no-" {
		negate = true
		init += 3
	}

	loc := spaceOrEqualRe.FindStringIndex(arg)
	if loc == nil {
		// no parameter delimiter; we know everything we need to know
		return &argumentDescription{
			argument: arg[init:],
			optional: false,
			short:    short,
			negate:   negate,
		}
	}

	// Now we know that the option may require an argument; it could be optional
	argument = arg[init:loc[0]]
	pos := loc[1]
	length := len(arg)

	optional = false
	if pos < len(arg) && arg[pos] == '[' {
		pos++
		length--
		optional = true
	}
	if pos > length { // be defensive
		return &argumentDescription{
			argument: argument,
			optional: optional,
			short:    short,
			negate:   negate,
		}
	}
	param = arg[pos:length]

	return &argumentDescription{
		argument: argument,
		param:    param,
		optional: optional,
		short:    short,
		negate:   negate,
	}
}

// prints the nice help output
func (op *OptionParser) formatAndOutput(start int, stop int, dashShort string, short string, comma string, dashLong string, long string, lines []string) {
	// clamp field widths to avoid negative widths
	pad := func(x int) int {
		if x < 0 {
			return 0
		}
		return x
	}

	if long == "" && len(short) > 2 {
		formatString := fmt.Sprintf("%%-1s%%-%d.%ds %%s\n", pad(start-3), pad(stop-3))
		fmt.Fprintf(op.Out, formatString, dashShort, short, lines[0])
	} else {
		formatString := fmt.Sprintf("%%-1s%%-1s%%1s %%-2s%%-%d.%ds %%s\n", pad(start-8), pad(stop-8))
		fmt.Fprintf(op.Out, formatString, dashShort, short, comma, dashLong, long, lines[0])
	}
	if len(lines) > 0 {
		formatString := fmt.Sprintf("%%%ds%%s\n", pad(start-1))
		for i := 1; i < len(lines); i++ {
			fmt.Fprintf(op.Out, formatString, " ", lines[i])
		}
	}
}

func set(obj *allowedOptions, hasNoPrefix bool, param string) {
	if obj.function != nil {
		obj.function(param)
	}
	if obj.stringvalue != nil {
		*obj.stringvalue = param
	}
	if obj.stringmap != nil {
		name := obj.long
		if name == "" {
			name = obj.short
		}
		var value string
		if param != "" {
			value = param
		} else if hasNoPrefix {
			value = "false"
		} else {
			value = "true"
		}
		obj.stringmap[name] = value
	}
	if obj.stringslice != nil && param != "" {
		*obj.stringslice = append(*obj.stringslice, strings.Split(param, ",")...)
	}
	if obj.functionNoArgs != nil {
		obj.functionNoArgs()
	}
	if obj.boolvalue != nil {
		*obj.boolvalue = !hasNoPrefix
	}
}

// Command defines optional arguments to the command line. These are written in
// a separate section called 'Commands' on --help.
func (op *OptionParser) Command(cmd string, helptext string) {
	cmds := command{cmd, helptext}
	op.commands = append(op.commands, cmds)
}

// On defines arguments and parameters. Each argument is one of:
//   - a short option, such as "-x",
//   - a long option, such as "--extra",
//   - a long option with an argument such as "--extra FOO" (or "--extra=FOO") for a mandatory argument,
//   - a long option with an argument in brackets, e.g. "--extra [FOO]" for a parameter with optional argument,
//   - a string (not starting with "-") used for the parameter description, e.g. "This parameter does this and that",
//   - a string variable in the form of &str that is used for saving the result of the argument,
//   - a variable of type map[string]string which is used to store the result (the parameter name is the key, the value is either the string true or the argument given on the command line)
//   - a variable of type *[]string which gets a comma separated list of values,
//   - a bool variable (in the form &bool) to hold a boolean value, or
//   - a function in the form of func() or in the form of func(string) which gets called if the command line parameter is found.
//
// On panics if the user supplies a type in its argument other than the ones
// given above.
//
//	op := optionparser.NewOptionParser()
//	op.On("-a", "--func", "call myfunc", myfunc)
//	op.On("--bstring FOO", "set string to FOO", &somestring)
//	op.On("-c", "set boolean option (try --no-c)", options)
//	op.On("-d", "--dlong VAL", "set option", options)
//	op.On("-e", "--elong [VAL]", "set option with optional parameter", options)
//	op.On("-f", "boolean option", &truefalse)
//	op.On("-g VALUES", "give multiple values", &stringslice)
//
// and running the program with --help gives the following output:
//
//	go run main.go --help
//	Usage: [parameter] command
//	   -h, --help                   Show this help
//	   -a, --func                   call myfunc
//	       --bstring=FOO            set string to FOO
//	   -c                           set boolean option (try --no-c)
//	   -d, --dlong=VAL              set option
//	   -e, --elong[=VAL]            set option with optional parameter
//	   -f                           boolean option
//	   -g=VALUES                    give multiple values
func (op *OptionParser) On(a ...interface{}) {
	option := new(allowedOptions)
	op.options = append(op.options, option)
	for _, i := range a {
		switch x := i.(type) {
		case string:
			// a short option, a long option or a help text
			if isOption(x) {
				ret := splitOn(x)
				if ret.short {
					// check duplicates
					if _, ok := op.short[ret.argument]; ok {
						panic(fmt.Sprintf("short option -%s already registered", ret.argument))
					}
					// short argument ('-s')
					op.short[ret.argument] = option
					option.short = ret.argument
				} else {
					// check duplicates
					if _, ok := op.long[ret.argument]; ok {
						panic(fmt.Sprintf("long option --%s already registered", ret.argument))
					}
					// long argument ('--something')
					op.long[ret.argument] = option
					option.long = ret.argument
				}
				if ret.optional {
					option.optional = true
				}
				if ret.param != "" {
					option.param = ret.param
				}
				if ret.negate {
					option.boolParameter = true
				}
			} else {
				// a string, probably the help text
				option.helptext = x
			}
		case func(string):
			option.function = x
		case func():
			option.functionNoArgs = x
		case *bool:
			option.boolvalue = x
		case *string:
			option.stringvalue = x
		case map[string]string:
			option.stringmap = x
		case *[]string:
			option.stringslice = x
		default:
			panic(fmt.Sprintf("Unknown parameter type: %#T\n", x))
		}
	}
}

// ParseFrom takes a slice of string arguments and interprets them. If it finds
// an unknown option or a missing mandatory argument, it returns an error.
// Note: this function expects args like os.Args (program name at index 0) and
// starts parsing at index 1.
func (op *OptionParser) ParseFrom(args []string) error {
	i := 1
	for i < len(args) {
		switch {
		// Users can pass -- to mark the end of flag parsing. This
		// check must come first since isOption will treat -- as
		// a (special) option boundary.
		case args[i] == "--":
			if i+1 < len(args) {
				op.Extra = append(op.Extra, args[i+1:]...)
			}
			if op.showHelp {
				op.Help()
				return ErrHelp
			}
			return nil

		case isOption(args[i]):
			ret := splitOn(args[i])

			var option *allowedOptions
			if ret.short {
				option = op.short[ret.argument]
			} else {
				option = op.long[ret.argument]
			}

			if option == nil {
				return fmt.Errorf("unknown option %q", args[i])
			}

			// If no inline parameter was provided (no '=' or space in same token),
			// check the next argument for a value if it doesn't look like an option.
			consumedNext := false
			if ret.param == "" && i+1 < len(args) && !isOption(args[i+1]) {
				ret.param = args[i+1]
				consumedNext = true
			}

			if ret.param != "" {
				if option.param != "" {
					// OK, we've got a parameter and we expect one
					set(option, ret.negate, ret.param)
				} else {
					// we've got a parameter but didn't expect one,
					// so let's push it onto the extras
					op.Extra = append(op.Extra, ret.param)
					set(option, ret.negate, "")
				}
			} else {
				// no parameter found
				if option.param != "" {
					// parameter expected
					if !option.optional {
						return fmt.Errorf("missing required parameter for %q", args[i])
					}
				}
				set(option, ret.negate, "")
			}

			if consumedNext {
				i += 2
			} else {
				i++
			}
			continue

		default:
			// not an option, we push it onto the extra array
			op.Extra = append(op.Extra, args[i])
		}
		i++
	}

	if op.showHelp {
		op.Help()
		return ErrHelp
	}
	return nil
}

// Parse takes the command line arguments as found in os.Args and interprets
// them. If it finds an unknown option or a missing mandatory argument, it
// returns an error.
func (op *OptionParser) Parse() error {
	return op.ParseFrom(os.Args)
}

// Help prints help text generated from the "On" commands to op.Out.
func (op *OptionParser) Help() {
	if op.Out == nil {
		op.Out = os.Stdout
	}
	fmt.Fprintln(op.Out, op.Banner)
	wd := op.Stop - op.Start
	for _, o := range op.options {
		short := o.short
		long := o.long
		if o.boolParameter && o.long != "" {
			long = "[no-]" + o.long
		}
		if o.long != "" {
			if o.param != "" {
				if o.optional {
					long = fmt.Sprintf("%s[=%s]", o.long, o.param)
				} else {
					long = fmt.Sprintf("%s=%s", o.long, o.param)
				}
			}
		} else { // short only
			if o.param != "" {
				if o.optional {
					short = fmt.Sprintf("%s[=%s]", o.short, o.param)
				} else {
					short = fmt.Sprintf("%s=%s", o.short, o.param)
				}
			}
		}
		dashShort := "-"
		dashLong := "--"
		comma := ","
		if short == "" {
			dashShort = ""
			comma = ""
		}
		if long == "" {
			dashLong = ""
			comma = ""
		}
		lines := wordwrap(o.helptext, wd)
		op.formatAndOutput(op.Start, op.Stop, dashShort, short, comma, dashLong, long, lines)
	}
	if len(op.commands) > 0 {
		fmt.Fprintln(op.Out, "\nCommands")
		for _, cmd := range op.commands {
			lines := wordwrap(cmd.helptext, wd)
			op.formatAndOutput(op.Start, op.Stop, "", "", "", "", cmd.name, lines)
		}
	}
	if op.Coda != "" {
		fmt.Fprintln(op.Out, op.Coda)
	}
}

// NewOptionParser initializes the OptionParser struct with sane settings for
// Banner, Start and Stop and adds a "-h", "--help" option for convenience.
func NewOptionParser() *OptionParser {
	a := &OptionParser{
		Extra:  []string{},
		Banner: " Usage: [parameter] command",
		Start:  30,
		Stop:   79,
		short:  map[string]*allowedOptions{},
		long:   map[string]*allowedOptions{},
		Out:    os.Stdout,
	}
	// Register a help option that prints help and returns ErrHelp from Parse()/ParseFrom()
	a.On("-h", "--help", "Show this help", func() { a.showHelp = true })
	return a
}
