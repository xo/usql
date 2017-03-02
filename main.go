package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	isatty "github.com/mattn/go-isatty"
)

func main() {
	// parse args
	args := &Args{
		UserHistoryPrefix: ".usql_history_",
	}
	arg.MustParse(args)

	var err error
	interactive := isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd())
	if interactive {
		err = runInteractive(args)
	} else {

	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)

		// extra output for when the oracle driver is not available
		if err == ErrOracleDriverNotAvailable {
			fmt.Fprint(os.Stderr, "\ntry:\n\n  go get -u -tags oracle github.com/knq/usql\n\n")
		}

		os.Exit(1)
	}
}

func runInteractive(args *Args) error {
	h := &Handler{args: args}
	if args.DSN != "" {
		err := h.Open(args.DSN)
		if err != nil {
			return err
		}
	}
	defer h.Close()

	// run
	return h.Run()
}
