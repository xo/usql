package drivers_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/xo/dbmeta"
	dbdump "github.com/xo/dbmeta/dump"
	"github.com/xo/dburl"
	"github.com/xo/rline"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/dump"
	"github.com/xo/usql/env"
	"github.com/xo/usql/feed"
	"github.com/xo/usql/handler"
)

// The tests of this file dump a database with usql, load the script into
// another database, and compare. The fixture of each dialect is in
// testdata/dump. Each holds one of each object and type that a dump writes,
// and rows that test the literals.
//
// TestDumpEmbedded needs no server, and runs on every go test. TestDump runs
// against the database of USQL_TEST_URL, as the other tests of dbrun_test.go
// do, and skips when it is not set.

// dumpKey is the key of the \restrict line of every dump of the tests, so
// that two dumps of the same database are the same text.
const dumpKey = "usqldumptest"

// dumpSide is one database of a dump test.
type dumpSide struct {
	// dsn is the URL that usql opens.
	dsn string
	// name is the database in a server, which a dump of MySQL names. It is
	// empty for a file.
	name string
	// hide hides the password in a message.
	hide func(string) string
}

// embedded returns a side that is a new file of an embedded database.
func embedded(t *testing.T, scheme string) dumpSide {
	t.Helper()
	if !drivers.Registered(scheme) {
		t.Skipf("this build of usql has no driver %s", scheme)
	}
	path := filepath.ToSlash(filepath.Join(t.TempDir(), "dump."+scheme))
	return dumpSide{dsn: scheme + ":" + path, hide: func(s string) string { return s }}
}

