package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xo/dburl"
)

// seeRE matches the See: line in a driver's package comment.
var seeRE = regexp.MustCompile(`(?m)^See:\s+(.*)$`)

// TestDriverSeeMatchesDburl checks that every driver's See: line is the same
// URL that dburl records for the scheme.
//
// gen.go stopped reading See: when the driver metadata moved to dburl, so
// nothing else compares the two. Without this the lines rot silently: the
// Hive driver's See: was corrected by hand when the driver was replaced, and
// nothing would have caught it if it had not been.
func TestDriverSeeMatchesDburl(t *testing.T) {
	t.Parallel()
	schemes := schemeIndex()
	var checked int
	for _, tag := range driverTags(t) {
		scheme, ok := schemes[tag]
		if !ok {
			t.Errorf("driver %s has no dburl scheme.\n"+
				"Every driver needs one, because gen.go reads the driver's\n"+
				"package, description and URL from it.", tag)
			continue
		}
		see, ok := driverSee(t, tag)
		if !ok {
			t.Errorf("driver %s has no See: line in its package comment.\n"+
				"It should read: See: %s", tag, scheme.DriverURL)
			continue
		}
		checked++
		if see != scheme.DriverURL {
			t.Errorf("driver %s says See: %s and dburl says %s.\n"+
				"dburl owns this URL. Change it there and copy it here, or\n"+
				"delete the See: line.", tag, see, scheme.DriverURL)
		}
	}
	// A test that checks nothing passes. Guard the count as well as the
	// values, so that a walk which stops finding drivers fails.
	if checked < 40 {
		t.Errorf("compared %d See: lines against dburl, expected at least 40.\n"+
			"The driver walk is probably not finding them any more.", checked)
	}
}

// TestDriverSeeIsNotDuplicated checks that no two drivers claim the same
// upstream, which would mean one of them names the wrong project.
func TestDriverSeeIsNotDuplicated(t *testing.T) {
	t.Parallel()
	seen := make(map[string]string)
	for _, tag := range driverTags(t) {
		see, ok := driverSee(t, tag)
		if !ok {
			continue
		}
		if other, dup := seen[see]; dup {
			t.Errorf("drivers %s and %s both say See: %s.\n"+
				"Two drivers cannot share one upstream project.", other, tag, see)
			continue
		}
		seen[see] = tag
	}
}

// schemeIndex indexes dburl's schemes by driver name and by alias, so that a
// build tag resolves even when it differs from the scheme, as dynamodb does
// against godynamo.
func schemeIndex() map[string]dburl.Scheme {
	m := make(map[string]dburl.Scheme)
	for _, scheme := range dburl.BaseSchemes() {
		m[scheme.Driver] = scheme
		for _, alias := range scheme.Aliases {
			if _, ok := m[alias]; !ok {
				m[alias] = scheme
			}
		}
	}
	return m
}

// driverTags returns the build tag of every driver, which is the name of its
// directory under drivers/.
func driverTags(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir("drivers")
	if err != nil {
		t.Fatalf("reading drivers: %v", err)
	}
	// These hold a file named after the directory but are support packages
	// rather than drivers. gen.go skips the same two.
	skip := map[string]bool{"completer": true, "metadata": true}
	var tags []string
	for _, entry := range entries {
		if !entry.IsDir() || skip[entry.Name()] {
			continue
		}
		tag := entry.Name()
		// A driver is a directory holding a file of the same name.
		if _, err := os.Stat(filepath.Join("drivers", tag, tag+".go")); err != nil {
			continue
		}
		tags = append(tags, tag)
	}
	return tags
}

// driverSee returns the URL on a driver's See: line.
func driverSee(t *testing.T, tag string) (string, bool) {
	t.Helper()
	name := filepath.Join("drivers", tag, tag+".go")
	f, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parsing %s: %v", name, err)
	}
	m := seeRE.FindStringSubmatch(f.Doc.Text())
	if m == nil {
		return "", false
	}
	return strings.TrimSpace(m[1]), true
}
