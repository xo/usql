package env

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/knq/dburl"

	"github.com/knq/usql/text"
)

// getenv tries retrieving successive keys from os environment variables.
func getenv(keys ...string) string {
	for _, key := range keys {
		if s := os.Getenv(key); s != "" {
			return s
		}
	}

	return ""
}

// Expand expands the tilde (~) in the front of a path to a the supplied
// directory.
func Expand(path string, dir string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(dir, strings.TrimPrefix(path, "~/"))
	}

	return path
}

// HistoryFile returns the path to the history file.
//
// Defaults to ~/.<command name>_history, overridden by environment variable
// <COMMAND NAME>_HISTORY (ie, ~/.usql_history and USQL_HISTORY).
func HistoryFile(u *user.User) string {
	n := text.CommandUpper() + "_HISTORY"
	path := "~/." + strings.ToLower(n)
	if s := getenv(n); s != "" {
		path = s
	}

	return Expand(path, u.HomeDir)
}

// RCFile returns the path to the RC file.
//
// Defaults to ~/.<command name>rc, overridden by environment variable
// <COMMAND NAME>RC (ie, ~/.usqlrc and USQLRC).
func RCFile(u *user.User) string {
	n := text.CommandUpper() + "RC"
	path := "~/." + strings.ToLower(n)
	if s := getenv(n); s != "" {
		path = s
	}

	return Expand(path, u.HomeDir)
}

// PassFile returns the path to the password file.
//
// Defaults to ~/.<command name>pass, overridden by environment variable
// <COMMAND NAME>PASS (ie, ~/.usqlpass and USQLPASS).
func PassFile(u *user.User) string {
	n := text.CommandUpper() + "PASS"
	path := "~/." + strings.ToLower(n)
	if s := getenv(n); s != "" {
		path = s
	}

	return Expand(path, u.HomeDir)
}

// EditFile edits a file. If path is empty, then a temporary file will be created.
func EditFile(path, line, s string) ([]rune, error) {
	var err error

	ed := getenv(text.CommandUpper()+"_EDITOR", "EDITOR", "VISUAL")
	if ed == "" {
		return nil, text.ErrNoEditorDefined
	}

	if path == "" {
		var f *os.File
		f, err = ioutil.TempFile("", text.CommandLower())
		if err != nil {
			return nil, err
		}

		err = f.Close()
		if err != nil {
			return nil, err
		}

		path = f.Name()
		err = ioutil.WriteFile(path, []byte(strings.TrimSuffix(s, "\n")+"\n"), 0644)
		if err != nil {
			return nil, err
		}
	}

	// setup args
	args := []string{path}
	if line != "" {
		prefix := getenv(text.CommandUpper() + "_EDITOR_LINENUMBER_ARG")
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
	err = c.Run()
	if err != nil {
		return nil, err
	}

	// read
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return []rune(strings.TrimSuffix(string(buf), "\n")), nil
}

// PassFileEntry reads
func PassFileEntry(v *dburl.URL, u *user.User) (*url.Userinfo, error) {
	// check if v already has password defined ...
	var username string
	if v.User != nil {
		username = v.User.Username()
		if _, ok := v.User.Password(); ok {
			return nil, nil
		}
	}

	// check if pass file exists
	path := PassFile(u)
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

	// read pass file entries
	entries, err := readPassEntries(path)
	if err != nil {
		return nil, err
	}

	// find matching entry
	n := strings.Split(v.Normalize(":", "", 3), ":")
	if len(n) < 3 {
		return nil, errors.New("unknown error encountered normalizing URL")
	}
	for _, entry := range entries {
		if u, p, ok := matchPassEntry(n, entry); ok {
			if u == "*" {
				u = username
			}
			return url.UserPassword(u, p), nil
		}
	}

	return nil, nil
}

var commentRE = regexp.MustCompile(`#.*`)

// readPassEntries reads the pass file entries from path.
func readPassEntries(path string) ([][]string, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries [][]string
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

		entries = append(entries, v)
	}

	return entries, nil
}

// matchPassEntry takes a normalized n, and a password entry along with the
// read username and pass, and determines if all of the components in n match entry.
func matchPassEntry(n, entry []string) (string, string, bool) {
	for i := 0; i < len(n); i++ {
		if entry[i] != "*" && entry[i] != n[i] {
			return "", "", false
		}
	}

	return entry[4], entry[5], true
}
