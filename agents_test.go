package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// TestClaudeImportsAgents holds D1. AGENTS.md holds the rules, because Codex
// and the other agents read it, and CLAUDE.md imports it, so that Claude Code
// reads the same rules. A rule written in CLAUDE.md would reach Claude Code
// alone. A symbolic link would not do, because a Windows checkout writes a
// link as a small text file.
func TestClaudeImportsAgents(t *testing.T) {
	t.Parallel()
	info, err := os.Lstat("CLAUDE.md")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("CLAUDE.md is a symbolic link. Make it a file that holds @AGENTS.md. See D1")
	}
	if got := strings.TrimSpace(readFile(t, "CLAUDE.md")); got != "@AGENTS.md" {
		t.Errorf("CLAUDE.md holds %q. It holds only @AGENTS.md, and the rules go in"+
			" AGENTS.md. See D1", got)
	}
}

// TestTheRootHoldsFourDocuments holds rule 11 of AGENTS.md. A document that
// appears in the root is one nobody filed.
func TestTheRootHoldsFourDocuments(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{
		"README.md": true, "AGENTS.md": true, "CLAUDE.md": true, "CONTRIBUTING.md": true,
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if !allowed[e.Name()] {
			t.Errorf("%s is in the repository root. Only README, AGENTS, CLAUDE and"+
				" CONTRIBUTING belong there, and everything else goes in docs/. See D1.", e.Name())
		}
	}
	for _, name := range slices.Sorted(maps.Keys(allowed)) {
		if _, err := os.Stat(name); err != nil {
			t.Errorf("expected %s in the repository root", name)
		}
	}
}

// TestSkillsAreCopies holds D1. Each skill in skills-lock.json is an ordinary
// folder under .agents/skills and under .claude/skills, and the two folders
// hold the same files.
//
// A symbolic link is refused because a Windows checkout writes one as a text
// file that holds the target path. Claude Code then finds a file where it
// expects a folder, and it loads no skill and reports nothing.
func TestSkillsAreCopies(t *testing.T) {
	t.Parallel()
	var lock struct {
		Skills map[string]json.RawMessage `json:"skills"`
	}
	if err := json.Unmarshal([]byte(readFile(t, "skills-lock.json")), &lock); err != nil {
		t.Fatalf("reading skills-lock.json: %v", err)
	}
	if len(lock.Skills) == 0 {
		t.Fatal("skills-lock.json names no skill")
	}
	roots := []string{filepath.Join(".agents", "skills"), filepath.Join(".claude", "skills")}
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			t.Errorf("%s: %v. Install the skills with the command in CONTRIBUTING.md", root, err)
			continue
		}
		for _, e := range entries {
			if _, ok := lock.Skills[e.Name()]; !ok {
				t.Errorf("%s is not in skills-lock.json, so nobody can install it again. "+
					"Add it with the command in CONTRIBUTING.md", filepath.Join(root, e.Name()))
			}
		}
	}
	for _, name := range slices.Sorted(maps.Keys(lock.Skills)) {
		agents := skillFiles(t, filepath.Join(roots[0], name))
		claude := skillFiles(t, filepath.Join(roots[1], name))
		if agents == nil || claude == nil {
			continue
		}
		for _, path := range slices.Sorted(maps.Keys(agents)) {
			switch other, ok := claude[path]; {
			case !ok:
				t.Errorf("%s: %s is in %s and not in %s", name, path, roots[0], roots[1])
			case other != agents[path]:
				t.Errorf("%s: %s differs between %s and %s", name, path, roots[0], roots[1])
			}
		}
		for _, path := range slices.Sorted(maps.Keys(claude)) {
			if _, ok := agents[path]; !ok {
				t.Errorf("%s: %s is in %s and not in %s", name, path, roots[1], roots[0])
			}
		}
	}
}

// skillFiles returns the content of every file in one copy of a skill, keyed
// by its path inside the copy. It returns nil after reporting a copy that is
// missing or that is a symbolic link.
func skillFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	fi, err := os.Lstat(dir)
	switch {
	case err != nil:
		t.Errorf("%s is missing. Install the skill with the command in CONTRIBUTING.md", dir)
		return nil
	case !fi.IsDir():
		t.Errorf("%s is not a folder. A Windows checkout writes a symbolic link as a text file, "+
			"so install the skill with --copy. See D1", dir)
		return nil
	}
	out := make(map[string]string)
	err = fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.Type()&fs.ModeSymlink != 0:
			t.Errorf("%s is a symbolic link. See D1", filepath.Join(dir, path))
		case d.Type().IsRegular():
			body, err := fs.ReadFile(os.DirFS(dir), path)
			if err != nil {
				return err
			}
			out[path] = string(body)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}
	return out
}

