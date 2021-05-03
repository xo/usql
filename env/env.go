// Package env contains runtime environment variables for usql, along with
// various helper funcs to determine the user's configuration.
package env

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/xo/dburl"
	"github.com/xo/usql/text"
	"github.com/zaf/temp"
)

// Getenv tries retrieving successive keys from os environment variables.
func Getenv(keys ...string) string {
	for _, key := range keys {
		if s := os.Getenv(key); s != "" {
			return s
		}
	}
	return ""
}

// Expand expands the tilde (~) in the front of a path to a the supplied
// directory.
func Expand(u *user.User, path string) string {
	switch {
	case path == "~":
		return u.HomeDir
	case strings.HasPrefix(path, "~/"):
		return filepath.Join(u.HomeDir, strings.TrimPrefix(path, "~/"))
	}
	return path
}

// Chdir changes the current working directory to the specified path, or to the
// user's home directory if path is not specified.
func Chdir(u *user.User, path string) error {
	if path != "" {
		path = Expand(u, path)
	} else {
		path = u.HomeDir
	}
	return os.Chdir(path)
}

// OpenFile opens a file for reading, returning the full, expanded path of the
// file.  All callers are responsible for closing the returned file.
func OpenFile(u *user.User, path string, relative bool) (string, *os.File, error) {
	path, err := filepath.EvalSymlinks(Expand(u, path))
	switch {
	case err != nil && os.IsNotExist(err):
		return "", nil, text.ErrNoSuchFileOrDirectory
	case err != nil:
		return "", nil, err
	}
	fi, err := os.Stat(path)
	switch {
	case err != nil && os.IsNotExist(err):
		return "", nil, text.ErrNoSuchFileOrDirectory
	case err != nil:
		return "", nil, err
	case fi.IsDir():
		return "", nil, text.ErrCannotIncludeDirectories
	}
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return "", nil, err
	}
	return path, f, nil
}

