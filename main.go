package main

//go:generate ./gen-license.sh

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

	"github.com/alexflint/go-arg"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
	"github.com/xo/usql/handler"
	"github.com/xo/usql/internal"
	"github.com/xo/usql/rline"
	"github.com/xo/usql/text"
)

func main() {
	// get available drivers and known build tags
	available, known := drivers.Available(), internal.KnownBuildTags()

	// report if database is supported
	if len(os.Args) == 2 &&
		strings.HasPrefix(os.Args[1], "--has-") &&
		strings.HasSuffix(os.Args[1], "-support") {

		n := os.Args[1][6 : len(os.Args[1])-8]
		if v, ok := known[n]; ok {
			n = v
		}

		var out int
		if _, ok := available[n]; ok {
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
		Username: cur.Username,
	}
	arg.MustParse(args)

	// run
	err = run(args, cur)
	if err != nil && err != io.EOF && err != rline.ErrInterrupt {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)

		if e, ok := err.(*drivers.Error); ok {
			// extra output for when a driver is not available
			if e.Err == text.ErrDriverNotAvailable {
				tag := e.Driver
				if t, ok := known[tag]; ok {
					tag = t
				}

				fmt.Fprint(os.Stderr, "\ntry:\n\n  go get -u -tags "+tag+" github.com/xo/usql\n\n")
			}
		}

		os.Exit(1)
	}
}

// run processes args, processing args.Commands if non-empty, or args.File if
// specified, otherwise launch an interactive readline from stdin.
func run(args *Args, u *user.User) error {
	var err error

	// get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// create input/output
	l, err := rline.New(args.Commands, args.File, args.Out, env.HistoryFile(u))
	if err != nil {
		return err
	}
	defer l.Close()

	// create handler
	h := handler.New(l, u, wd, args.NoPassword)

	// force a password ...
	dsn := args.DSN
	if args.ForcePassword {
		dsn, err = h.Password(dsn)
		if err != nil {
			return err
		}
	}

	// open dsn
	err = h.Open(dsn)
	if err != nil {
		return err
	}

	// start transaction
	if args.SingleTransaction {
		if h.IO().Interactive() {
			return text.ErrSingleTransactionCannotBeUsedWithInteractiveMode
		}
		err = h.Begin()
		if err != nil {
			return err
		}
	}

	// rc file
	if rc := env.RCFile(u); !args.NoRC && rc != "" {
		err = h.Include(rc, false)
		if err != nil && err != text.ErrNoSuchFileOrDirectory {
			return err
		}
	}

	// setup runner
	f := h.Run
	if len(args.Commands) != 0 {
		f = h.CommandRunner(args.Commands)
	}

	// run
	err = f()
	if err != nil {
		return err
	}

	// commit
	if args.SingleTransaction {
		return h.Commit()
	}

	return nil
}
