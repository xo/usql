// Package readme fills the examples of the README with what usql prints.
//
// An example is a code block of the README with a marker on the line above it:
//
//	<!-- EXAMPLE name -->
//
// gen.go runs it with the -examples flag. Generate builds usql, runs each
// example in [Examples], and writes what usql printed into the block that
// follows the marker of the same name. An example that types input runs usql
// under a pseudo-terminal, so the block shows the prompts and the input as a
// person sees them. A pseudo-terminal is a pair of file descriptors that acts
// as a terminal for another program.
//
// An example that needs a database runs against a server of dbmeta's dbrun,
// with dbmeta's fixture examples: the tables authors and books. gen.go does
// not start or stop the server. The README shows SQL Server, and any other
// base database can run the examples for a comparison.
package readme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/xo/usql/internal/capture"
)

// Example is a code block of the README that usql fills.
type Example struct {
	// Name is the name in the marker of the block.
	Name string
	// Tags are the build tags of the usql that runs the example. Empty is the
	// most group, which a release ships.
	Tags string
	// Lines are written at the top of the block as they are, such as a
	// comment or a go build command.
	Lines []string
	// Prog is the command that the block shows for usql, "usql" when empty.
	Prog string
	// Args are the arguments of usql.
	Args []string
	// Command shows the command line and what usql prints when it starts. An
	// example with no Input always shows them.
	Command bool
	// Setup is typed before Input, and the block does not show it or what it
	// printed.
	Setup []string
	// Input is typed a line at a time. A line that ends with a Tab is sent
	// without Enter, so that the block shows the completions. With no input,
	// and Command not set, usql runs without a terminal, and the block shows
	// what it printed.
	Input []string
	// Products replaces Input for the products of dbrun that it names, such
	// as oracle, for a statement that is written differently there.
	Products map[string][]string
	// DB runs the example against the database. "{url}" in Args, Setup and
	// Input is its URL.
	DB bool
	// Env holds extra environment entries, each "NAME=value".
	Env []string
	// Files are written into the directory that usql runs in before it
	// starts, each a name and its text. The block shows each with cat.
	Files [][2]string
}

// Options say where Generate runs the examples and where it writes them.
type Options struct {
	// Dir is the directory of usql, which holds the README.
	Dir string
	// DB is the selector of dbrun for the database, such as sqlserver-2025.
	DB string
	// Dbmeta is the directory of a checkout of dbmeta, whose test directory
	// runs dbrun.
	Dbmeta string
	// Out, when set, is a directory that gets each example as NAME.txt, and
	// the README is left as it is.
	Out string
}

// db is the database of the examples.
type db struct {
	product string
	url     string
	// port and pass are the port and the password of the server, which the
	// README must not show.
	port, pass string
	// shown replaces the text of the URL that the README must not show, such
	// as the password and the port that dbrun chose.
	shown *strings.Replacer
	// display is the URL as the README shows it.
	display string
}

// env is the environment of usql, so that nothing on the machine that runs
// gen.go changes what the README shows. An empty PAGER turns the pager off.
var env = []string{"USER=user", "LANG=C.UTF-8", "PAGER="}

// Generate runs every example, and writes what usql printed into the README.
func Generate(ctx context.Context, opts Options) error {
	dir := opts.Dir
	name := filepath.Join(dir, "README.md")
	readme, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	if err := check(readme); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "usql-readme-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	var d *db
	if slices.ContainsFunc(Examples, func(ex Example) bool { return ex.DB }) {
		if d, err = openDB(ctx, opts.Dbmeta, opts.DB); err != nil {
			return err
		}
	}
	bins := make(map[string]string)
	for _, ex := range Examples {
		if ex.Tags == "" {
			ex.Tags = "most"
		}
		bin, ok := bins[ex.Tags]
		if !ok {
			bin = filepath.Join(tmp, fmt.Sprintf("usql-%d", len(bins)))
			if err := build(ctx, dir, bin, ex.Tags); err != nil {
				return err
			}
			bins[ex.Tags] = bin
		}
		text, err := ex.run(ctx, bin, d)
		if err != nil {
			return fmt.Errorf("running the example %s: %w", ex.Name, err)
		}
		if opts.Out != "" {
			if err := os.WriteFile(filepath.Join(opts.Out, ex.Name+".txt"), []byte(text+"\n"), 0o644); err != nil {
				return err
			}
			continue
		}
		if readme, err = Replace(readme, ex.Name, text); err != nil {
			return err
		}
	}
	if opts.Out != "" {
		return nil
	}
	return os.WriteFile(name, readme, 0o644)
}