// EditFile edits a file. If path is empty, then a temporary file will be created.
func EditFile(u *user.User, path, line, s string) ([]rune, error) {
	ed := Getenv(text.CommandUpper()+"_EDITOR", "EDITOR", "VISUAL")
	if ed == "" {
		return nil, text.ErrNoEditorDefined
	}
	if path != "" {
		path = Expand(u, path)
	} else {
		f, err := temp.File("", text.CommandLower(), "sql")
		if err != nil {
			return nil, err
		}
		err = f.Close()
		if err != nil {
			return nil, err
		}
		path = f.Name()
		err = ioutil.WriteFile(path, []byte(strings.TrimSuffix(s, "\n")+"\n"), 0o644)
		if err != nil {
			return nil, err
		}
	}
	// setup args
	args := []string{path}
	if line != "" {
		prefix := Getenv(text.CommandUpper() + "_EDITOR_LINENUMBER_ARG")
		if prefix == "" {
			prefix = "+"
		}
		args = append(args, prefix+line)
	}
	// create command
	c := exec.Command(ed, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	// run
	if err := c.Run(); err != nil {
		return nil, err
	}
	// read
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return []rune(strings.TrimSuffix(string(buf), "\n")), nil
}

// HistoryFile returns the path to the history file.
//
// Defaults to ~/.<command name>_history, overridden by environment variable
// <COMMAND NAME>_HISTORY (ie, ~/.usql_history and USQL_HISTORY).
func HistoryFile(u *user.User) string {
	n := text.CommandUpper() + "_HISTORY"
	path := "~/." + strings.ToLower(n)
	if s := Getenv(n); s != "" {
		path = s
	}
	return Expand(u, path)
}

// RCFile returns the path to the RC file.
//
// Defaults to ~/.<command name>rc, overridden by environment variable
// <COMMAND NAME>RC (ie, ~/.usqlrc and USQLRC).
func RCFile(u *user.User) string {
	n := text.CommandUpper() + "RC"
	path := "~/." + strings.ToLower(n)
	if s := Getenv(n); s != "" {
		path = s
	}
	return Expand(u, path)
}

// PassFile returns the path to the password file.
//
// Defaults to ~/.<command name>pass, overridden by environment variable
// <COMMAND NAME>PASS (ie, ~/.usqlpass and USQLPASS).
func PassFile(u *user.User) string {
	n := text.CommandUpper() + "PASS"
	path := "~/." + strings.ToLower(n)
	if s := Getenv(n); s != "" {
		path = s
	}
	return Expand(u, path)
}

// PassFileMatch returns a Userinfo from a passfile entry matching
// the database URL.
func PassFileMatch(u *user.User, v *dburl.URL) (*url.Userinfo, error) {
	// check if v already has password defined ...
	var username string
	if v.User != nil {
		username = v.User.Username()
		if _, ok := v.User.Password(); ok {
			return nil, nil
		}
	}
	entries, err := readPassEntries(PassFile(u))
	if err != nil || entries == nil {
		return nil, err
	}
	// find matching entry
	n := strings.SplitN(v.Normalize(":", "", 3), ":", 6)
	if len(n) < 3 {
		return nil, errors.New("unknown error encountered normalizing URL")
	}
	m := newPassFileEntry(n)
	for _, entry := range entries {
		if entry.equal(m) {
			u := entry.Username
			if entry.Username == "*" {
				u = username
			}
			return url.UserPassword(u, entry.Password), nil
		}
	}
	return nil, nil
}

func PassFileEntries(u *user.User) ([]PassFileEntry, error) {
	return readPassEntries(PassFile(u))
}

// commentRE matches comment entries in a pass file.
var commentRE = regexp.MustCompile(`#.*`)

// readPassEntries reads the pass file entries from path.
func readPassEntries(path string) ([]PassFileEntry, error) {
	// check if pass file exists
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	// check pass file is not directory
	if fi.IsDir() {
		return nil, fmt.Errorf(text.BadPassFile, path)
	}
	// check pass file is not group/world readable/writable/executable
	if runtime.GOOS != "windows" && fi.Mode()&0x3f != 0 {
		return nil, fmt.Errorf(text.BadPassFileMode, path)
	}

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var entries []PassFileEntry
	s := bufio.NewScanner(f)
	i := 0
	for s.Scan() {
		i++
		// grab next line
		line := strings.TrimSpace(commentRE.ReplaceAllString(s.Text(), ""))
		if line == "" {
			continue
		}
		// split and check length
		v := strings.Split(line, ":")
		if len(v) != 6 {
			return nil, fmt.Errorf(text.BadPassFileLine, i)
		}
		// make sure no blank entries exist
		for j := 0; j < len(v); j++ {
			if v[j] == "" {
				return nil, fmt.Errorf(text.BadPassFileFieldEmpty, i, j)
			}
		}
		entries = append(entries, newPassFileEntry(v))
	}
	return entries, nil
}

type PassFileEntry struct {
	Protocol, Host, Port, DBName, Username, Password string
}

func newPassFileEntry(v []string) PassFileEntry {
	// make sure there's always at least 6 elements
	v = append(v, "", "", "", "", "", "")
	return PassFileEntry{
		Protocol: v[0],
		Host:     v[1],
		Port:     v[2],
		DBName:   v[3],
		Username: v[4],
		Password: v[5],
	}
}

func (e PassFileEntry) equal(b PassFileEntry) bool {
	return (e.Protocol == "*" || e.Protocol == b.Protocol) &&
		(e.Host == "*" || e.Host == b.Host) &&
		(e.Port == "*" || e.Port == b.Port)
}

// Getshell returns the user's defined SHELL, or system default (if found on
// path) and the appropriate command-line argument for the returned shell.
//
// Looks at the SHELL environment variable first, and then COMSPEC/ComSpec on
// Windows. Defaults to sh on non-Windows systems, and to cmd.exe on Windows.
func Getshell() (string, string) {
	var shell, param string
	shell, param = Getenv("SHELL"), "-c"
	if shell == "" && runtime.GOOS == "windows" {
		shell, param = Getenv("COMSPEC", "ComSpec"), "/c"
	}
	// look up path for "cmd.exe" if no other SHELL
	if shell == "" && runtime.GOOS == "windows" {
		shell, _ = exec.LookPath("cmd.exe")
		if shell != "" {
			param = "/c"
		}
	}
	// lookup path for "sh" if no other SHELL
	if shell == "" {
		shell, _ = exec.LookPath("sh")
		if shell != "" {
			param = "-c"
		}
	}
	return shell, param
}

// Shell runs s as a shell. When s is empty the user's SHELL or COMSPEC is
// used. See Getshell.
func Shell(s string) error {
	shell, param := Getshell()
	if shell == "" {
		return text.ErrNoShellAvailable
	}
	s = strings.TrimSpace(s)
	var params []string
	if s != "" {
		params = append(params, param, s)
	}
	// drop to shell
	cmd := exec.Command(shell, params...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	_ = cmd.Run()
	return nil
}

// Pipe starts a command and returns its input for writing.
func Pipe(c string) (io.WriteCloser, *exec.Cmd, error) {
	shell, param := Getshell()
	if shell == "" {
		return nil, nil, text.ErrNoShellAvailable
	}
	cmd := exec.Command(shell, param, c)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	out, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	return out, cmd, cmd.Start()
}

// Exec executes s using the user's SHELL / COMSPEC with -c (or /c) and
// returning the captured output. See Getshell.
//
// When SHELL or COMSPEC is not defined, then "sh" / "cmd.exe" will be used
// instead, assuming it is found on the system's PATH.
func Exec(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	shell, param := Getshell()
	if shell == "" {
		return "", text.ErrNoShellAvailable
	}
	buf, err := exec.Command(shell, param, s).CombinedOutput()
	if err != nil {
		return "", err
	}
	// remove ending \r\n
	buf = bytes.TrimSuffix(buf, []byte{'\n'})
	buf = bytes.TrimSuffix(buf, []byte{'\r'})
	return string(buf), nil
}

var cleanDoubleRE = regexp.MustCompile(`(^|[^\\])''`)

// Dequote unquotes a string.
func Dequote(s string, quote byte) (string, error) {
	if len(s) < 2 || s[len(s)-1] != quote {
		return "", text.ErrUnterminatedQuotedString
	}
	s = s[1 : len(s)-1]
	if quote == '\'' {
		s = cleanDoubleRE.ReplaceAllString(s, "$1\\'")
	}

	// this is the last part of strconv.Unquote
	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	for len(s) > 0 {
		c, multibyte, ss, err := strconv.UnquoteChar(s, quote)
		switch {
		case err != nil && err == strconv.ErrSyntax:
			return "", text.ErrInvalidQuotedString
		case err != nil:
			return "", err
		}
		s = ss
		if c < utf8.RuneSelf || !multibyte {
			buf = append(buf, byte(c))
		} else {
			n := utf8.EncodeRune(runeTmp[:], c)
			buf = append(buf, runeTmp[:n]...)
		}
	}
	return string(buf), nil
}

// Getvar retrieves an environment variable.
func Getvar(s string, v Vars) (bool, string, error) {
	q, n := "", s
	if c := s[0]; c == '\'' || c == '"' {
		var err error
		if n, err = Dequote(s, c); err != nil {
			return false, "", err
		}
		q = string(c)
	}
	if val, ok := v[n]; ok {
		return true, q + val + q, nil
	}
	return false, s, nil
}

// Unquote returns a func that unquotes strings for the user.
//
// When exec is true, backtick'd strings (``) will be executed using the
// provided user's shell (see Exec).
func Unquote(u *user.User, exec bool, v Vars) func(string, bool) (bool, string, error) {
	return func(s string, isvar bool) (bool, string, error) {
		// log.Printf(">>> UNQUOTE: %q", s)
		if isvar {
			return Getvar(s, v)
		}
		if len(s) < 2 {
			return false, "", text.ErrInvalidQuotedString
		}
		c := s[0]
		z, err := Dequote(s, c)
		if err != nil {
			return false, "", err
		}
		if c == '\'' || c == '"' {
			return true, z, nil
		}
		if c != '`' {
			return false, "", text.ErrInvalidQuotedString
		}
		if !exec {
			return true, z, nil
		}
		res, err := Exec(z)
		if err != nil {
			return false, "", err
		}
		return true, res, nil
	}
}
