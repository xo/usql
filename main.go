package main

//go:generate ./gen-license.sh

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/mattn/go-isatty"

	"github.com/knq/usql/drivers"
	"github.com/knq/usql/handler"
)

func main() {
	// circumvent all logic to determine if usql was built with support for a specific driver
	if len(os.Args) == 2 &&
		strings.HasPrefix(os.Args[1], "--has-") &&
		strings.HasSuffix(os.Args[1], "-support") {

		n := os.Args[1][6 : len(os.Args[1])-8]
		if v, ok := drivers.KnownDrivers[n]; ok {
			n = v
		}

		var out int
		if _, ok := drivers.Drivers[n]; ok {
			out = 1
		}
		fmt.Fprintf(os.Stdout, "%d", out)
		return
	}

	var err error

	// load current user
	cur, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// parse args
	args := &Args{
		Username:    cur.Username,
		HistoryFile: filepath.Join(cur.HomeDir, ".usql_history"),
	}
	arg.MustParse(args)

	// run
	err = run(args, cur.HomeDir)
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)

		// extra output for when the oracle driver is not available
		if e, ok := err.(*handler.Error); ok && e.Err == handler.ErrDriverNotAvailable && e.Driver == "ora" {
			fmt.Fprint(os.Stderr, "\ntry:\n\n  go get -u -tags oracle github.com/knq/usql\n\n")
		}
		os.Exit(1)
	}
}

// run processes args, processing args.Commands if non-empty, or args.File if
// specified, otherwise launch an interactive readline from stdin.
func run(args *Args, homedir string) error {
	var err error

	// get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// determine interactive/cygwin
	interactive := isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd())
	cygwin := isatty.IsCygwinTerminal(os.Stdout.Fd()) && isatty.IsCygwinTerminal(os.Stdin.Fd())

	// create handler
	h, err := handler.New(args.HistoryFile, homedir, wd, interactive || cygwin, cygwin)
	if err != nil {
		return err
	}

	// open dsn
	err = h.Open(args.DSN)
	if err != nil {
		return err
	}

	// short circuit if commands provided as args
	if len(args.Commands) > 0 {
		return h.RunCommands(args.Commands)
	}

	return h.RunReadline(args.File, args.Out)
}
