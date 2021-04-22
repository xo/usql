// +build ignore

// Command testcli runs goexpect tests against a built usql binary.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	gexpect "github.com/google/goexpect"
)

func main() {
	deadline := flag.Duration("deadline", 5*time.Minute, "deadline")
	timeout := flag.Duration("timeout", 2*time.Minute, "timeout")
	flag.Parse()
	if err := run(context.Background(), *deadline, *timeout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, deadline, timeout time.Duration) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(deadline))
	defer cancel()
	tests, err := cliTests()
	if err != nil {
		return err
	}
	for _, test := range tests {
		log.Printf(">>> RUNNING: %s", test.name)
		if err := test.do(ctx, timeout); err != nil {
			return fmt.Errorf("test %s: %v", test.name, err)
		}
		log.Printf(">>> COMPLETED: %s", test.name)
	}
	return nil
}

type Test struct {
	name string
	args []string
	env  []string
	buf  []byte
}

func cliTests() ([]Test, error) {
	env := append(os.Environ(), "TERM=xterm-256color")
	buf, err := ioutil.ReadFile("./contrib/sqlite3/test.sql")
	if err != nil {
		return nil, err
	}
	return []Test{
		{
			"complex sqlite3 test script",
			[]string{"sqlite://test.db"},
			env, buf,
		},
		{
			"complex moderncsqlite test script",
			[]string{"mq://test2.db"},
			env, buf,
		},
	}, nil
}

func (test Test) do(ctx context.Context, timeout time.Duration) error {
	exp, errch, err := gexpect.SpawnWithArgs(
		append([]string{"./usql"}, test.args...),
		timeout,
		gexpect.SetEnv(test.env),
		gexpect.Tee(&noopWriteCloser{os.Stdout}),
	)
	if err != nil {
		return err
	}
	for _, line := range bytes.Split(test.buf, []byte{'\n'}) {
		if err := exp.Send(string(line) + "\n"); err != nil {
			return err
		}
	}
	select {
	case <-ctx.Done():
		defer exp.Close()
		return ctx.Err()
	case err := <-errch:
		defer exp.Close()
		return err
	}
	return exp.Close()
}

type noopWriteCloser struct {
	io.Writer
}

func (*noopWriteCloser) Close() error {
	return nil
}
