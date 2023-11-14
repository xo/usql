//go:build ignore

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/xo/dburl"
	"github.com/yookoala/realpath"
)

type DriverInfo struct {
	// Tag is the build Tag / name of the directory the driver lives in.
	Tag string
	// Driver is the Go SQL Driver Driver (parsed from the import tagged with //
	// DRIVER: <Driver>), otherwise same as the tag / directory Driver.
	Driver string
	// Pkg is the imported driver package, taken from the import tagged with
	// DRIVER.
	Pkg string
	// Desc is the descriptive text of the driver, parsed from doc comment, ie,
	// "Package <tag> defines and registers usql's <Desc>."
	Desc string
	// URL is the driver's reference URL, parsed from doc comment's "See: <URL>".
	URL string
	// CGO is whether or not the driver requires CGO, based on presence of
	// 'Requires CGO.' in the comment
	CGO bool
	// Aliases are the parsed Alias: entries.
	Aliases [][]string
	// Wire indicates it is a Wire compatible driver.
	Wire bool
	// Group is the build Group
	Group string
}

// baseDrivers are drivers included in a build with no build tags listed.
var baseDrivers = map[string]DriverInfo{}

// mostDrivers are drivers included with the most tag. Populated below.
var mostDrivers = map[string]DriverInfo{}

// allDrivers are drivers forced to 'all' build tag.
var allDrivers = map[string]DriverInfo{}

// badDrivers are drivers forced to 'bad' build tag.
var badDrivers = map[string]DriverInfo{}

// wireDrivers are the wire compatible drivers.
var wireDrivers = map[string]DriverInfo{}

func main() {
	licenseStart := flag.Int("license-start", 2016, "license start year")
	licenseAuthor := flag.String("license-author", "Kenneth Shaw", "license author")
	dburlGen := flag.Bool("dburl-gen", false, "enable dburl generation")
	dburlDir := flag.String("dburl-dir", getDburlDir(), "dburl dir")
	dburlLicenseStart := flag.Int("dburl-license-start", 2015, "dburl license start year")
	flag.Parse()
	if err := run(*licenseStart, *licenseAuthor, *dburlGen, *dburlDir, *dburlLicenseStart); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(licenseStart int, licenseAuthor string, dburlGen bool, dburlDir string, dburlLicenseStart int) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := loadDrivers(filepath.Join(wd, "drivers")); err != nil {
		return err
	}
	if err := writeInternal(filepath.Join(wd, "internal"), baseDrivers, mostDrivers, allDrivers, badDrivers); err != nil {
		return err
	}
	if err := writeReadme(wd, true); err != nil {
		return err
	}
	if err := writeLicenseFiles(licenseStart, licenseAuthor); err != nil {
		return err
	}
	if dburlGen {
		if err := writeReadme(dburlDir, false); err != nil {
			return err
		}
		if err := writeDburlLicense(dburlDir, dburlLicenseStart, licenseAuthor); err != nil {
			return err
		}
	}
	return nil
}

func getDburlDir() string {
	dir := filepath.Join(os.Getenv("GOPATH"), "src/github.com/xo/dburl")
	var err error
	if dir, err = realpath.Realpath(dir); err != nil {
		panic(err)
	}
	return dir
}

var dirRE = regexp.MustCompile(`^([^/]+)/([^\./]+)\.go$`)

