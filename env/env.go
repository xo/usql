// Package env contains runtime environment variables for usql, along with
// various helper funcs to determine the user's configuration.
package env

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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

// Setting returns a setting of the application, such as PAGER.
func Setting(name string) string {
	value, _ := vars.Setting(name)
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
//
// The editor runs through the shell, as psql runs it, so that the editor
// setting can hold arguments, such as code --wait.
func EditFile(u *user.User, path, line string, buf []byte) ([]byte, error) {
	ed, _ := vars.Setting("editor")
	if ed == "" {
		vi, err := exec.LookPath("vi")
		if err != nil {
			return nil, text.ErrNoEditorDefined
		}
		ed = shellQuote(vi)
	}
	if path != "" {
		path = passfile.Expand(u.HomeDir, path)
	} else {
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
	// the line goes before the path, as psql passes it
	cmd := ed
	if line != "" {
		if s, ok := Getenv(text.CommandUpper() + "_EDITOR_LINENUMBER_ARG"); ok {
			cmd += " " + shellQuote(s+line)
		} else {
			cmd += " " + shellQuote("+"+line)
		}
	}
	cmd += " " + shellQuote(path)
	shell, param := Getshell()
	if shell == "" {
		return nil, text.ErrNoShellAvailable
	}
	c := exec.Command(shell, param, cmd)
	c.Stdin, c.Stdout, c.Stderr, c.Env = os.Stdin, os.Stdout, os.Stderr, vars.Environ()
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

// shellQuote quotes s as one argument for the shell that Getshell returns.
func shellQuote(s string) string {
	if runtime.GOOS == "windows" {
		return `"` + s + `"`
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// HistoryFile returns the path of the history file: the file that
// USQL_HISTORY names, or else history in the state directory of usql. It
// returns an empty path when there is no such directory.
func HistoryFile(u *user.User) string {
	if s, ok := Getenv(text.CommandUpper() + "_HISTORY"); ok {
		return passfile.Expand(u.HomeDir, s)
	}
	dir, err := StateDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "history")
}

// InitFile returns the path of the file that runs at start: the file that
// USQLRC names, or else init.sql in the configuration directory of usql. It
// returns an empty path when there is no such directory.
func InitFile(u *user.User) string {
	if s, ok := Getenv(text.CommandUpper() + "RC"); ok {
		return passfile.Expand(u.HomeDir, s)
	}
	dir, err := ConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "init.sql")
}

// ConfigDir returns the configuration directory of usql:
// $XDG_CONFIG_HOME/usql on Linux and other Unix systems, which is
// ~/.config/usql when XDG_CONFIG_HOME is not set, ~/Library/Application
// Support/usql on macOS, and %AppData%\usql on Windows.
func ConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, text.CommandName), nil
}

// StateDir returns the directory of what usql keeps from one run to the next,
// such as its history: $XDG_STATE_HOME/usql on Linux and other Unix systems,
// which is ~/.local/state/usql when XDG_STATE_HOME is not set, the
// configuration directory on macOS, and %LocalAppData%\usql on Windows.
//
// The XDG Base Directory specification says to ignore a relative path in
// XDG_STATE_HOME, and so does StateDir.
func StateDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		dir := os.Getenv("LocalAppData")
		if dir == "" {
			return "", errors.New("%LocalAppData% is not defined")
		}
		return filepath.Join(dir, text.CommandName), nil
	case "darwin", "ios":
		return ConfigDir()
	}
	dir := os.Getenv("XDG_STATE_HOME")
	if !filepath.IsAbs(dir) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(dir, text.CommandName), nil
}

// Getshell returns the user's defined shell, or system default (if found on
// path) and the appropriate command-line argument for the returned shell.
//
// Looks at the USQL_SHELL and SHELL environment variables first, and then
// COMSPEC/ComSpec on Windows. Defaults to sh on non-Windows systems, and to
// cmd.exe on Windows. cmd.exe takes /c, and any other shell -c.
func Getshell() (string, string) {
	shell, ok := Getenv(text.CommandUpper()+"_SHELL", "SHELL")
	param := "-c"
	if ok && strings.TrimSuffix(strings.ToLower(filepath.Base(shell)), ".exe") == "cmd" {
		param = "/c"
	}
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

// Shell runs s as a shell. When s is empty the user's shell or COMSPEC is
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
	cmd.Stdin, cmd.Stdout, cmd.Stderr, cmd.Env = os.Stdin, os.Stdout, os.Stderr, vars.Environ()
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
	cmd.Stdout, cmd.Stderr, cmd.Env = stdout, stderr, vars.Environ()
	out, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	return out, cmd, cmd.Start()
}

