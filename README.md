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
 * `function`: Call the function. The function must have the signature `func()`.
 * `map[string]string`: Set an entry of the map to the value of the given parameter and the key of the argument.
 * `[]string` Set the slice values to a comma separated list.

Help usage
----------

The options `-h` and `--help` are included by default. The example below output this on `cmd -h`:

    Usage: [parameter] command
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

Settings
--------

After calling  `op := optionparser.NewOptionParser()` you can set `op.Banner` to set the first line of the help output. The default value is `Usage: [parameter] command`.

To control the first and last column of the help output, set `op.Start` and `op.Stop`. The default values are the integer values of 30 and 79.


Example usage
-------------

```go
package main

import (
	"fmt"
	"log"

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

	err := op.Parse()
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

and the output of `go run main.go -a --bstring foo -c -d somevalue -e x -f -g a,b,c y z`

is:

    myfunc called
    string `somestring' is now "foo"
    options map[c:true dlong:somevalue elong:x]
    -f true
    -g [a b c]
    Extra: []string{"y", "z"}



**State**: Actively maintained, and used in production. Without warranty, of course.<br>
**Maturity level**: 5/5 (works well in all tested repositories, there will be no API change)<br>
**License**: Free software (MIT License)<br>
**Installation**: Just run `go get github.com/speedata/optionparser`<br>
**API documentation**: https://pkg.go.dev/github.com/speedata/optionparser<br>
**Contact**: <gundlach@speedata.de>, [@speedata@typo.social](https://typo.social/@speedata)<br>
**Repository**: https://github.com/speedata/optionparser<br>
**Dependencies**: None<br>
**Contribution**: We like to get any kind of feedback (success stories, bug reports, merge requests, ...)

