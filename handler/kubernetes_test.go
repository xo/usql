package handler

import (
	"bytes"
	"context"
	"fmt"
	"os/user"
	"slices"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/xo/dbfwd"
	"github.com/xo/dburl"
	"github.com/xo/rline"
	"github.com/xo/usql/feed"
)

// listCall is what \k passed to dbfwd.ListURL.
type listCall struct {
	scope string
	where []dbfwd.Field
}

// ListURL keeps what \k passed, and returns a URL of the namespace that the
// fields name, or of prod. dbfwd owns the rules of a scope, and tests them.
func (b *fakeBackend) ListURL(scope string, where []dbfwd.Field) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listed = append(b.listed, listCall{scope: scope, where: where})
	namespace := "prod"
	for _, f := range where {
		if f.Name == "namespace" {
			namespace = f.Value
		}
	}
	return "k://x." + namespace, nil
}

// List lists two databases in the namespace prod, and skips the Services. A
// URL of every namespace is forbidden, and a URL of the namespace empty finds
// nothing.
func (b *fakeBackend) List(_ context.Context, u *dburl.URL, _ dbfwd.Config) (*dbfwd.Listing, error) {
	scope := func(namespace string) []dbfwd.Field {
		return []dbfwd.Field{{Name: "context", Value: "kind-test"}, {Name: "namespace", Value: namespace}}
	}
	switch {
	case strings.HasSuffix(u.Host, ".*"):
		return nil, fmt.Errorf("listing every namespace: %w", dbfwd.ErrForbidden)
	case strings.HasSuffix(u.Host, ".empty"):
		return &dbfwd.Listing{Scope: scope("empty")}, nil
	}
	return &dbfwd.Listing{
		Scope: scope("prod"),
		Entries: []dbfwd.Entry{
			{
				Target:     dbfwd.Target{Backend: "k8s", Scope: scope("prod"), Kind: "Cluster", Name: "pg", Type: "postgres"},
				Operator:   "CloudNativePG",
				URL:        "k://pg.prod/",
				Roles:      []string{"rw", "ro"},
				Ready:      2,
				Pods:       3,
				Databases:  []string{"app"},
				Credential: "Secret pg-app key password",
			},
			{
				Target:     dbfwd.Target{Backend: "k8s", Scope: scope("prod"), Kind: "StatefulSet", Name: "my", Type: "mysql"},
				URL:        "k://my.prod/",
				Ready:      1,
				Pods:       1,
				Credential: "pod spec",
			},
		},
		Skipped: []dbfwd.Skip{{Scope: scope("prod"), Kind: "services", Reason: "not allowed"}},
	}, nil
}

// TestKubernetes runs \k against the fake backend, and checks the URL that
// each form of the command asks for, and what it writes.
func TestKubernetes(t *testing.T) {
	field := func(name, value string) dbfwd.Field { return dbfwd.Field{Name: name, Value: value} }
	tests := []struct {
		cmd    string
		scope  string
		where  []dbfwd.Field
		out    []string
		notOut []string
		stderr string
		err    string
	}{
		{
			cmd:    `\k`,
			out:    []string{"List of Kubernetes databases in context kind-test, namespace prod", "Namespace", "Operator", "CloudNativePG", "k://pg.prod/", "k://my.prod/", "(2 rows)"},
			notOut: []string{"Credential", "Secret pg-app"},
			stderr: "note: skipped services in context kind-test, namespace prod: not allowed",
		},
		{
			cmd:   `\k+ -n prod pg*`,
			scope: "pg*",
			where: []dbfwd.Field{field("namespace", "prod")},
			out:   []string{"Roles", "rw, ro", "2/3", "Databases", "Secret pg-app key password", "pod spec"},
		},
		{
			cmd:   `\k --namespace=prod --context kind-test`,
			where: []dbfwd.Field{field("namespace", "prod"), field("context", "kind-test")},
		},
		{
			cmd:   `\k --context=kind-test cnpg://*`,
			scope: "cnpg://*",
			where: []dbfwd.Field{field("context", "kind-test")},
		},
		{
			cmd:   `\k k://*.prod/`,
			scope: "k://*.prod/",
		},
		{
			cmd:   `\k -n empty`,
			where: []dbfwd.Field{field("namespace", "empty")},
			out:   []string{"Did not find any Kubernetes databases in context kind-test, namespace empty."},
		},
		{cmd: `\k -A`, where: []dbfwd.Field{field("namespace", "*")}, err: `\k without -A lists the current namespace`},
		{cmd: `\k -n`, err: `\k: -n needs a value`},
		{cmd: `\k -A -n prod`, err: `\k: -A lists every namespace, and -n names prod`},
		{cmd: `\k -x`, err: `invalid option "-x"`},
		{cmd: `\k pg* my*`, err: `\k: extra argument "my*"`},
	}
	for _, test := range tests {
		t.Run(test.cmd, func(t *testing.T) {
			stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
			l, err := rline.New(
				rline.WithStdin(strings.NewReader("")),
				rline.WithStdout(stdout),
				rline.WithStderr(stderr),
				rline.WithColor(false),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer l.Close()
			h, err := New(l, feed.NewLines(strings.NewReader("")), &user.User{}, ".", memfs.New(), true)
			if err != nil {
				t.Fatal(err)
			}
			n := len(backend.lists())
			err = h.RunCommand(test.cmd)
			switch {
			case test.err != "" && (err == nil || !strings.Contains(err.Error(), test.err)):
				t.Fatalf("RunCommand = %v, want an error with %q", err, test.err)
			case test.err == "" && err != nil:
				t.Fatalf("RunCommand: %v", err)
			}
			// a command that fails before it reaches dbfwd lists nothing
			lists := backend.lists()
			switch called := len(lists) > n; {
			case !called && test.err != "" && test.where == nil:
			case !called:
				t.Fatal("\\k did not call dbfwd.ListURL")
			case lists[n].scope != test.scope || !slices.Equal(lists[n].where, test.where):
				t.Errorf("ListURL(%q, %v), want ListURL(%q, %v)", lists[n].scope, lists[n].where, test.scope, test.where)
			}
			for _, s := range test.out {
				if !strings.Contains(stdout.String(), s) {
					t.Errorf("stdout lacks %q:\n%s", s, stdout)
				}
			}
			for _, s := range test.notOut {
				if strings.Contains(stdout.String(), s) {
					t.Errorf("stdout has %q:\n%s", s, stdout)
				}
			}
			if !strings.Contains(stderr.String(), test.stderr) {
				t.Errorf("stderr lacks %q:\n%s", test.stderr, stderr)
			}
		})
	}
}
