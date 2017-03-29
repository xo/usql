package env

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/knq/usql/text"
)

var (
	// ErrNoEditorDefined is the no editor defined error.
	ErrNoEditorDefined = errors.New("no editor defined")
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
func HistoryFile(u *user.User) string {
	n := text.CommandUpper() + "_HISTORY"
	path := "~/." + strings.ToLower(n)
	if s := getenv(n); s != "" {
		path = s
	}

	return Expand(path, u.HomeDir)
}

// RCFile returns the path to the RC file.
func RCFile(u *user.User) string {
	n := text.CommandUpper() + "RC"
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
		return nil, ErrNoEditorDefined
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