// Exec executes s using the user's shell / COMSPEC with -c (or /c) and
// returning the captured output. See Getshell.
//
// When USQL_SHELL, SHELL or COMSPEC is not defined, then "sh" / "cmd.exe" will be used
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
	cmd := exec.Command(shell, param, s)
	cmd.Env = vars.Environ()
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	// remove ending \r\n
	buf = bytes.TrimSuffix(buf, lineend)
	buf = bytes.TrimSuffix(buf, []byte{'\r'})
	return string(buf), nil
}

// Unquote unquotes a string.
//
// A string between single quotes is read as psql reads the argument of a
// meta command: ” is a quote, \n, \t, \b, \r and \f are control
// characters, one to three octal digits and \x with one or two hex digits are
// a byte, and a backslash before any other character is that character, as in
// \q for q. A string between double quotes or backticks loses its quotes, and
// keeps the escapes of Go.
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
		return unquoteSingle(s)
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

// unquoteSingle decodes the text between single quotes as psql decodes the
// argument of a meta command.
func unquoteSingle(s string) (string, error) {
	buf := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\'' && i+1 < len(s) && s[i+1] == '\'':
			buf = append(buf, c)
			i++
		case c == '\'':
			return "", text.ErrInvalidQuotedString
		case c != '\\' || i+1 == len(s):
			buf = append(buf, c)
		default:
			i++
			switch c = s[i]; {
			case c == 'n':
				buf = append(buf, '\n')
			case c == 't':
				buf = append(buf, '\t')
			case c == 'b':
				buf = append(buf, '\b')
			case c == 'r':
				buf = append(buf, '\r')
			case c == 'f':
				buf = append(buf, '\f')
			case '0' <= c && c <= '7':
				// one to three octal digits
				v, j := 0, i
				for ; j < len(s) && j < i+3 && '0' <= s[j] && s[j] <= '7'; j++ {
					v = v*8 + int(s[j]-'0')
				}
				buf, i = append(buf, byte(v)), j-1
			case c == 'x' && i+1 < len(s) && isHex(s[i+1]):
				// one or two hex digits
				v, j := 0, i+1
				for ; j < len(s) && j < i+3 && isHex(s[j]); j++ {
					v = v*16 + hexValue(s[j])
				}
				buf, i = append(buf, byte(v)), j-1
			default:
				buf = append(buf, c)
			}
		}
	}
	return string(buf), nil
}

// isHex reports whether c is a hex digit.
func isHex(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

// hexValue returns the value of the hex digit c.
func hexValue(c byte) int {
	switch {
	case c <= '9':
		return int(c - '0')
	case c <= 'F':
		return int(c-'A') + 10
	}
	return int(c-'a') + 10
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
	s, _ := vars.Setting("term_graphics")
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

// ParseVariableBool reads value as a boolean, as psql's ParseVariableBool
// reads the expression of \if. It takes true, false, yes, no, on, off, 1 and
// 0, in any case. A prefix of a word is enough when it is unique, so t is
// true and n is no. A prefix of on or off needs two letters, because o starts
// both. It reports false when value is not a boolean.
func ParseVariableBool(value string) (v, ok bool) {
	switch {
	case hasPrefixFold("true", value, 1), hasPrefixFold("yes", value, 1), hasPrefixFold("on", value, 2), value == "1":
		return true, true
	case hasPrefixFold("false", value, 1), hasPrefixFold("no", value, 1), hasPrefixFold("off", value, 2), value == "0":
		return false, true
	}
	return false, false
}

// hasPrefixFold reports whether s is a prefix of word of at least n bytes,
// where word is in lower case. Only the ASCII letters of s fold, as in psql,
// so the long s (U+017F) does not match an s.
func hasPrefixFold(word, s string, n int) bool {
	if len(s) < n || len(s) > len(word) {
		return false
	}
	for i := range len(s) {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != word[i] {
			return false
		}
	}
	return true
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
