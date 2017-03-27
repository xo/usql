package env

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	// ErrNoEditorDefined is the no editor defined error.
	ErrNoEditorDefined = errors.New("no editor defined")
)

// EditFile edits a file. If path is empty, then a temporary file will be created.
func EditFile(path, line, s string) ([]rune, error) {
	var err error

	ed := getenv("USQL_EDITOR", "EDITOR", "VISUAL")
	if ed == "" {
		return nil, ErrNoEditorDefined
	}

	if path == "" {
		var f *os.File
		f, err = ioutil.TempFile("", "usql")
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
		args = append(args, "+"+line)
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
