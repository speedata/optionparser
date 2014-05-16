optionparser
============

Ruby (OptionParser) like command line arguments processor


Example usage
-------------

    package main

    import (
        "fmt"
        "github.com/speedata/optionparser"
        "log"
    )

    func myfunc() {
        fmt.Println("myfunc called")
    }

    func main() {
        var somestring string
        var truefalse bool
        options := make(map[string]string)

        op := optionparser.NewOptionParser()
        op.On("-a", "--func", "call myfunc", myfunc)
        op.On("--bstring FOO", "set string to FOO", &somestring)
        op.On("-c", "set boolean option (try -no-c)", options)
        op.On("-d VAL", "set option", options)
        op.On("-e", "boolean option", &truefalse)
        op.Command("y", "Run command y")
        op.Command("z", "Run command z")

        err := op.Parse()
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("string `somestring' is now %q\n", somestring)
        fmt.Printf("options %v\n", options)
        fmt.Printf("-e %v\n", truefalse)
        fmt.Printf("Extra: %#v\n", op.Extra)
    }

and the output of `go run main.go -a --bstring foo -c -d somevalue -e  y z`

is:

    myfunc called
    string `somestring' is now "foo"
    options map[c:true d:somevalue]
    -e true
    Extra: []string{"y", "z"}