// openDB builds the fixture examples on the database that selector names, and
// returns its URL. The server must be running already.
func openDB(ctx context.Context, dbmeta, selector string) (*db, error) {
	dbrun := func(args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, "go", append([]string{"run", "./cmd/dbrun"}, args...)...)
		cmd.Dir = filepath.Join(dbmeta, "test")
		cmd.Env = append(os.Environ(), "DBMETA_OWNER_NAME=usql")
		out, err := cmd.Output()
		if err != nil {
			var stderr []byte
			if e, ok := err.(*exec.ExitError); ok {
				stderr = e.Stderr
			}
			return nil, fmt.Errorf("dbrun %s: %w\n%s\nStart the server with: dbrun start %s", strings.Join(args, " "), err, stderr, selector)
		}
		return out, nil
	}
	if _, err := dbrun("fixture", "examples", selector); err != nil {
		return nil, err
	}
	out, err := dbrun("dsn", "--json", selector)
	if err != nil {
		return nil, err
	}
	var servers []struct {
		Product    string `json:"product"`
		Principals []struct {
			URL string `json:"url"`
		} `json:"principals"`
	}
	if err := json.Unmarshal(out, &servers); err != nil {
		return nil, fmt.Errorf("reading the URL from dbrun: %w", err)
	}
	if len(servers) == 0 || len(servers[0].Principals) == 0 {
		return nil, fmt.Errorf("dbrun gave no URL for %s", selector)
	}
	s := servers[0]
	u, err := url.Parse(s.Principals[0].URL)
	if err != nil {
		return nil, err
	}
	// on MySQL and MariaDB the fixture is in the database examples
	if s.Product == "mysql" || s.Product == "mariadb" {
		u.Path = "/examples"
	}
	d := &db{product: s.Product, url: u.String(), port: u.Port()}
	if pass, ok := u.User.Password(); ok {
		d.pass = pass
	}
	// the prompt shows the host up to its first dot, so 127.0.0.1 shows as
	// 127. localhost is not used to connect, because a driver can try only
	// its IPv6 address, where the server does not listen.
	pairs := []string{u.String(), "", "@127=>", "@localhost=>", "@127->", "@localhost->", "127.0.0.1", "localhost"}
	shown := &url.URL{Scheme: u.Scheme, Host: "localhost", Path: u.Path, Opaque: u.Opaque}
	if u.User != nil {
		shown.User = url.UserPassword(u.User.Username(), "pass")
		if pass, ok := u.User.Password(); ok && pass != "" {
			pairs = append(pairs, url.QueryEscape(pass), "pass", pass, "pass")
		}
	}
	if port := u.Port(); port != "" {
		pairs = append(pairs, "localhost:"+port, "localhost", "127.0.0.1:"+port, "localhost")
	}
	d.display = shown.String()
	pairs[1] = d.display
	d.shown = strings.NewReplacer(pairs...)
	return d, nil
}

// markerRE matches the marker of an example.
var markerRE = regexp.MustCompile(`(?m)^<!-- EXAMPLE ([a-z0-9-]+) -->\n`)

// check reports a marker with no example, and an example with no marker.
func check(readme []byte) error {
	var marked []string
	for _, m := range markerRE.FindAllSubmatch(readme, -1) {
		marked = append(marked, string(m[1]))
	}
	for _, name := range marked {
		if !slices.ContainsFunc(Examples, func(ex Example) bool { return ex.Name == name }) {
			return fmt.Errorf("the README marks the example %s, which does not exist", name)
		}
	}
	for _, ex := range Examples {
		if n := slices.Index(marked, ex.Name); n == -1 {
			return fmt.Errorf("the README has no marker for the example %s", ex.Name)
		}
	}
	return nil
}

// Replace sets the text of the code block after the marker of name.
func Replace(readme []byte, name, text string) ([]byte, error) {
	marker := []byte("<!-- EXAMPLE " + name + " -->\n")
	i := bytes.Index(readme, marker)
	if i == -1 {
		return nil, fmt.Errorf("the README has no marker for the example %s", name)
	}
	start := i + len(marker)
	open, _, ok := bytes.Cut(readme[start:], []byte("\n"))
	if !ok || !bytes.HasPrefix(open, []byte("```")) {
		return nil, fmt.Errorf("the marker of the example %s is not followed by a code block", name)
	}
	body := start + len(open) + 1
	end := bytes.Index(readme[body:], []byte("\n```\n"))
	if end == -1 {
		return nil, fmt.Errorf("the code block of the example %s does not end", name)
	}
	var out []byte
	out = append(out, readme[:body]...)
	out = append(out, text...)
	out = append(out, readme[body+end:]...)
	return out, nil
}