// newDumpHandler returns a handler of usql that writes to out, and is
// connected to dsn.
func newDumpHandler(t *testing.T, side dumpSide, out *bytes.Buffer) *handler.Handler {
	t.Helper()
	l, err := rline.New(
		rline.WithStdin(strings.NewReader("")),
		rline.WithStdout(out),
		rline.WithStderr(out),
		rline.WithColor(false),
	)
	if err != nil {
		t.Fatalf("making the line editor: %v", err)
	}
	t.Cleanup(func() { _ = l.Close() })
	u := &user.User{Username: "usql", HomeDir: t.TempDir()}
	h, err := handler.New(l, feed.NewLines(strings.NewReader("")), u, t.TempDir(), memfs.New(), true)
	if err != nil {
		t.Fatalf("making the handler: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := h.Open(ctx, side.dsn); err != nil {
		t.Fatalf("opening the database: %s", side.hide(err.Error()))
	}
	return h
}

// load runs the script in the file path against the database, as usql -f
// does, and stops at the first error.
func (side dumpSide) load(t *testing.T, path string) {
	t.Helper()
	if err := env.Vars().SetSetting("on_error_stop", "on"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = env.Vars().SetSetting("on_error_stop", "off") }()
	out := new(bytes.Buffer)
	h := newDumpHandler(t, side, out)
	defer h.Close()
	if err := h.Include(path, false); err != nil {
		t.Fatalf("loading %s: %s\n%s", filepath.Base(path), side.hide(err.Error()), side.hide(tail(out.String(), 20)))
	}
}

// dump dumps the database with the options, and returns the script and the
// warnings that usql wrote.
func (side dumpSide) dump(t *testing.T, opts dump.Options) (string, string) {
	t.Helper()
	opts.RestrictKey = dumpKey
	out, script := new(bytes.Buffer), new(bytes.Buffer)
	h := newDumpHandler(t, side, out)
	defer h.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := h.Dump(ctx, script, opts); err != nil {
		t.Fatalf("dumping: %s\n%s", side.hide(err.Error()), side.hide(out.String()))
	}
	return script.String(), out.String()
}

// dumpFile dumps the database to a file, and returns its path.
func (side dumpSide) dumpFile(t *testing.T, opts dump.Options) string {
	t.Helper()
	script, warnings := side.dump(t, opts)
	if warnings != "" {
		t.Logf("the dump warned:\n%s", side.hide(warnings))
	}
	path := filepath.Join(t.TempDir(), "dump.sql")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// rowCounts returns the number of rows of each table, by the name of the
// table. The statement that reads the rows comes from dbmeta, so that the
// names are quoted as the dialect quotes them.
func (side dumpSide) rowCounts(t *testing.T, exclude ...string) map[string]int {
	t.Helper()
	ctx := context.Background()
	u, err := dburl.Parse(side.dsn)
	if err != nil {
		t.Fatal(side.hide(err.Error()))
	}
	db, err := drivers.Open(ctx, u, nil, nil)
	if err != nil {
		t.Fatal(side.hide(err.Error()))
	}
	defer db.Close()
	m, err := drivers.Meta(ctx, u, db)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := dbdump.Read(ctx, m, db, dbdump.Filter{ExcludeTables: exclude})
	if err != nil {
		t.Fatalf("reading the schema: %v", err)
	}
	counts := make(map[string]int)
	for _, sc := range snap.Schemas {
		for _, tbl := range sc.Tables {
			l, err := dbdump.NewLoader(snap, sc.Name, tbl.Name, m.Dialect(), dbdump.Policy{})
			if err != nil {
				t.Fatalf("reading %s.%s: %v", sc.Name, tbl.Name, err)
			}
			rows, err := db.QueryContext(ctx, l.SelectAll())
			if err != nil {
				t.Fatalf("reading %s.%s: %v", sc.Name, tbl.Name, err)
			}
			n := 0
			for rows.Next() {
				n++
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			_ = rows.Close()
			counts[tbl.Name] = n
		}
	}
	return counts
}

// fixture returns the path of the fixture of a dialect, and skips when there
// is none.
func fixture(t *testing.T, dialect string) string {
	t.Helper()
	path := filepath.Join("testdata", "dump", dialect+".sql")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("there is no fixture %s", path)
	}
	return path
}

// sameDump compares two dumps of databases that should hold the same thing.
// The name of each database is replaced, because a dump of MySQL names it.
func sameDump(t *testing.T, what string, want, got string, a, b dumpSide) {
	t.Helper()
	if a.name != "" && b.name != "" {
		got = strings.ReplaceAll(got, b.name, a.name)
	}
	if want != got {
		t.Errorf("%s differs:\n%s", what, a.hide(lineDiff(want, got)))
	}
}

// sameCounts compares the rows of each table of two databases.
func sameCounts(t *testing.T, want, got map[string]int) {
	t.Helper()
	if len(want) == 0 {
		t.Fatal("the source has no tables")
	}
	for name, n := range want {
		if m, ok := got[name]; !ok {
			t.Errorf("the target has no table %s", name)
		} else if m != n {
			t.Errorf("the target has %d rows in %s, want %d", m, name, n)
		}
	}
}

// tail returns the last n lines of s.
func tail(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// TestDumpEmbedded dumps each embedded database to itself and to the other,
// and needs no server.
func TestDumpEmbedded(t *testing.T) {
	for _, test := range []struct {
		from, to string
	}{
		{"sqlite3", "sqlite3"},
		{"duckdb", "duckdb"},
		{"sqlite3", "duckdb"},
		{"duckdb", "sqlite3"},
	} {
		t.Run(test.from+" to "+test.to, func(t *testing.T) {
			src, dst := embedded(t, test.from), embedded(t, test.to)
			src.load(t, fixture(t, test.from))
			opts := dump.Options{}
			if test.from != test.to {
				opts = dump.Options{Target: test.to, Lossy: true, Text: true}
			}
			dst.load(t, src.dumpFile(t, opts))
			sameCounts(t, src.rowCounts(t), dst.rowCounts(t))
			if test.from == test.to {
				want, _ := src.dump(t, dump.Options{})
				got, _ := dst.dump(t, dump.Options{})
				sameDump(t, "the dump of the copy", want, got, src, dst)
			}
		})
	}
}

// TestDump dumps the database of USQL_TEST_URL, and loads the script into
// another database of the same server, into each embedded database, and
// from an embedded database. When the dump tool of the product is on the
// path, it compares what that tool sees in a copy that usql made with what
// it sees in a copy that the tool made.
//
// The test makes databases named usql_dump_ and a suffix on the server, and
// drops them at the end.
func TestDump(t *testing.T) {
	s, db := openServer(t)
	dialect := dbmeta.Dialect(s.u.Dialect)
	if dialect.Embedded() {
		t.Skipf("%s is a file, and TestDumpEmbedded tests it", s.name)
	}
	if c := dbdump.Support(dialect); !c.Source || !c.Target {
		t.Skipf("%s cannot be dumped by usql yet: %v %v", s.name, c.SourceWhy, c.TargetWhy)
	}
	path := fixture(t, string(dialect))
	ctx := context.Background()
	if dialect == "mysql" && time.Local != time.UTC {
		// dbmeta sets the session of the source to UTC, but dburl gives the
		// MySQL driver loc=Local, so the driver reads the clock of the
		// session as local time, and the loader moves it to UTC. A TIMESTAMP
		// then moves on each copy in any zone but UTC. CI runs in UTC.
		// Reported to dbmeta on 2026-11-09, after v0.39.0.
		local := time.Local
		time.Local = time.UTC
		t.Cleanup(func() { time.Local = local })
	}
	side := func(t *testing.T, suffix string) dumpSide {
		t.Helper()
		name := "usql_dump_" + suffix
		for _, stmt := range []string{"DROP DATABASE IF EXISTS " + name, "CREATE DATABASE " + name} {
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				t.Fatalf("%s: %s", stmt, s.hide(err.Error()))
			}
		}
		t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DROP DATABASE IF EXISTS "+name) })
		u, err := url.Parse(s.raw)
		if err != nil {
			t.Fatal("the URL does not parse")
		}
		u.Path = "/" + name
		return dumpSide{dsn: u.String(), name: name, hide: s.hide}
	}
	src := side(t, "src")
	src.load(t, path)
	want, warnings := src.dump(t, dump.Options{})
	if warnings != "" {
		t.Logf("the dump of %s warned:\n%s", s.name, s.hide(warnings))
	}
	srcFile := filepath.Join(t.TempDir(), "src.sql")
	if err := os.WriteFile(srcFile, []byte(want), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Run("roundtrip", func(t *testing.T) {
		dst := side(t, "dst")
		dst.load(t, srcFile)
		got, _ := dst.dump(t, dump.Options{})
		sameDump(t, "the dump of the copy", want, got, src, dst)
	})

	t.Run("schema then data", func(t *testing.T) {
		dst := side(t, "split")
		dst.load(t, src.dumpFile(t, dump.Options{SchemaOnly: true}))
		dst.load(t, src.dumpFile(t, dump.Options{DataOnly: true}))
		got, _ := dst.dump(t, dump.Options{})
		sameDump(t, "the dump of the copy", want, got, src, dst)
	})

	for _, to := range []string{"sqlite3", "duckdb"} {
		t.Run("to "+to, func(t *testing.T) {
			if why, ok := knownCrossFailures[string(dialect)+" to "+to]; ok {
				t.Skip(why)
			}
			exclude := crossExclude[string(dialect)+" to "+to]
			dst := embedded(t, to)
			dst.load(t, src.dumpFile(t, dump.Options{Target: to, Lossy: true, Text: true, ExcludeTables: exclude}))
			sameCounts(t, src.rowCounts(t, exclude...), dst.rowCounts(t))
		})
		t.Run("from "+to, func(t *testing.T) {
			if why, ok := knownCrossFailures[to+" to "+string(dialect)]; ok {
				t.Skip(why)
			}
			from := embedded(t, to)
			from.load(t, fixture(t, to))
			dst := side(t, "from_"+to)
			dst.load(t, from.dumpFile(t, dump.Options{Target: string(dialect), Lossy: true, Text: true}))
			sameCounts(t, from.rowCounts(t), dst.rowCounts(t))
		})
	}

	t.Run("against the dump tool", func(t *testing.T) {
		tool := dumpTools[dialect]
		if tool == nil {
			t.Skipf("no dump tool is known for %s", dialect)
		}
		viaUsql, viaTool := side(t, "usql"), side(t, "tool")
		viaUsql.load(t, srcFile)
		toolDump := filepath.Join(t.TempDir(), "tool.sql")
		if err := tool.dump(s.u, src.name, toolDump); err != nil {
			t.Skipf("the dump tool did not run: %s", s.hide(err.Error()))
		}
		// a tool of another release or product, such as the mysqldump of
		// MariaDB with a server of MySQL, can write what its own client
		// does not load
		if err := tool.load(s.u, viaTool.name, toolDump); err != nil {
			t.Skipf("the client of the dump tool did not load its own dump: %s", s.hide(err.Error()))
		}
		a, b := filepath.Join(t.TempDir(), "a.sql"), filepath.Join(t.TempDir(), "b.sql")
		if err := tool.dump(s.u, viaTool.name, a); err != nil {
			t.Fatal(s.hide(err.Error()))
		}
		if err := tool.dump(s.u, viaUsql.name, b); err != nil {
			t.Fatal(s.hide(err.Error()))
		}
		known := knownDifferences(t, tool.name)
		wantText, gotText := tool.normalize(t, a, known), tool.normalize(t, b, known)
		gotText = strings.ReplaceAll(gotText, viaUsql.name, viaTool.name)
		if wantText != gotText {
			t.Errorf("%s sees a copy that usql made differently from a copy that %s made:\n%s", tool.name, tool.name, s.hide(lineDiff(wantText, gotText)))
		}
	})
}

// knownCrossFailures are the dumps from one dialect to another that fail
// because of a gap in dbmeta, by the source and the target, with the reason.
// A dump to another dialect is phase 3 of W77.
var knownCrossFailures = map[string]string{
	"postgres to duckdb": "dbmeta writes an array in the text form of PostgreSQL, which DuckDB does not read as a list. Reported to dbmeta on 2026-11-08",
}

// crossExclude are the tables that a dump from one dialect to another leaves
// out, by the source and the target, because the target cannot hold their
// values. SQLite has no NaN and no infinity, and dbmeta refuses such a row.
var crossExclude = map[string][]string{
	"postgres to sqlite3": {"app.floats"},
}

// dumpTool is the dump tool of a product, and its client.
type dumpTool struct {
	// name names the tool and the file of its known differences.
	name string
	// dump writes the database name to the file out, and load runs the file
	// in against the database name.
	dump func(u *dburl.URL, name, out string) error
	load func(u *dburl.URL, name, in string) error
}

// dumpTools are the dump tools that TestDump compares against.
var dumpTools = map[dbmeta.Dialect]*dumpTool{
	"postgres": {
		name: "pg_dump",
		dump: func(u *dburl.URL, name, out string) error {
			return runTool(nil, out, "", "pg_dump", "--no-owner", "--no-privileges", "-d", withDatabase(u, name))
		},
		load: func(u *dburl.URL, name, in string) error {
			return runTool(nil, "", in, "psql", "-X", "-q", "-v", "ON_ERROR_STOP=1", "-d", withDatabase(u, name))
		},
	},
	"mysql": {
		name: "mysqldump",
		dump: func(u *dburl.URL, name, out string) error {
			args := append(mysqlArgs(u), "--skip-dump-date", "--skip-comments", "--skip-extended-insert", name)
			return runTool(mysqlEnv(u), out, "", firstTool("mysqldump", "mariadb-dump"), args...)
		},
		load: func(u *dburl.URL, name, in string) error {
			return runTool(mysqlEnv(u), "", in, firstTool("mysql", "mariadb"), append(mysqlArgs(u), name)...)
		},
	},
}

// withDatabase returns the URL of the database name of the server of u, as
// the clients of PostgreSQL take it.
func withDatabase(u *dburl.URL, name string) string {
	v := u.URL
	v.Path = "/" + name
	return v.String()
}

// mysqlArgs are the options of the clients of MySQL for the server of u. The
// password goes in the environment, so that it is not on a command line.
func mysqlArgs(u *dburl.URL) []string {
	args := []string{"-h", u.Hostname(), "-u", u.User.Username()}
	if p := u.Port(); p != "" {
		args = append(args, "-P", p)
	}
	return args
}

// mysqlEnv is the environment of the clients of MySQL, with the password.
func mysqlEnv(u *dburl.URL) []string {
	pass, _ := u.User.Password()
	return append(os.Environ(), "MYSQL_PWD="+pass)
}

// firstTool returns the first of the names that is on the path.
func firstTool(names ...string) string {
	for _, name := range names {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return names[0]
}

// runTool runs a tool, with its standard output in the file out and its
// standard input from the file in, when they are named.
func runTool(environ []string, out, in, name string, args ...string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is not on the path", name)
	}
	cmd := exec.Command(name, args...)
	cmd.Env = environ
	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr
	if out != "" {
		f, err := os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd.Stdout = f
	}
	if in != "" {
		f, err := os.Open(in)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd.Stdin = f
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	go func() {
		<-ctx.Done()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// knownDifferences reads the patterns of the lines that the dump of a tool
// can differ in, in testdata/dump/NAME.known. Each pattern has a comment that
// says why, and names the work that removes it.
func knownDifferences(t *testing.T, name string) []*regexp.Regexp {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "dump", name+".known"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var res []*regexp.Regexp
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		re, err := regexp.Compile(line)
		if err != nil {
			t.Fatalf("%s.known: %v", name, err)
		}
		res = append(res, re)
	}
	return res
}

// normalize reads the dump of a tool, and leaves out the lines that differ
// from one run to the next, and the known differences.
func (tool *dumpTool) normalize(t *testing.T, path string, known []*regexp.Regexp) string {
	t.Helper()
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for line := range strings.SplitSeq(string(buf), "\n") {
		// pg_dump writes a random key on these
		if strings.HasPrefix(line, `\restrict `) || strings.HasPrefix(line, `\unrestrict `) {
			continue
		}
		if matchAny(known, line) {
			continue
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

// matchAny reports whether a pattern matches s.
func matchAny(res []*regexp.Regexp, s string) bool {
	for _, re := range res {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}
