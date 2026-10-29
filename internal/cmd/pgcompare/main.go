// Command pgcompare compares how psql and usql print every built-in type of
// PostgreSQL. It makes values of each type, some fixed and some random, and
// runs the same query through psql and through usql with the pgx and the
// libpq drivers. It prints a report of the output that differs, and it keeps
// the output of each client for each type.
//
// The command starts no database. Start PostgreSQL with dbmeta's dbrun, as
// "Command line tests" in CONTRIBUTING.md says, and pass its URL as -dsn. Build
// the usql to test first:
//
//	go build -tags all -o /tmp/usql .
//	go run ./internal/cmd/pgcompare -usql /tmp/usql -dsn "$(cd ../dbmeta/test && go run ./cmd/dbrun dsn --json postgres-18 | jq -r '.[0].url')"
//
// The same -seed makes the same values. The report is REPORT.md in the -out
// folder, and results.tsv in the same folder has one line for each type,
// format and driver.
package main

import (
	"bytes"
	"cmp"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand/v2"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	flagSeed  = flag.Uint64("seed", 1, "seed of the random values")
	flagN     = flag.Int("n", 25, "random values of each type")
	flagUsql  = flag.String("usql", "usql", "the usql binary")
	flagDSN   = flag.String("dsn", "", "URL of the PostgreSQL database, as postgres://user:pass@host:port/db")
	flagOut   = flag.String("out", "", "folder of the results (default a new temporary folder)")
	flagTypes = flag.String("types", "", "only these types, separated by commas")
)

// dsn is the URL of -dsn.
var dsn *url.URL

// format is a way to print a result, with the flags of psql and of usql.
type format struct {
	name       string
	psql, usql []string
}

var formats = []format{
	{"aligned", []string{"-P", "pager=off"}, []string{"-P", "pager=off"}},
	{"unaligned", []string{"-A", "-P", "pager=off"}, []string{"-A", "-P", "pager=off"}},
	{"csv", []string{"--csv"}, []string{"-C"}},
}

var clients = []string{"psql", "libpq", "pgx"}

func main() {
	flag.Parse()
	var err error
	if dsn, err = url.Parse(*flagDSN); err != nil || *flagDSN == "" {
		fatal(fmt.Errorf("-dsn must be the URL of a PostgreSQL database: %q", *flagDSN))
	}
	if *flagOut == "" {
		if *flagOut, err = os.MkdirTemp("", "pgcompare-"); err != nil {
			fatal(err)
		}
	}
	if err := os.MkdirAll(*flagOut, 0o755); err != nil {
		fatal(err)
	}
	only := map[string]bool{}
	for _, t := range strings.Split(*flagTypes, ",") {
		if t != "" {
			only[t] = true
		}
	}
	r := rand.New(rand.NewPCG(*flagSeed, *flagSeed^0x9e3779b97f4a7c15))
	var report strings.Builder
	fmt.Fprintf(&report, "seed %d, %d random values of each type\n\n", *flagSeed, *flagN)
	same, differ := 0, 0
	tsv, err := os.Create(filepath.Join(*flagOut, "results.tsv"))
	if err != nil {
		fatal(err)
	}
	defer tsv.Close()
	for _, c := range allCases() {
		if len(only) != 0 && !only[c.name] {
			continue
		}
		exprs := append([]string{}, c.edges...)
		for range *flagN {
			exprs = append(exprs, c.gen(r))
		}
		exprs = append(exprs, "NULL::"+c.name)
		exprs = validate(c.name, exprs)
		if len(exprs) == 0 {
			fmt.Fprintf(&report, "## %s\n\nno value is valid\n\n", c.name)
			continue
		}
		q := query(exprs)
		file := filepath.Join(*flagOut, safe(c.name)+".sql")
		if err := os.WriteFile(file, []byte(q), 0o644); err != nil {
			fatal(err)
		}
		var sb strings.Builder
		for _, f := range formats {
			want, err := run("psql", f, file)
			if err != nil {
				fmt.Fprintf(&sb, "- %s: psql fails: %s\n", f.name, firstLine(err.Error()))
				continue
			}
			for _, cl := range clients[1:] {
				got, err := run(cl, f, file)
				if err != nil {
					got = append(got, []byte("\nERROR: "+err.Error())...)
				}
				_ = os.WriteFile(filepath.Join(*flagOut, safe(c.name)+"."+f.name+"."+cl+".txt"), got, 0o644)
				_ = os.WriteFile(filepath.Join(*flagOut, safe(c.name)+"."+f.name+".psql.txt"), want, 0o644)
				d := compare(f.name, want, got)
				status := "same"
				switch {
				case strings.HasPrefix(d, "  - only the spaces"):
					status = "trailing"
				case d != "":
					status = "differs"
				}
				fmt.Fprintf(tsv, "%s\t%s\t%s\t%s\n", c.name, f.name, cl, status)
				if d != "" {
					differ++
					fmt.Fprintf(&sb, "- %s, %s:\n%s", f.name, cl, d)
				} else {
					same++
				}
			}
		}
		status := "same"
		if sb.Len() != 0 {
			status = "differs"
		}
		fmt.Fprintf(&report, "## %s: %s (%d values)\n\n%s\n", c.name, status, len(exprs), sb.String())
		fmt.Fprintf(os.Stderr, "%-16s %s\n", c.name, status)
	}
	fmt.Fprintf(&report, "\n%d comparisons are the same, %d differ\n", same, differ)
	if err := os.WriteFile(filepath.Join(*flagOut, "REPORT.md"), []byte(report.String()), 0o644); err != nil {
		fatal(err)
	}
	fmt.Fprintf(os.Stderr, "%d same, %d differ, report in %s\n", same, differ, filepath.Join(*flagOut, "REPORT.md"))
}

