[![GoDoc](https://pkg.go.dev/badge/github.com/speedata/optionparser)](https://pkg.go.dev/github.com/speedata/optionparser)

optionparser
============

Mature command line arguments processor for Go.

Inspired by Ruby's (OptionParser) command line arguments processor.

Installation
------------

    go get github.com/speedata/optionparser

Usage
-----

    op := optionparser.NewOptionParser()
    op.On(arguments ...interface{})
    ...
    err := op.Parse()

where `arguments` is one of:

 * `"-a"`: a short argument
 * `"--argument"`: a long argument
 * `"--argument [FOO]"` a long argument with an optional parameter
 * `"--argument FOO"` a long argument with a mandatory parameter
 * `"Some text"`: The description text for the command line parameter
 * `&aboolean`: Set the given boolean to `true` if the argument is given, set to `false` if parameter is prefixed with `no-`, such as `--no-foo`.
 * `&astring`: Set the string to the value of the given parameter
 * `function`: Call the function. The function can have signature `func()` or `func(string)` where the string parameter receives the option's value.
 * `map[string]string`: Set an entry of the map to the value of the given parameter and the key of the argument.
 * `[]string` Set the slice values to a comma separated list.

**Note on callback functions**: When using `func(string)` callbacks, the function receives the parameter value as passed by the user. If your callback expects a specific format (e.g., `key=value`), make sure to validate the input and handle errors gracefully to avoid panics. Starting from v1.1.1, the parser will catch panics in callbacks and return them as errors, but it's still recommended to handle validation explicitly in your callback.

Help usage
----------

The options `-h` and `--help` are included by default.
Starting with **v1.1.0**, calling `op.Parse()` or `op.ParseFrom()` with `--help` **no longer exits the program automatically**.
Instead, `Parse()` prints the help text to standard output and returns the sentinel error `optionparser.ErrHelp`.
This makes the package safer and easier to integrate in larger applications or libraries.

To emulate the previous behavior (exit after printing help), you can simply do:

```go
if errors.Is(err, optionparser.ErrHelp) {
    os.Exit(0)
}
```

The example below shows the help output that appears when running `cmd -h`:

```
 Usage: [awesome-options] <awesome-command>
 -h, --help                   Show this help
 -a, --func                   call myfunc
     --bstring=FOO            set string to FOO
 -c                           set boolean option (try -no-c)
 -d, --dlong=VAL              set option
 -e, --elong[=VAL]            set option with optional parameter
 -f                           boolean option

 Commands
       y                      Run command y
       z                      Run command z

 For more information or to contribute, visit https://github.com/speedata/optionparser.
```

Settings
--------

After calling  `op := optionparser.NewOptionParser()` you can set `op.Banner` and `op.Coda`.
`op.Banner` customizes the first line of the help output. The default value is `"Usage: [parameter] command"`.
By default, `op.Coda` is an empty string. Set a value for `op.Coda` if you want to add text at the end of the help output.

To control the first and last column of the help output, set `op.Start` and `op.Stop`.
The default values are `30` and `79`.

Shell completion
----------------

`GenerateCompletion(shell, programName, w)` writes a tab-completion script for
`bash`, `zsh` or `fish` to the given `io.Writer`. The script is generated from
the options and commands already registered on the parser, so a single source
of truth stays in your `op.On(...)` and `op.Command(...)` calls.

A common pattern is to expose this through a hidden flag and let packagers
invoke it at build time:

```go
op.On("--generate-completion SHELL", "Print shell completion script (bash, zsh or fish) to stdout and exit", func(shell string) {
    if err := op.GenerateCompletion(shell, "myprog", os.Stdout); err != nil {
        log.Fatal(err)
    }
    os.Exit(0)
})
```

End users (or your installer) then run, for example:

```
myprog --generate-completion=zsh  > ~/.zsh/completions/_myprog
myprog --generate-completion=bash > /usr/share/bash-completion/completions/myprog
myprog --generate-completion=fish > ~/.config/fish/completions/myprog.fish
```

The generated scripts handle:

- short and long flag forms,
- the `--no-` toggle form for options registered with a `--no-` prefix,
- registered commands as positional completions, and
- file completion for any option that takes a parameter (this is a
  best-effort default because the parser has no knowledge of value
  semantics — a `--port` option will offer file names too).

Example usage
-------------

```go
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/speedata/optionparser"
)

func myfunc() {
	fmt.Println("myfunc called")
}

func main() {
	var somestring string
	var truefalse bool
	options := make(map[string]string)
	stringslice := []string{}

	op := optionparser.NewOptionParser()
	op.On("-a", "--func", "call myfunc", myfunc)
	op.On("--bstring FOO", "set string to FOO", &somestring)
	op.On("-c", "set boolean option (try -no-c)", options)
	op.On("-d", "--dlong VAL", "set option", options)
	op.On("-e", "--elong [VAL]", "set option with optional parameter", options)
	op.On("-f", "boolean option", &truefalse)
	op.On("-g VALUES", "give multiple values", &stringslice)

	op.Command("y", "Run command y")
	op.Command("z", "Run command z")

	op.Banner = "Usage: [awesome-options] <awesome-command>"
	op.Coda = "\nFor more information or to contribute, visit https://github.com/speedata/optionparser."

	err := op.Parse()
	if errors.Is(err, optionparser.ErrHelp) {
		// Exit cleanly after showing help
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("string `somestring' is now %q\n", somestring)
	fmt.Printf("options %v\n", options)
	fmt.Printf("-f %v\n", truefalse)
	fmt.Printf("-g %v\n", stringslice)
	fmt.Printf("Extra: %#v\n", op.Extra)
}
```

The output of `go run main.go -a --bstring foo -c -d somevalue -e x -f -g a,b,c y z` is:

```
myfunc called
string `somestring' is now "foo"
options map[c:true dlong:somevalue elong:x]
-f true
-g [a b c]
Extra: []string{"y", "z"}
```

**State**: Actively maintained, and used in production. Without warranty, of course.<br>
**Maturity level**: 5/5 (works well in all tested repositories, no breaking API changes expected)<br>
**License**: Free software (MIT License)<br>
**Installation**: Just run `go get github.com/speedata/optionparser`<br>
**API documentation**: https://pkg.go.dev/github.com/speedata/optionparser<br>
**Contact**: <gundlach@speedata.de>, [@speedata@typo.social](https://typo.social/@speedata)<br>
**Repository**: https://github.com/speedata/optionparser<br>
**Dependencies**: None<br>
**Contribution**: We like to get any kind of feedback (success stories, bug reports, merge requests, ...)
