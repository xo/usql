package drivers_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/xo/dburl"
	"github.com/xo/rline"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/feed"
	"github.com/xo/usql/handler"
	_ "github.com/xo/usql/internal"
)

// update rewrites the recordings of TestDescribe instead of comparing against
// them.
//
// Read the diff before committing a rewrite. A recording holds what usql
// prints today, which is not yet what psql prints (W46), so rewriting one
// accepts whatever usql prints now.
var update = flag.Bool("update", false, "rewrite the recordings in testdata/dbrun")

// testServer is the database that the environment names.
type testServer struct {
	// name is the dbrun name of the database, such as postgres-18. It names
	// the recording.
	name string
	// raw is the URL as usql takes it, with the password.
	raw string
	// u is raw, parsed.
	u *dburl.URL
	// secrets are the parts of the URL that a recording or a log must not
	// show: the password, and the password as the URL escapes it.
	secrets []string
}

// The tests in this file run against a database that dbmeta's dbrun starts.
// No test here starts a database. Each takes the URL of the database from the
// variable serverEnv, as dbrun dsn --json prints it in the field url, and
// skips when it is not set. So go test passes on a machine with no database.
// nameEnv holds the dbrun name of the database, such as postgres-18.
// CONTRIBUTING.md says how to run them.
const (
	serverEnv = "USQL_TEST_URL"
	nameEnv   = "USQL_TEST_NAME"
)