// build builds usql with tags into bin.
func build(ctx context.Context, dir, bin, tags string) error {
	cmd := exec.CommandContext(ctx, "go", "build", "-o", bin, "-tags", tags, ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("building usql with the tags %q: %w\n%s", tags, err, out)
	}
	return nil
}

// run runs the example with the usql at bin, and returns the text of its
// block.
func (ex Example) run(ctx context.Context, bin string, d *db) (string, error) {
	shownArgs := ex.Args
	if ex.DB {
		if d == nil {
			return "", fmt.Errorf("the example %s needs a database", ex.Name)
		}
		with := func(v []string) []string {
			out := make([]string, len(v))
			for i, s := range v {
				out[i] = strings.ReplaceAll(s, "{url}", d.url)
			}
			return out
		}
		if input, ok := ex.Products[d.product]; ok {
			ex.Input = input
		}
		// the command line is quoted for the URL that the README shows
		shownArgs = make([]string, len(ex.Args))
		for i, a := range ex.Args {
			shownArgs[i] = strings.ReplaceAll(a, "{url}", d.display)
		}
		ex.Args, ex.Setup, ex.Input = with(ex.Args), with(ex.Setup), with(ex.Input)
	}
	lines := slices.Clone(ex.Lines)
	for _, f := range ex.Files {
		lines = append(lines, "$ cat "+f[0], strings.TrimRight(f[1], "\n"))
	}
	prog := ex.Prog
	if prog == "" {
		prog = "usql"
	}
	command := strings.TrimSpace("$ " + prog + " " + shellJoin(shownArgs))
	var out string
	var err error
	switch {
	case len(ex.Input) == 0 && !ex.Command:
		out, err = ex.plain(ctx, bin)
		lines = append(lines, command)
	default:
		out, err = ex.session(ctx, bin)
		if ex.Command {
			lines = append(lines, command)
		}
	}
	if err != nil {
		return "", err
	}
	text := strings.Join(append(lines, out), "\n")
	if d != nil {
		text = d.shown.Replace(text)
		if (d.port != "" && strings.Contains(text, d.port)) || (d.pass != "" && strings.Contains(text, d.pass)) {
			return "", fmt.Errorf("the text shows the port or the password of the server:\n%s", strings.ReplaceAll(strings.ReplaceAll(text, d.pass, "<password>"), d.port, "<port>"))
		}
	}
	return text, nil
}

// home makes the directory that usql runs in, with the files of the example.
func (ex Example) home() (string, error) {
	home, err := os.MkdirTemp("", "usql-readme-home-")
	if err != nil {
		return "", err
	}
	for _, f := range ex.Files {
		if err := os.WriteFile(filepath.Join(home, f[0]), []byte(f[1]), 0o644); err != nil {
			return "", err
		}
	}
	return home, nil
}

// plain runs usql with no terminal, and returns what it printed.
func (ex Example) plain(ctx context.Context, bin string) (string, error) {
	home, err := ex.home()
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(home)
	cmd := exec.CommandContext(ctx, bin, ex.Args...)
	cmd.Dir = home
	cmd.Env = append(append([]string{"HOME=" + home, "PATH=" + os.Getenv("PATH")}, env...), ex.Env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w\n%s", err, out)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// session runs usql under a pseudo-terminal, types the input, and returns the
// text that the terminal shows.
func (ex Example) session(ctx context.Context, bin string) (string, error) {
	home, err := ex.home()
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(home)
	var steps []capture.Step
	for _, s := range append(slices.Clone(ex.Setup), ex.Input...) {
		if !strings.HasSuffix(s, capture.KeyTab) {
			s += capture.KeyEnter
		}
		steps = append(steps, capture.Step{Send: s})
	}
	// a session that types nothing still waits for the prompt
	if len(steps) == 0 {
		steps = append(steps, capture.Step{})
	}
	t, err := capture.Record(ctx, bin, capture.Session{
		Name:  ex.Name,
		Args:  ex.Args,
		Env:   append(slices.Clone(env), ex.Env...),
		Dir:   home,
		Cols:  200,
		Rows:  500,
		Steps: steps,
		Quiet: 500 * time.Millisecond,
	})
	if err != nil {
		return "", err
	}
	// the first exchange holds what usql printed when it started
	from := 1 + len(ex.Setup)
	if ex.Command && len(ex.Setup) == 0 {
		from = 0
	}
	var out []byte
	for _, e := range t.Exchanges[from:] {
		out = append(out, e.Recv...)
	}
	return Render(out), nil
}

// shellJoin writes args as a shell reads them, quoting the ones that need it.
func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		if a == "" || strings.ContainsAny(a, " \t\\'\"$`*?;&|<>()[]{}!#~") {
			a = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
		}
		quoted[i] = a
	}
	return strings.Join(quoted, " ")
}
