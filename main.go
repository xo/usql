package main

import (
	"fmt"
	"os"

	"github.com/kenshaw/go-arg"
	"github.com/mattn/go-isatty"
)

func main() {
	// parse args
	args := &Args{
		UserHistoryPrefix: ".usql_history_",
	}
	arg.MustParse(args)

	h := &Handler{
		args:        args,
		interactive: isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd()),
	}

	// run
	err := h.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)

		// extra output for when the oracle driver is not available
		if err == ErrOracleDriverNotAvailable {
			fmt.Fprint(os.Stderr, "\ntry:\n\n  go get -u -tags oracle github.com/knq/usql\n\n")
		}

		os.Exit(1)
	}
}
