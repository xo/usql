package main

//go:generate ./gen-license.sh
import (
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

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
	// load current user
	cur, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	args := NewArgs()
	// run
	err = run(args, cur)
	if err != nil && err != io.EOF && err != rline.ErrInterrupt {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if e, ok := err.(*drivers.Error); ok && e.Err == text.ErrDriverNotAvailable {
			m := make(map[string]string, len(known))
			for k, v := range known {
				m[v] = k
			}
			tag := e.Driver
			if t, ok := m[tag]; ok {
				tag = t
			}
			fmt.Fprintf(os.Stderr, "\ntry:\n\n  go get -u -tags %s github.com/xo/usql\n\n", tag)
		}
		os.Exit(1)
	}
}

// run processes args, processing args.CommandOrFiles if non-empty, if
// specified, otherwise launch an interactive readline from stdin.
func run(args *Args, u *user.User) error {
	// get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	// handle variables
	for _, v := range args.Variables {
		if i := strings.Index(v, "="); i != -1 {
			env.Set(v[:i], v[i+1:])
		} else {
			env.Unset(v)
		}
	}
	for _, v := range args.PVariables {
		if i := strings.Index(v, "="); i != -1 {
			if _, err = env.Pset(v[:i], v[i+1:]); err != nil {
				return err
			}
		} else {
			if _, err = env.Ptoggle(v, ""); err != nil {
				return err
			}
		}
	}
	// create input/output
	l, err := rline.New(len(args.CommandOrFiles) != 0, args.Out, env.HistoryFile(u))
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
	if err = h.Open(dsn); err != nil {
		return err
	}
	// start transaction
	if args.SingleTransaction {
		if h.IO().Interactive() {
			return text.ErrSingleTransactionCannotBeUsedWithInteractiveMode
		}
		if err = h.Begin(); err != nil {
			return err
		}
	}
	// rc file
	if rc := env.RCFile(u); !args.NoRC && rc != "" {
		if err = h.Include(rc, false); err != nil && err != text.ErrNoSuchFileOrDirectory {
			return err
		}
	}
	// setup runner
	f := h.Run
	if len(args.CommandOrFiles) != 0 {
		f = runCommandOrFiles(h, args.CommandOrFiles)
	}
	// run
	if err = f(); err != nil {
		return err
	}
	// commit
	if args.SingleTransaction {
		return h.Commit()
	}
	return nil
}

// runCommandOrFiles proccesses all the supplied commands or files.
func runCommandOrFiles(h *handler.Handler, commandsOrFiles []CommandOrFile) func() error {
	return func() error {
		for _, x := range commandsOrFiles {
			h.SetSingleLineMode(x.Command)
			if x.Command {
				h.Reset([]rune(x.Value))
				if err := h.Run(); err != nil && err != io.EOF {
					return err
				}
			} else {
				if err := h.Include(x.Value, false); err != nil {
					return err
				}
			}
		}
		return nil
	}
}
