package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/kenshaw/go-arg"
	"github.com/mattn/go-isatty"
)

func main() {
	// circumvent all logic to just determine if usql was built with oracle
	// support
	if len(os.Args) == 2 && os.Args[1] == "--has-oracle-support" {
		var out int
		if _, ok := drivers["ora"]; ok {
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

	interactive := (isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd())) ||
		(isatty.IsCygwinTerminal(os.Stdout.Fd()) && isatty.IsCygwinTerminal(os.Stdin.Fd()))
	h := &Handler{
		args:        args,
		interactive: interactive,
	}

	// run
	err = h.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)

		// extra output for when the oracle driver is not available
		if err == ErrOracleDriverNotAvailable {
			fmt.Fprint(os.Stderr, "\ntry:\n\n  go get -u -tags oracle github.com/knq/usql\n\n")
		}

		os.Exit(1)
	}
}