// allCases adds the array of each type that has one.
func allCases() []typeCase {
	var out []typeCase
	for _, c := range cases {
		out = append(out, c)
		if !c.array {
			continue
		}
		out = append(out, typeCase{
			name: c.name + "[]",
			gen: func(r *rand.Rand) string {
				n := r.IntN(4)
				parts := make([]string, n)
				for i := range parts {
					if r.IntN(6) == 0 {
						parts[i] = "NULL"
					} else {
						parts[i] = c.gen(r)
					}
				}
				if n == 0 {
					return "'{}'::" + c.name + "[]"
				}
				if r.IntN(5) == 0 && n > 0 {
					// two rows of an array of two dimensions
					return "ARRAY[ARRAY[" + strings.Join(parts, ",") + "]::" + c.name + "[], ARRAY[" + strings.Join(parts, ",") + "]::" + c.name + "[]]"
				}
				return "ARRAY[" + strings.Join(parts, ",") + "]::" + c.name + "[]"
			},
		})
	}
	return append(out, recordCase)
}

var recordCase = typeCase{name: "record", gen: func(r *rand.Rand) string {
	return "row(" + cases[2].gen(r) + ", " + lit(randText(r, 8), "text") + ", NULL::int, " + cases[16].gen(r) + ")"
}}

// query selects each expression as a row, with its number.
func query(exprs []string) string {
	var b strings.Builder
	b.WriteString("select n, v from (values\n")
	for i, e := range exprs {
		if i > 0 {
			b.WriteString(",\n")
		}
		fmt.Fprintf(&b, "  (%d, %s)", i+1, e)
	}
	b.WriteString("\n) as t(n, v) order by n;\n")
	return b.String()
}

// validate drops the expressions that PostgreSQL refuses, with each one in a
// query of its own, when the query of all of them fails.
func validate(typ string, exprs []string) []string {
	file := filepath.Join(*flagOut, safe(typ)+".check.sql")
	if err := os.WriteFile(file, []byte(query(exprs)), 0o644); err != nil {
		fatal(err)
	}
	if _, err := run("psql", formats[2], file); err == nil {
		return exprs
	}
	var ok []string
	for _, e := range exprs {
		if err := os.WriteFile(file, []byte(query([]string{e})), 0o644); err != nil {
			fatal(err)
		}
		if _, err := run("psql", formats[2], file); err == nil {
			ok = append(ok, e)
		} else {
			fmt.Fprintf(os.Stderr, "%s: dropping %s: %s\n", typ, e, firstLine(err.Error()))
		}
	}
	return ok
}