// loadDrivers loads the driver descriptions.
func loadDrivers(wd string) error {
	skipDirs := []string{"completer", "metadata"}
	err := fs.WalkDir(os.DirFS(wd), ".", func(n string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir():
			return nil
		}
		m := dirRE.FindAllStringSubmatch(n, -1)
		if m == nil || m[0][1] != m[0][2] || slices.Contains(skipDirs, m[0][1]) {
			return nil
		}
		tag, dest := m[0][1], mostDrivers
		driver, err := parseDriverInfo(tag, filepath.Join(wd, n))
		switch {
		case err != nil:
			return err
		case driver.Group == "base":
			dest = baseDrivers
		case driver.Group == "most":
		case driver.Group == "all":
			dest = allDrivers
		case driver.Group == "bad":
			dest = badDrivers
		default:
			return fmt.Errorf("driver %s has invalid group %q", tag, driver.Group)
		}
		dest[tag] = driver
		if dest[tag].Aliases != nil {
			for _, alias := range dest[tag].Aliases {
				wireDrivers[alias[0]] = DriverInfo{
					Tag:    tag,
					Driver: alias[0],
					Pkg:    dest[tag].Pkg,
					Desc:   alias[1],
					Wire:   true,
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func parseDriverInfo(tag, filename string) (DriverInfo, error) {
	f, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.ParseComments)
	if err != nil {
		return DriverInfo{}, err
	}
	name := tag
	var pkg string
	for _, imp := range f.Imports {
		if imp.Comment == nil || len(imp.Comment.List) == 0 || !strings.Contains(imp.Comment.List[0].Text, "DRIVER") {
			continue
		}
		pkg = imp.Path.Value[1 : len(imp.Path.Value)-1]
		if i := strings.Index(imp.Comment.List[0].Text, ":"); i != -1 {
			name = strings.TrimSpace(imp.Comment.List[0].Text[i+1:])
		}
		break
	}
	// parse doc comment
	comment := f.Doc.Text()
	prefix := "Package " + tag + " defines and registers usql's "
	if !strings.HasPrefix(comment, prefix) {
		return DriverInfo{}, fmt.Errorf("invalid doc comment prefix for driver %q", tag)
	}
	desc := strings.TrimPrefix(comment, prefix)
	i := strings.Index(desc, " driver.")
	if i == -1 {
		return DriverInfo{}, fmt.Errorf("cannot find description suffix for driver %q", tag)
	}
	desc = strings.TrimSpace(desc[:i])
	if desc == "" {
		return DriverInfo{}, fmt.Errorf("unable to parse description for driver %q", tag)
	}
	// parse alias:
	var aliases [][]string
	aliasesm := aliasRE.FindAllStringSubmatch(comment, -1)
	for _, m := range aliasesm {
		s := strings.Split(m[1], ",")
		aliases = append(aliases, []string{
			strings.TrimSpace(s[0]),
			strings.TrimSpace(s[1]),
		})
	}
	// parse see: url
	urlm := seeRE.FindAllStringSubmatch(comment, -1)
	if urlm == nil {
		return DriverInfo{}, fmt.Errorf("missing See: <URL> for driver %q", tag)
	}
	// parse group:
	group := "most"
	if groupm := groupRE.FindAllStringSubmatch(comment, -1); groupm != nil {
		group = strings.TrimSpace(groupm[0][1])
	}
	return DriverInfo{
		Tag:     tag,
		Driver:  name,
		Pkg:     pkg,
		Desc:    cleanRE.ReplaceAllString(desc, ""),
		URL:     strings.TrimSpace(urlm[0][1]),
		CGO:     strings.Contains(cleanRE.ReplaceAllString(comment, ""), "Requires CGO."),
		Aliases: aliases,
		Group:   group,
	}, nil
}

var (
	aliasRE = regexp.MustCompile(`(?m)^Alias:\s+(.*)$`)
	seeRE   = regexp.MustCompile(`(?m)^See:\s+(.*)$`)
	groupRE = regexp.MustCompile(`(?m)^Group:\s+(.*)$`)
	cleanRE = regexp.MustCompile(`[\r\n]`)
)

func writeInternal(wd string, drivers ...map[string]DriverInfo) error {
	// build known build tags
	var known []DriverInfo
	for _, m := range drivers {
		for _, v := range m {
			known = append(known, v)
		}
	}
	sort.Slice(known, func(i, j int) bool {
		return known[i].Tag < known[j].Tag
	})
	knownStr := ""
	for _, v := range known {
		knownStr += fmt.Sprintf("\n%q: %q, // %s", v.Tag, v.Driver, v.Pkg)
	}
	// format and write internal.go
	buf, err := format.Source([]byte(fmt.Sprintf(internalGo, knownStr)))
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(wd, "internal.go"), buf, 0o644); err != nil {
		return err
	}
	// write <tag>.go
	for _, v := range known {
		var tags string
		switch v.Group {
		case "base":
			tags = "(!no_base || " + v.Tag + ")"
		case "most":
			tags = "(all || most || " + v.Tag + ")"
		case "all":
			tags = "(all || " + v.Tag + ")"
		case "bad":
			tags = "(bad || " + v.Tag + ")"
		default:
			panic(v.Tag)
		}
		tags += " && !no_" + v.Tag
		buf, err := format.Source([]byte(fmt.Sprintf(internalTagGo, tags, "github.com/xo/usql/drivers/"+v.Tag, v.Desc)))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(wd, v.Tag+".go"), buf, 0o644); err != nil {
			return err
		}
	}
	return nil
}

const internalTagGo = `//go:build %s
package internal

// Code generated by gen.go. DO NOT EDIT.

import (
       _ %q // %s driver
)`

const internalGo = `// Package internal provides a way to obtain information about which database
// drivers were included at build.
package internal

// Code generated by gen.go. DO NOT EDIT.

// KnownBuildTags returns a map of known driver names to its respective build
// tags.
func KnownBuildTags() map[string]string{
	return map[string]string{%s
	}
}`

const (
	driverTableStart = "<!-- DRIVER DETAILS START -->"
	driverTableEnd   = "<!-- DRIVER DETAILS END -->"
)

func writeReadme(dir string, includeTagSummary bool) error {
	readme := filepath.Join(dir, "README.md")
	buf, err := os.ReadFile(readme)
	if err != nil {
		return err
	}
	start := bytes.Index(buf, []byte(driverTableStart))
	end := bytes.Index(buf, []byte(driverTableEnd))
	if start == -1 || end == -1 {
		return errors.New("unable to find driver table start/end in README.md")
	}
	b := new(bytes.Buffer)
	if _, err := b.Write(append(buf[:start+len(driverTableStart)], '\n')); err != nil {
		return err
	}
	if _, err := b.Write([]byte(buildDriverTable(includeTagSummary))); err != nil {
		return err
	}
	if _, err := b.Write(buf[end:]); err != nil {
		return err
	}
	return os.WriteFile(readme, b.Bytes(), 0o644)
}

func buildDriverTable(includeTagSummary bool) string {
	hdr := []string{"Database", "Scheme / Tag", "Scheme Aliases", "Driver Package / Notes"}
	widths := []int{len(hdr[0]), len(hdr[1]), len(hdr[2]), len(hdr[3])}
	baseRows, widths := buildRows(baseDrivers, widths)
	mostRows, widths := buildRows(mostDrivers, widths)
	allRows, widths := buildRows(allDrivers, widths)
	badRows, widths := buildRows(badDrivers, widths)
	wireRows, widths := buildRows(wireDrivers, widths)
	s := tableRows(widths, ' ', hdr)
	s += tableRows(widths, '-')
	s += tableRows(widths, ' ', baseRows...)
	s += tableRows(widths, ' ')
	s += tableRows(widths, ' ', mostRows...)
	s += tableRows(widths, ' ')
	s += tableRows(widths, ' ', allRows...)
	s += tableRows(widths, ' ')
	s += tableRows(widths, ' ', wireRows...)
	s += tableRows(widths, ' ')
	s += tableRows(widths, ' ', badRows...)
	if includeTagSummary {
		s += tableRows(widths, ' ')
		s += tableRows(widths, ' ',
			[]string{"**NO DRIVERS**", "`no_base`", "", "_no base drivers (useful for development)_"},
			[]string{"**MOST DRIVERS**", "`most`", "", "_all stable drivers_"},
			[]string{"**ALL DRIVERS**", "`all`", "", "_all drivers, excluding bad drivers_"},
			[]string{"**BAD DRIVERS**", "`bad`", "", "_bad drivers (broken/non-working drivers)_"},
			[]string{"**NO &lt;TAG&gt;**", "`no_<tag>`", "", "_exclude driver with `<tag>`_"},
		)
	}
	return s + "\n" + buildTableLinks(baseDrivers, mostDrivers, allDrivers, badDrivers)
}

var baseOrder = map[string]int{
	"postgres":   0,
	"mysql":      1,
	"sqlserver":  2,
	"oracle":     3,
	"sqlite3":    4,
	"clickhouse": 5,
	"csvq":       6,
}

func buildRows(m map[string]DriverInfo, widths []int) ([][]string, []int) {
	var drivers []DriverInfo
	for _, v := range m {
		drivers = append(drivers, v)
	}
	sort.Slice(drivers, func(i, j int) bool {
		switch {
		case drivers[i].Group == "base":
			return baseOrder[drivers[i].Driver] < baseOrder[drivers[j].Driver]
		}
		return strings.ToLower(drivers[i].Desc) < strings.ToLower(drivers[j].Desc)
	})
	var rows [][]string
	for i, v := range drivers {
		notes := ""
		if v.CGO {
			notes += " <sup>[†][f-cgo]</sup>"
		}
		if v.Wire {
			notes += " <sup>[‡][f-wire]</sup>"
		}
		rows = append(rows, []string{
			v.Desc,
			"`" + v.Tag + "`",
			buildAliases(v),
			fmt.Sprintf("[%s][d-%s]%s", v.Pkg, v.Tag, notes),
		})
		// calc max
		for j := 0; j < len(rows[i]); j++ {
			widths[j] = max(runewidth.StringWidth(rows[i][j]), widths[j])
		}
	}
	return rows, widths
}

func buildAliases(v DriverInfo) string {
	name := v.Tag
	if v.Wire {
		name = v.Driver
	}
	_, aliases := dburl.SchemeDriverAndAliases(name)
	if v.Wire {
		aliases = append(aliases, name)
	}
	for i := 0; i < len(aliases); i++ {
		if !v.Wire && aliases[i] == v.Tag {
			aliases[i] = v.Driver
		}
	}
	fileTypes := dburl.FileTypes()
	if slices.Contains(fileTypes, name) {
		aliases = append(aliases, `file`)
	}
	if len(aliases) > 0 {
		return "`" + strings.Join(aliases, "`, `") + "`"
	}
	return ""
}

func tableRows(widths []int, c rune, rows ...[]string) string {
	padding := string(c)
	if len(rows) == 0 {
		rows = [][]string{make([]string, len(widths))}
	}
	var s string
	for _, row := range rows {
		for i := 0; i < len(row); i++ {
			s += "|" + padding + row[i] + strings.Repeat(padding, widths[i]-runewidth.StringWidth(row[i])) + padding
		}
		s += "|\n"
	}
	return s
}

func buildTableLinks(drivers ...map[string]DriverInfo) string {
	var d []DriverInfo
	for _, m := range drivers {
		for _, v := range m {
			d = append(d, v)
		}
	}
	sort.Slice(d, func(i, j int) bool {
		return d[i].Tag < d[j].Tag
	})
	var s string
	for _, v := range d {
		s += fmt.Sprintf("[d-%s]: %s\n", v.Tag, v.URL)
	}
	return s
}

func writeLicenseFiles(licenseStart int, licenseAuthor string) error {
	s := fmt.Sprintf(license, licenseStart, time.Now().Year(), licenseAuthor)
	if err := os.WriteFile("LICENSE", append([]byte(s), '\n'), 0o644); err != nil {
		return err
	}
	textGo := fmt.Sprintf(licenseTextGo, s)
	if err := os.WriteFile("text/license.go", []byte(textGo), 0o644); err != nil {
		return err
	}
	return nil
}

func writeDburlLicense(dir string, licenseStart int, licenseAuthor string) error {
	s := fmt.Sprintf(license, licenseStart, time.Now().Year(), licenseAuthor)
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), append([]byte(s), '\n'), 0o644); err != nil {
		return err
	}
	return nil
}

const license = `The MIT License (MIT)

Copyright (c) %d-%d %s

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`

const licenseTextGo = `package text

// Code generated by gen.go. DO NOT EDIT.

// License contains the license text for usql.
const License = ` + "`%s`" + `
`
