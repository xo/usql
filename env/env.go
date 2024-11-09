// Package env contains runtime environment variables for usql, along with
// various helper funcs to determine the user's configuration.
package env

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/kenshaw/rasterm"
	"github.com/xo/dburl/passfile"
	"github.com/xo/usql/text"
)

// vars are environment variables.
var vars *Variables

func init() {
	vars = NewDefaultVars()
}

// Vars returns the environment variables.
func Vars() *Variables {
	return vars
}

// Get returns a standard variable.
func Get(name string) string {
	value, _ := vars.Get(name)
	return value
}

// Getenv tries retrieving successive keys from os environment variables.
func Getenv(keys ...string) (string, bool) {
	m := make(map[string]string)
	for _, v := range os.Environ() {
		if i := strings.Index(v, "="); i != -1 {
			m[v[:i]] = v[i+1:]
		}
	}
	for _, key := range keys {
		if v, ok := m[key]; ok {
			return v, true
		}
	}
	return "", false
}

// Chdir changes the current working directory to the specified path, or to the
// user's home directory if path is not specified.
func Chdir(u *user.User, path string) error {
	if path != "" {
		path = passfile.Expand(u.HomeDir, path)
	} else {
		path = u.HomeDir
	}
	return os.Chdir(path)
}

// OpenFile opens a file for read (os.O_RDONLY), returning the full, expanded
// path of the file. Callers are responsible for closing the returned file.
func OpenFile(u *user.User, path string) (string, *os.File, error) {
	path, err := filepath.EvalSymlinks(passfile.Expand(u.HomeDir, path))
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

// EditFile edits a file. If path is empty, then a temporary file will be
// created.
func EditFile(u *user.User, path, line string, buf []byte) ([]byte, error) {
	ed, _ := vars.Get("EDITOR")
	switch {
	case ed == "":
		if ed, _ = exec.LookPath("vi"); ed == "" {
			return nil, text.ErrNoEditorDefined
		}
	case path != "":
		path = passfile.Expand(u.HomeDir, path)
	default:
		f, err := os.CreateTemp("", text.CommandLower()+".*.sql")
		if err != nil {
			return nil, err
		}
		path = f.Name()
		if _, err = f.Write(append(bytes.TrimSuffix(buf, lineend), '\n')); err != nil {
			f.Close()
			return nil, err
		}
		if err = f.Close(); err != nil {
			return nil, err
		}
	}
	// setup args
	args := []string{path}
	if line != "" {
		if s, ok := Getenv(text.CommandUpper() + "_EDITOR_LINENUMBER_ARG"); ok {
			args = append(args, s+line)
		} else {
			args = append(args, "+"+line)
		}
	}
	// create command
	c := exec.Command(ed, args...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := c.Run(); err != nil {
		return nil, err
	}
	// read
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf, lineend), nil
}

// HistoryFile returns the path to the history file.
//
// Defaults to ~/.<command name>_history, overridden by environment variable
// <COMMAND NAME>_HISTORY (ie, ~/.usql_history and USQL_HISTORY).
func HistoryFile(u *user.User) string {
	n := text.CommandUpper() + "_HISTORY"
	path := "~/." + strings.ToLower(n)
	if s, ok := Getenv(n); ok {
		path = s
	}
	return passfile.Expand(u.HomeDir, path)
}

// RCFile returns the path to the RC file.
//
// Defaults to ~/.<command name>rc, overridden by environment variable
// <COMMAND NAME>RC (ie, ~/.usqlrc and USQLRC).
func RCFile(u *user.User) string {
	n := text.CommandUpper() + "RC"
	path := "~/." + strings.ToLower(n)
	if s, ok := Getenv(n); ok {
		path = s
	}
	return passfile.Expand(u.HomeDir, path)
}

// Getshell returns the user's defined SHELL, or system default (if found on
// path) and the appropriate command-line argument for the returned shell.
//
// Looks at the SHELL environment variable first, and then COMSPEC/ComSpec on
// Windows. Defaults to sh on non-Windows systems, and to cmd.exe on Windows.
func Getshell() (string, string) {
	shell, ok := Getenv("SHELL")
	param := "-c"
	if !ok && runtime.GOOS == "windows" {
		shell, _ = Getenv("COMSPEC", "ComSpec")
		param = "/c"
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
func Pipe(stdout, stderr io.Writer, c string) (io.WriteCloser, *exec.Cmd, error) {
	shell, param := Getshell()
	if shell == "" {
		return nil, nil, text.ErrNoShellAvailable
	}
	cmd := exec.Command(shell, param, c)
	cmd.Stdout, cmd.Stderr = stdout, stderr
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
	buf = bytes.TrimSuffix(buf, lineend)
	buf = bytes.TrimSuffix(buf, []byte{'\r'})
	return string(buf), nil
}

// Unquote unquotes a string.
func Unquote(s string) (string, error) {
	switch n := len(s); {
	case n == 0:
		return "", nil
	case n < 2, s[n-1] != s[0], s[0] != '\'' && s[0] != '"' && s[0] != '`':
		return "", text.ErrUnterminatedQuotedString
	}
	quote := s[0]
	s = s[1 : len(s)-1]
	if quote == '\'' {
		s = cleanDoubleRE.ReplaceAllString(s, `$1\'`)
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

// Untick returns a func that unquotes and unticks strings for the user.
//
// When exec is true, backtick'd strings will be executed using the provided
// user's shell (see Exec).
func Untick(u *user.User, v *Variables, exec bool) func(string, bool) (string, bool, error) {
	return func(s string, isvar bool) (string, bool, error) {
		// fmt.Fprintf(os.Stderr, "untick: %q\n", s)
		switch {
		case isvar:
			value, ok := v.Get(s)
			return value, ok, nil
		case len(s) < 2:
			return "", false, text.ErrInvalidQuotedString
		}
		c := s[0]
		z, err := Unquote(s)
		switch {
		case err != nil:
			return "", false, err
		case c == '\'', c == '"':
			return z, true, nil
		case c != '`':
			return "", false, text.ErrInvalidQuotedString
		case !exec:
			return z, true, nil
		}
		res, err := Exec(z)
		if err != nil {
			return "", false, err
		}
		return res, true, nil
	}
}

// Quote quotes a string.
func Quote(s string) string {
	s = strconv.QuoteToGraphic(s)
	return "'" + s[1:len(s)-1] + "'"
}

// TermGraphics returns the [rasterm.TermType] based on TERM_GRAPHICS
// environment variable.
func TermGraphics() rasterm.TermType {
	var typ rasterm.TermType
	s, _ := vars.Get("TERM_GRAPHICS")
	_ = typ.UnmarshalText([]byte(s))
	return typ
}

// ValidIdentifier returns an error when n is not a valid identifier.
func ValidIdentifier(n string) error {
	r := []rune(n)
	rlen := len(r)
	if rlen < 1 {
		return text.ErrInvalidIdentifier
	}
	for i := 0; i < rlen; i++ {
		if c := r[i]; c != '_' && !unicode.IsLetter(c) && !unicode.IsNumber(c) {
			return text.ErrInvalidIdentifier
		}
	}
	return nil
}

func ParseBool(value, name string) (string, error) {
	switch strings.ToLower(value) {
	case "1", "t", "tr", "tru", "true", "on":
		return "on", nil
	case "0", "f", "fa", "fal", "fals", "false", "of", "off":
		return "off", nil
	}
	return "", fmt.Errorf(text.FormatFieldInvalidValue, value, name, "Boolean")
}

func ParseKeywordBool(value, name string, keywords ...string) (string, error) {
	v := strings.ToLower(value)
	switch v {
	case "1", "t", "tr", "tru", "true", "on":
		return "on", nil
	case "0", "f", "fa", "fal", "fals", "false", "of", "off":
		return "off", nil
	}
	for _, k := range keywords {
		if v == k {
			return v, nil
		}
	}
	return "", fmt.Errorf(text.FormatFieldInvalid, value, name)
}

// lineend is the line ending.
var lineend = []byte{'\n'}

// cleanDoubleRE matches double quotes.
var cleanDoubleRE = regexp.MustCompile(`(^|[^\\])''`)