// run runs file with a client, in a format, and returns what it writes.
func run(client string, f format, file string) ([]byte, error) {
	var cmd *exec.Cmd
	switch client {
	case "psql":
		pass, _ := dsn.User.Password()
		args := []string{"-X", "-q", "-v", "ON_ERROR_STOP=1", "-h", dsn.Hostname(), "-p", cmp.Or(dsn.Port(), "5432"), "-U", dsn.User.Username(), "-d", strings.TrimPrefix(dsn.Path, "/")}
		cmd = exec.Command("psql", append(append(args, f.psql...), "-f", file)...)
		cmd.Env = append(os.Environ(), "PGPASSWORD="+pass)
	default:
		u := *dsn
		u.Scheme = map[string]string{"libpq": "pq", "pgx": "pgx"}[client]
		args := append([]string{"-X", "-q"}, f.usql...)
		cmd = exec.Command(*flagUsql, append(args, "-f", file, u.String())...)
		cmd.Env = os.Environ()
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	cmd.Env = append(cmd.Env, "PAGER=cat", "USQL_PAGER=cat")
	err := cmd.Run()
	switch {
	case err != nil:
		return stdout.Bytes(), fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	case stderr.Len() != 0:
		return stdout.Bytes(), fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// compare returns the lines where got differs from want, or the empty string.
func compare(name string, want, got []byte) string {
	if bytes.Equal(want, got) {
		return ""
	}
	var b strings.Builder
	if name == "csv" {
		w, err1 := csv.NewReader(bytes.NewReader(want)).ReadAll()
		g, err2 := csv.NewReader(bytes.NewReader(got)).ReadAll()
		if err1 == nil && err2 == nil {
			if len(w) != len(g) {
				fmt.Fprintf(&b, "  - rows: psql %d, usql %d\n", len(w), len(g))
			}
			n := 0
			for i := 0; i < len(w) && i < len(g) && n < 4; i++ {
				for j := 0; j < len(w[i]) && j < len(g[i]); j++ {
					if w[i][j] != g[i][j] {
						fmt.Fprintf(&b, "  - row %d: psql %q, usql %q\n", i, w[i][j], g[i][j])
						n++
					}
				}
			}
			if b.Len() == 0 {
				b.WriteString("  - the bytes differ, and the cells are the same\n")
			}
			return b.String()
		}
	}
	wl, gl := strings.Split(string(want), "\n"), strings.Split(string(got), "\n")
	trim := func(l []string) string {
		out := make([]string, len(l))
		for i, s := range l {
			out[i] = strings.TrimRight(s, " ")
		}
		return strings.Join(out, "\n")
	}
	if trim(wl) == trim(gl) {
		return "  - only the spaces at the end of the lines differ\n"
	}
	for i := range wl {
		wl[i] = strings.TrimRight(wl[i], " ")
	}
	for i := range gl {
		gl[i] = strings.TrimRight(gl[i], " ")
	}
	n := 0
	for i := 0; i < len(wl) || i < len(gl); i++ {
		var a, c string
		if i < len(wl) {
			a = wl[i]
		}
		if i < len(gl) {
			c = gl[i]
		}
		if a != c {
			fmt.Fprintf(&b, "  - line %d:\n    psql %q\n    usql %q\n", i+1, a, c)
			if n++; n == 3 {
				break
			}
		}
	}
	return b.String()
}

func safe(s string) string {
	return strings.NewReplacer("[]", "_array", "(", "_", ")", "", ",", "_", `"`, "q", " ", "_").Replace(s)
}

func firstLine(s string) string {
	s, _, _ = strings.Cut(s, "\n")
	return s
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