// openServer reads the database from the environment, and skips the test
// when there is none. It connects, and returns the connection.
func openServer(t *testing.T) (*testServer, *sql.DB) {
	t.Helper()
	raw := os.Getenv(serverEnv)
	if raw == "" {
		t.Skipf("%s is not set. CONTRIBUTING.md says how to run the tests against a database", serverEnv)
	}
	name := os.Getenv(nameEnv)
	if name == "" {
		t.Fatalf("%s is set and %s is not. Set it to the dbrun name of the database, such as postgres-18", serverEnv, nameEnv)
	}
	if strings.ContainsAny(name, `/\`) || name == "." || name == ".." {
		t.Fatalf("%s is %q, which is not a dbrun name", nameEnv, name)
	}
	s := &testServer{name: name}
	var err error
	if s.raw, err = examplesURL(raw); err != nil {
		t.Fatalf("%s is not a URL: %v", serverEnv, err)
	}
	if s.u, err = dburl.Parse(s.raw); err != nil {
		// The error of dburl can quote the URL, which holds the password.
		t.Fatalf("%s is not a URL that usql opens. Set it to the field url that dbrun dsn --json prints", serverEnv)
	}
	if s.u.User != nil {
		if pass, ok := s.u.User.Password(); ok && pass != "" {
			s.secrets = append(s.secrets, url.QueryEscape(pass), url.PathEscape(pass), pass)
		}
	}
	if !drivers.Registered(s.u.SchemeName) {
		t.Skipf("this build of usql has no driver %s", s.u.SchemeName)
	}
	ctx := context.Background()
	db, err := drivers.Open(ctx, s.u, nil, nil)
	if err != nil {
		t.Fatalf("opening %s: %s", s.name, s.hide(err.Error()))
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := drivers.Ping(ctx, s.u, db); err != nil {
		t.Fatalf("connecting to %s: %s", s.name, s.hide(err.Error()))
	}
	return s, db
}

// openExamples is openServer, and it also checks that the fixture examples
// of dbrun is there: the tables authors and books. dbrun has the fixture for
// some products only, so the test skips on a database that does not hold it.
// A database that has a recording held it once, so there the test fails.
func openExamples(t *testing.T) (*testServer, *sql.DB) {
	t.Helper()
	s, db := openServer(t)
	var n int
	if err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM books").Scan(&n); err != nil {
		msg := fmt.Sprintf("%s has no table books. dbrun fixture examples %s loads it, if dbrun has the fixture for the product", s.name, s.name)
		if _, err := os.Stat(s.recording()); err == nil {
			t.Fatal(msg)
		}
		t.Skip(msg)
	}
	if n != 4 {
		t.Fatalf("%s has %d books, and the fixture examples has 4. Load it again with: dbrun fixture examples %s", s.name, n, s.name)
	}
	return s, db
}

// TestConnect connects to the database of the environment, and reads its
// version as the banner of usql does. It needs no fixture, so it runs on
// every database that dbrun starts.
func TestConnect(t *testing.T) {
	s, db := openServer(t)
	ver, err := drivers.Version(context.Background(), s.u, db)
	if err != nil {
		t.Fatalf("reading the version of %s: %s", s.name, s.hide(err.Error()))
	}
	t.Logf("%s: %s", s.name, s.hide(ver))
}

// examplesURL returns the URL of the database that holds the fixture
// examples. On MySQL and MariaDB, dbrun puts the fixture in the database
// examples, and its URL names no database, so the path becomes /examples.
func examplesURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		// The error of net/url quotes the URL, which holds the password.
		return "", errors.New("it does not parse")
	}
	if strings.EqualFold(u.Scheme, "mysql") && (u.Path == "" || u.Path == "/") {
		u.Path = "/examples"
	}
	return u.String(), nil
}

// recording is the path of the recording of TestDescribe for the database.
func (s *testServer) recording() string {
	return filepath.Join("testdata", "dbrun", s.name+".txt")
}

// hide replaces each secret in s.
func (s *testServer) hide(str string) string {
	for _, secret := range s.secrets {
		str = strings.ReplaceAll(str, secret, "PASSWORD")
	}
	return str
}

// run runs script in a handler of usql that is connected to dsn, or to
// nothing when dsn is empty, and returns what it wrote to its standard output
// and to its standard error, in the order it wrote them.
func (s *testServer) run(t *testing.T, dsn, script string) string {
	t.Helper()
	// one buffer for both, so that an error shows where it happened
	out := new(bytes.Buffer)
	l, err := rline.New(
		rline.WithStdin(strings.NewReader("")),
		rline.WithStdout(out),
		rline.WithStderr(out),
		rline.WithColor(false),
	)
	if err != nil {
		t.Fatalf("making the line editor: %v", err)
	}
	defer l.Close()
	u := &user.User{Username: "usql", HomeDir: t.TempDir()}
	h, err := handler.New(l, feed.NewLines(strings.NewReader(script)), u, t.TempDir(), memfs.New(), true)
	if err != nil {
		t.Fatalf("making the handler: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := h.Open(ctx, dsn); err != nil {
		t.Fatalf("opening the database: %s", s.hide(err.Error()))
	}
	defer h.Close()
	// An error of a statement or a command is in the output already, and the
	// recording shows it. Run returns the last one.
	_ = h.Run()
	return out.String()
}

// describeScript is what TestDescribe runs. Each command names the tables of
// the fixture, because a server of dbrun can hold the objects of other tests,
// which a bare \dt or \di would list.
var describeScript = []string{
	`\dt authors`,
	`\dt books`,
	`\d authors`,
	`\d books`,
	`\di *authors*`,
	`\di *books*`,
	`SELECT book_id, author_id, title FROM books ORDER BY book_id;`,
}

// randomNameRE matches the names that SQL Server makes up for a constraint
// that the statement did not name, such as PK__books__3213E83F1F4C7A2B and
// FK__books__author_id__3296789C. The hexadecimal end differs on each load of
// the fixture.
var randomNameRE = regexp.MustCompile(`\b((?:PK|FK|UQ|DF|CK)__[A-Za-z0-9_]+?__)([0-9A-F]{8}(?:[0-9A-F]{8})?)\b`)

// scrub replaces what changes between two runs against the same release. A
// name keeps its length, so that the columns of a table stay aligned.
func scrub(s string) string {
	return randomNameRE.ReplaceAllStringFunc(s, func(m string) string {
		sub := randomNameRE.FindStringSubmatch(m)
		return sub[1] + strings.Repeat("X", len(sub[2]))
	})
}

// TestDescribe runs the describe commands against the database of the
// environment, and compares what usql prints with the recording of that
// database in testdata/dbrun.
//
// A database with no recording still runs every command, so that a panic or
// a hang shows up, and the test skips the comparison. Run it with -update to
// write the recording.
func TestDescribe(t *testing.T) {
	s, _ := openExamples(t)
	var b strings.Builder
	for _, cmd := range describeScript {
		fmt.Fprintf(&b, "-- %s\n", cmd)
		got := s.run(t, s.raw, cmd+"\n")
		b.WriteString(got)
		if !strings.HasSuffix(got, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	got := recordingHeader(s.name) + scrub(s.hide(b.String()))
	path := s.recording()
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", path)
		return
	}
	want, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		t.Skipf("%s has no recording. The commands ran, and nothing was compared. Write it with: go test -run TestDescribe ./drivers -update", s.name)
	case err != nil:
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("the output differs from %s. Read the diff, and run go test -run TestDescribe ./drivers -update if the change is right.\n%s",
			path, lineDiff(string(want), got))
	}
}

// recordingHeader is the first lines of a recording.
func recordingHeader(name string) string {
	return "# What usql prints for the describe commands on " + name + ", with dbrun's\n" +
		"# fixture examples. TestDescribe in drivers/dbrun_test.go writes this file.\n" +
		"# Do not edit it. Run the test with -update, and read the diff.\n\n"
}

// lineDiff lists the lines of want and got that differ, with their line
// numbers. It is enough to find the change, and git diff shows the rest after
// an -update.
func lineDiff(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	var b strings.Builder
	for i := range max(len(w), len(g)) {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			fmt.Fprintf(&b, "line %d:\n  want: %q\n  got:  %q\n", i+1, wl, gl)
		}
	}
	return b.String()
}

// book is a row of the table books of the fixture examples.
type book struct {
	id, author int
	title      string
}

// readBooks reads the rows of a table that has the columns of books.
func readBooks(t *testing.T, db *sql.DB, table string) []book {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT book_id, author_id, title FROM "+table+" ORDER BY book_id")
	if err != nil {
		t.Fatalf("reading %s: %v", table, err)
	}
	defer rows.Close()
	var out []book
	for rows.Next() {
		var b book
		if err := rows.Scan(&b.id, &b.author, &b.title); err != nil {
			t.Fatalf("reading %s: %v", table, err)
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("reading %s: %v", table, err)
	}
	return out
}

// copyTable is the table that TestCopy makes on each side. CREATE TABLE in
// this form works on every database that dbrun's fixture examples supports.
const (
	copyTable  = "usql_copy_books"
	copyCreate = "CREATE TABLE " + copyTable + " (book_id INTEGER, author_id INTEGER, title VARCHAR(200))"
)

// TestCopy copies the table books with \copy, from the database of the
// environment to a new SQLite database, and from there back into a new table
// of the first database. Each copy must arrive whole.
func TestCopy(t *testing.T) {
	s, db := openExamples(t)
	if !drivers.Registered("sqlite3") {
		t.Skip("this build of usql has no driver sqlite3, which the other side of the copy needs")
	}
	ctx := context.Background()
	want := readBooks(t, db, "books")

	// the other side of the copy
	other := "sqlite3:" + filepath.ToSlash(filepath.Join(t.TempDir(), "copy.db"))
	ou, err := dburl.Parse(other)
	if err != nil {
		t.Fatal(err)
	}
	odb, err := drivers.Open(ctx, ou, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = odb.Close() })
	if _, err := odb.ExecContext(ctx, copyCreate); err != nil {
		t.Fatal(err)
	}

	// from the database to SQLite
	out := s.run(t, "", copyCommand(s.raw, other, "SELECT book_id, author_id, title FROM books", copyTable))
	if got := strings.TrimSpace(s.hide(out)); got != "COPY 4" {
		t.Fatalf("copying from %s to SQLite printed %q, want COPY 4", s.name, got)
	}
	if got := readBooks(t, odb, copyTable); !slices.Equal(got, want) {
		t.Fatalf("SQLite holds %v after the copy, want %v", got, want)
	}

	// and back, into a new table of the database
	_, _ = db.ExecContext(ctx, "DROP TABLE "+copyTable)
	if _, err := db.ExecContext(ctx, copyCreate); err != nil {
		t.Fatalf("making %s on %s: %s", copyTable, s.name, s.hide(err.Error()))
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DROP TABLE "+copyTable) })
	out = s.run(t, "", copyCommand(other, s.raw, "SELECT book_id, author_id, title FROM "+copyTable, copyTable))
	if got := strings.TrimSpace(s.hide(out)); got != "COPY 4" {
		t.Fatalf("copying from SQLite to %s printed %q, want COPY 4", s.name, got)
	}
	if got := readBooks(t, db, copyTable); !slices.Equal(got, want) {
		t.Fatalf("%s holds %v after the copy, want %v", s.name, got, want)
	}
}

// copyCommand is a \copy command. Each argument is in single quotes.
func copyCommand(src, dst, query, table string) string {
	q := func(s string) string {
		return "'" + strings.ReplaceAll(s, "'", `\'`) + "'"
	}
	return `\copy ` + q(src) + " " + q(dst) + " " + q(query) + " " + q(table) + "\n"
}