// decision is one file in docs/decisions.
type decision struct {
	num    string
	title  string
	status string
	file   string
}

// decisionFile names a decision file: D, the number in three digits, and the
// title in lower case words joined by hyphens.
var decisionFile = regexp.MustCompile(`^D(\d{3})-[a-z0-9-]+\.md$`)

// decisions reads every decision in docs/decisions, in order. Each opens with
// its number, its title and its status, as D2 decided:
//
//	# D1. Every xo repository is set up for coding agents the same way
//
//	Status: Decided.
func decisions(t *testing.T) []decision {
	t.Helper()
	dir := filepath.Join("docs", "decisions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	head := regexp.MustCompile(`\A# D(\d+)\. (.+)\n\nStatus: (.+)\.\n`)
	var out []decision
	for _, e := range entries {
		if e.Name() == "README.md" {
			continue
		}
		name := decisionFile.FindStringSubmatch(e.Name())
		if name == nil {
			t.Errorf("%s: a decision file is named D, three digits, a hyphen and the"+
				" title in lower case words, such as D002-a-large-project.md", e.Name())
			continue
		}
		m := head.FindStringSubmatch(readFile(t, filepath.Join(dir, e.Name())))
		if m == nil {
			t.Errorf("%s: a decision opens with \"# D<n>. <title>\", a blank line and"+
				" \"Status: <status>.\"", e.Name())
			continue
		}
		if n, _ := strconv.Atoi(name[1]); strconv.Itoa(n) != m[1] {
			t.Errorf("%s holds D%s. The file name and the heading name one decision.", e.Name(), m[1])
		}
		out = append(out, decision{num: m[1], title: m[2], status: m[3], file: e.Name()})
	}
	// A test that reads nothing passes, so guard the count.
	if len(out) < 2 {
		t.Fatalf("expected at least 2 decisions in %s, found %d", dir, len(out))
	}
	return out
}

// TestTheDecisionIndexIsComplete checks the table in docs/decisions/README.md
// against the decision files. A reader finds a decision by its number in that
// table, so a missing row or a stale status there hides it.
func TestTheDecisionIndexIsComplete(t *testing.T) {
	t.Parallel()
	index := readFile(t, filepath.Join("docs", "decisions", "README.md"))
	rows := make(map[string]string)
	for _, m := range regexp.MustCompile(`(?m)^\| \[D(\d+)\]\(.*$`).FindAllStringSubmatch(index, -1) {
		rows[m[1]] = m[0]
	}
	written := make(map[string]bool)
	for _, d := range decisions(t) {
		written[d.num] = true
		want := fmt.Sprintf("| [D%s](%s) | %s | %s |", d.num, d.file, d.title, d.status)
		switch got, ok := rows[d.num]; {
		case !ok:
			t.Errorf("D%s has no row in docs/decisions/README.md. Add:\n%s", d.num, want)
		case got != want:
			t.Errorf("D%s: the row in docs/decisions/README.md is\n%s\nand the file says\n%s", d.num, got, want)
		}
	}
	for _, num := range slices.Sorted(maps.Keys(rows)) {
		if !written[num] {
			t.Errorf("docs/decisions/README.md has a row for D%s, and no file holds it", num)
		}
	}
}

// TestAnAmendmentPointsBothWays checks that a decision which amends or
// replaces another says so in its status, and that the other one names it
// back. Each file opens with its status, so a reader who finds the older
// decision sees at once that it no longer holds as written. See D2.
func TestAnAmendmentPointsBothWays(t *testing.T) {
	t.Parallel()
	status := make(map[string]string)
	for _, d := range decisions(t) {
		status[d.num] = d.title + ". " + d.status
	}
	naming := regexp.MustCompile(`(?i)\b(?:amends|supersedes|superseded by|replaced by|amended by) D(\d+)`)
	for _, num := range slices.Sorted(maps.Keys(status)) {
		head := status[num]
		for _, m := range naming.FindAllStringSubmatch(head, -1) {
			other := m[1]
			if _, ok := status[other]; !ok {
				t.Errorf("D%s names D%s, which is not a decision", num, other)
				continue
			}
			if !strings.Contains(status[other], "D"+num) {
				t.Errorf("D%s says %q, and D%s does not mention D%s. An amendment has"+
					" to be visible from both sides. See D2.", num, head, other, num)
			}
		}
	}
}

// readFile returns the content of a file in the repository.
func readFile(t *testing.T, name string) string {
	t.Helper()
	body, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
