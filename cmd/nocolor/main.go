package main

import (
	"errors"
	"io"
	"os"

	"github.com/wkhere/nocolor"
)

type action struct {
	help func()
}

func parseArgs(args []string) (a action, _ error) {
	const usage = `Usage: nocolor

Copies stdin to stdout line by line, removing ansi color sequences in each line.
Errors on binary data (most of it).
`
	rest := make([]string, 0, len(args))

	for ; len(args) > 0; args = args[1:] {
		switch arg := args[0]; {

		case arg == "-h" || arg == "--help":
			a.help = func() { io.WriteString(os.Stdout, usage) }
			return a, nil

		default:
			rest = append(rest, arg)
		}
	}

	if len(rest) > 0 {
		return a, errors.New("expected no args")
	}
	return a, nil
}

func main() {
	a, err := parseArgs(os.Args[1:])
	if err != nil {
		die(2, err)
	}
	if a.help != nil {
		a.help()
		os.Exit(0)
	}

	err = nocolor.Run(os.Stdin, os.Stdout)
	if err != nil {
		die(1, err)
	}
}

func die(code int, err error) {
	io.WriteString(os.Stderr, err.Error())
	os.Exit(code)
}
