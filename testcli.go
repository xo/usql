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
	"regexp"
	"time"

	gexpect "github.com/google/goexpect"
)

func main() {
	binpath := flag.String("binpath", "./usql", "bin path")
	deadline := flag.Duration("deadline", 5*time.Minute, "total execution deadline")
	timeout := flag.Duration("timeout", 2*time.Minute, "individual test timeout")
	re := flag.String("run", "", "test name regexp to run")
	flag.Parse()
	if err := run(context.Background(), *binpath, *deadline, *timeout, *re); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, binpath string, deadline, timeout time.Duration, re string) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(deadline))
	defer cancel()
	tests, err := cliTests()
	if err != nil {
		return err
	}
	var nameRE *regexp.Regexp
	if re != "" {
		nameRE, err = regexp.Compile(re)
		if err != nil {
			return err
		}
	}
	for _, test := range tests {
		if nameRE != nil && !nameRE.MatchString(test.name) {
			log.Printf(">>> SKIPPING: %s", test.name)
			continue
		}
		log.Printf(">>> RUNNING: %s", test.name)
		if err := test.do(ctx, binpath, timeout); err != nil {
			return fmt.Errorf("test %s: %v", test.name, err)
		}
		log.Printf(">>> COMPLETED: %s", test.name)
	}
	return nil
}

type Test struct {
	name   string
	script string
	args   []string
	env    []string
}

func cliTests() ([]Test, error) {
	env := append(os.Environ(), "TERM=xterm-256color")
	return []Test{
		{
			"complex/postgres",
			"./contrib/postgres/test.sql",
			[]string{"pgsql://postgres:P4ssw0rd@localhost", "--pset=pager=off"},
			env,
		},
		{
			"complex/mysql",
			"./contrib/mysql/test.sql",
			[]string{"my://root:P4ssw0rd@localhost", "--pset=pager=off"},
			env,
		},
		{
			"complex/sqlite3",
			"./contrib/sqlite3/test.sql",
			[]string{"sqlite:./testdata/sqlite3_test.db", "--pset=pager=off"},
			env,
		},
		{
			"complex/moderncsqlite",
			"./contrib/sqlite3/test.sql",
			[]string{"mq:./testdata/moderncsqlite_test.db", "--pset=pager=off"},
			env,
		},
		{
			"complex/sqlserver",
			"./contrib/sqlserver/test.sql",
			[]string{"sqlserver://sa:Adm1nP@ssw0rd@localhost/", "--pset=pager=off"},
			env,
		},
		{
			"complex/cassandra",
			"./contrib/cassandra/test.sql",
			[]string{"ca://cassandra:cassandra@localhost", "--pset=pager=off"},
			env,
		},
		{
			"copy/a_bit_of_everything",
			"./testdata/copy.sql",
			[]string{"--pset=pager=off"},
			env,
		},
	}, nil
}

func (test Test) do(ctx context.Context, binpath string, timeout time.Duration) error {
	exp, errch, err := gexpect.SpawnWithArgs(
		append([]string{binpath}, test.args...),
		timeout,
		gexpect.SetEnv(test.env),
		gexpect.Tee(&noopWriteCloser{os.Stdout}),
	)
	if err != nil {
		return err
	}
	buf, err := ioutil.ReadFile(test.script)
	if err != nil {
		return err
	}
	for _, line := range bytes.Split(buf, []byte{'\n'}) {
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
