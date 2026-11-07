package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/xo/dbfwd"
	"github.com/xo/dburl"
	"github.com/xo/rline"
	_ "github.com/xo/usql/drivers/moderncsqlite"
	"github.com/xo/usql/feed"
	"github.com/xo/usql/metacmd"
	"github.com/xo/usql/text"
)

// fakeBackend stands in for the k8s backend of dbfwd. Each forward it opens
// leads to the same SQLite file.
type fakeBackend struct {
	mu      sync.Mutex
	url     string
	tunnels []*fakeTunnel
}

// backend is registered once, because dbfwd panics on a second registration
// of the same scheme. The handler package does not import the k8s backend,
// so the scheme is free.
var backend = func() *fakeBackend {
	b := new(fakeBackend)
	dbfwd.Register("k8s", b)
	return b
}()

func (b *fakeBackend) Lookup(context.Context, *dburl.URL, dbfwd.Config) (*dbfwd.Match, error) {
	return nil, dbfwd.ErrUnsupported
}

func (b *fakeBackend) Open(context.Context, *dburl.URL, dbfwd.Config) (*dbfwd.Opened, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	t := &fakeTunnel{done: make(chan struct{})}
	b.tunnels = append(b.tunnels, t)
	return &dbfwd.Opened{
		URL:      b.url,
		User:     "app",
		Database: "app",
		Target: dbfwd.Target{
			Backend: "k8s",
			Scope:   []dbfwd.Field{{Name: "context", Value: "kind-test"}, {Name: "namespace", Value: "prod"}},
			Kind:    "Service",
			Name:    "pg-rw",
			Owner:   "CloudNativePG Cluster pg",
			Type:    "postgres",
		},
		Tunnel: t,
	}, nil
}

// opened returns the tunnels that the backend opened.
func (b *fakeBackend) opened() []*fakeTunnel {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]*fakeTunnel(nil), b.tunnels...)
}

// fakeTunnel is a forward that stops when it is closed or lost.
type fakeTunnel struct {
	done chan struct{}
	once sync.Once
	err  error
}

func (t *fakeTunnel) Done() <-chan struct{} { return t.done }

func (t *fakeTunnel) Err() error { return t.err }

func (t *fakeTunnel) Close() error {
	t.once.Do(func() { close(t.done) })
	return nil
}

// lose stops the tunnel as kubectl does when it exits on its own.
func (t *fakeTunnel) lose() {
	t.err = errors.New("kubectl exited")
	_ = t.Close()
}

// stopped reports whether the tunnel stopped.
func (t *fakeTunnel) stopped() bool {
	select {
	case <-t.done:
		return true
	default:
		return false
	}
}

// TestForward opens a k8s URL through a fake backend of dbfwd. The handler
// connects to the local URL, shows the forward in the prompt, stops statements
// once the forward is lost, opens it again with \c, and stops it on close.
func TestForward(t *testing.T) {
	file := filepath.Join(t.TempDir(), "app.db")
	backend.url = "moderncsqlite:" + file
	l, err := rline.New(
		rline.WithStdin(strings.NewReader("")),
		rline.WithStdout(io.Discard),
		rline.WithStderr(io.Discard),
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
	ctx := context.Background()
	query := func() error {
		return h.Execute(ctx, new(bytes.Buffer), metacmd.Option{}, "select 1")
	}
	// open the forward
	if err := h.Open(ctx, "k://pg.prod/app"); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if got := h.URL().SchemeName; got != "moderncsqlite" {
		t.Errorf("scheme of the local URL = %q, want moderncsqlite", got)
	}
	fwdURL, fwd := h.Forward()
	if fwd == nil || fwdURL.Hostname() != "pg.prod" {
		t.Fatalf("Forward = %v, %v, want the forward of pg.prod", fwdURL, fwd)
	}
	if got, want := fwd.String(), "Service pg-rw in context kind-test, namespace prod, owned by CloudNativePG Cluster pg"; got != want {
		t.Errorf("forward = %q, want %q", got, want)
	}
	if got, want := h.render("%M %m [%>]", false), "pg.prod pg []"; got != want {
		t.Errorf("prompt = %q, want %q", got, want)
	}
	if err := query(); err != nil {
		t.Fatalf("query through the forward: %v", err)
	}
	// a password prompt does not apply to a forward
	if _, err := h.Password("k://pg.prod/app"); !errors.Is(err, text.ErrPasswordForward) {
		t.Errorf("Password = %v, want %v", err, text.ErrPasswordForward)
	}
	// a lost forward stops the statements
	first := backend.opened()[0]
	first.lose()
	if err := query(); !errors.Is(err, dbfwd.ErrForwardLost) {
		t.Fatalf("query after the forward was lost = %v, want %v", err, dbfwd.ErrForwardLost)
	}
	// \c with no argument starts the forward again
	if err := h.Open(ctx); err != nil {
		t.Fatalf("Open with no argument: %v", err)
	}
	tunnels := backend.opened()
	if len(tunnels) != 2 {
		t.Fatalf("forwards opened = %d, want 2", len(tunnels))
	}
	if err := query(); err != nil {
		t.Fatalf("query through the new forward: %v", err)
	}
	// a file that \i reads keeps the forward, and its \c starts it again
	if err := h.IncludeReader(strings.NewReader("\\c\n"), "reconnect.sql"); err != nil {
		t.Fatalf("IncludeReader: %v", err)
	}
	if tunnels = backend.opened(); len(tunnels) != 3 {
		t.Fatalf("forwards opened after \\c in a file = %d, want 3", len(tunnels))
	}
	if !tunnels[1].stopped() {
		t.Error("the forward that \\c in a file replaced still runs")
	}
	if _, fwd := h.Forward(); fwd == nil {
		t.Fatal("Forward = nil after the file ran, want the forward")
	}
	// another URL replaces the connection, and stops the forward
	if err := h.Open(ctx, "moderncsqlite:"+file); err != nil {
		t.Fatalf("Open without a forward: %v", err)
	}
	if !tunnels[2].stopped() {
		t.Error("the forward runs after another URL replaced it")
	}
	if _, fwd := h.Forward(); fwd != nil {
		t.Errorf("Forward = %v, want nil", fwd)
	}
	// \c with no argument does nothing when there is no forward
	if err := h.Open(ctx); err != nil {
		t.Fatalf("Open with no argument: %v", err)
	}
	if n := len(backend.opened()); n != 3 {
		t.Errorf("forwards opened = %d, want 3", n)
	}
	// close stops the forward
	if err := h.Open(ctx, "k://pg.prod/app"); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	tunnels = backend.opened()
	if last := tunnels[len(tunnels)-1]; !last.stopped() {
		t.Error("the forward runs after Close")
	}
}
