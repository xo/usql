package env

import (
	"fmt"
	"io"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/xo/dburl"
	"github.com/xo/usql/text"
)

// The sections of the configuration. A setting has the name SECTION.NAME, as
// in print.border.
const (
	SectionUsql        = "usql"
	SectionPrint       = "print"
	SectionVars        = "vars"
	SectionCharts      = "charts"
	SectionEnv         = "env"
	SectionConnections = "connections"
	SectionAliases     = "aliases"
)

// Sections are the sections of the configuration, in the order that \config
// shows them.
var Sections = []string{
	SectionUsql,
	SectionPrint,
	SectionVars,
	SectionCharts,
	SectionEnv,
	SectionConnections,
	SectionAliases,
}

// SplitName splits the name of a setting into its section and its name. A
// name that is only a section returns an empty name. A name with no section
// fails, and the error names the section that holds the name, when one does.
func (v *Variables) SplitName(name string) (string, string, error) {
	section, key, ok := strings.Cut(name, ".")
	switch {
	case !slices.Contains(Sections, section) && !ok:
		if s := v.home(name); s != "" {
			return "", "", fmt.Errorf(text.ConfigUseFullName, name, s+"."+name)
		}
		return "", "", fmt.Errorf(text.ConfigUnknownName, name)
	case !slices.Contains(Sections, section):
		return "", "", fmt.Errorf(text.ConfigUnknownSection, section)
	case ok && key == "":
		return "", "", fmt.Errorf(text.ConfigUnknownName, name)
	}
	return section, key, nil
}

// home returns the section that holds the setting name, or an empty string.
func (v *Variables) home(name string) string {
	if _, ok := v.uset[name]; ok {
		return SectionUsql
	}
	if _, ok := v.prnt[name]; ok {
		return SectionPrint
	}
	if _, ok := v.charts[name]; ok {
		return SectionCharts
	}
	return ""
}

// SetConfig sets the setting name, as in print.border, to value. Each section
// reads the value as its own command does: print as \pset, usql as \uset,
// vars as \set, connections as \cset and env as \setenv. Only the
// configuration file sets an alias.
func (v *Variables) SetConfig(name, value string) error {
	section, key, err := v.SplitName(name)
	switch {
	case err != nil:
		return err
	case key == "":
		return fmt.Errorf(text.ConfigNeedsName, section)
	}
	switch section {
	case SectionUsql:
		return v.SetSetting(key, value)
	case SectionPrint:
		_, err := v.SetPrint(key, value)
		return err
	case SectionVars:
		return v.Set(key, value)
	case SectionCharts:
		return v.SetChart(key, value)
	case SectionEnv:
		return v.SetEnv(key, value)
	case SectionConnections:
		return v.SetConn(key, value)
	}
	return text.ErrAliasConfigOnly
}

// SetOption applies an option of the command line: NAME=VALUE, as -O gives
// it, where NAME is SECTION.NAME. A value of print or connections between
// quotes loses them. A NAME with no value has the meaning that its section
// gives it on the command line of psql: a script variable is unset, a print
// setting switches, a connection is removed, a switch of usql turns on, and a
// variable of the environment is removed.
func (v *Variables) SetOption(opt string) error {
	name, value, ok := strings.Cut(opt, "=")
	section, key, err := v.SplitName(name)
	switch {
	case err != nil:
		return err
	case key == "":
		return fmt.Errorf(text.ConfigNeedsName, section)
	case section == SectionAliases:
		return text.ErrAliasConfigOnly
	case ok && (section == SectionPrint || section == SectionConnections) && len(value) > 1 && (value[0] == '\'' || value[0] == '"'):
		if value, err = Unquote(value); err != nil {
			return err
		}
	}
	switch {
	case ok:
		return v.SetConfig(name, value)
	case section == SectionVars:
		return v.Unset(key)
	case section == SectionPrint:
		_, err := v.TogglePrint(key, "")
		return err
	case section == SectionConnections:
		return v.SetConn(key)
	case section == SectionEnv:
		return v.UnsetEnv(key)
	}
	return v.SetConfig(name, "")
}

// SetChart sets a setting of \chart. path is the only one.
func (v *Variables) SetChart(name, value string) error {
	if _, ok := v.charts[name]; !ok {
		return fmt.Errorf(text.UnknownSetting, SectionCharts+"."+name)
	}
	v.charts[name] = value
	return nil
}

// Chart returns a setting of \chart.
func (v *Variables) Chart(name string) string {
	return v.charts[name]
}

// SetEnv sets a variable of the environment of the programs that usql
// starts. The environment of usql does not change.
func (v *Variables) SetEnv(name, value string) error {
	if name == "" || strings.ContainsAny(name, "=\x00") {
		return fmt.Errorf(text.InvalidEnvName, name)
	}
	v.env[name] = value
	delete(v.envUnset, name)
	return nil
}

// UnsetEnv removes a variable from the environment of the programs that usql
// starts.
func (v *Variables) UnsetEnv(name string) error {
	if name == "" || strings.ContainsAny(name, "=\x00") {
		return fmt.Errorf(text.InvalidEnvName, name)
	}
	delete(v.env, name)
	v.envUnset[name] = true
	return nil
}

// Getenv returns a variable of the environment that the programs of usql
// see: the setting of env, or else the variable of the environment of usql.
func (v *Variables) Getenv(name string) (string, bool) {
	if value, ok := v.env[name]; ok {
		return value, true
	}
	if v.envUnset[name] {
		return "", false
	}
	return os.LookupEnv(name)
}

// Environ returns the environment of a program that usql starts: the
// environment of usql, with the settings of env.
func (v *Variables) Environ() []string {
	var environ []string
	for _, s := range os.Environ() {
		name, _, _ := strings.Cut(s, "=")
		if _, ok := v.env[name]; !ok && !v.envUnset[name] {
			environ = append(environ, s)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(v.env)) {
		environ = append(environ, name+"="+v.env[name])
	}
	return environ
}

// SetAlias sets an alias: the meta command \name runs the text. Only the
// configuration file sets an alias.
func (v *Variables) SetAlias(name, value string) error {
	if err := ValidIdentifier(name); err != nil {
		return err
	}
	v.aliases[name] = value
	return nil
}

// Alias returns the text of an alias.
func (v *Variables) Alias(name string) (string, bool) {
	value, ok := v.aliases[name]
	return value, ok
}

// AliasNames returns the names of the aliases, in order.
func (v *Variables) AliasNames() []string {
	return v.names(SectionAliases)
}

// Config returns the value of a setting as \config shows it, and whether the
// setting is set. The password of a connection URL is hidden unless password
// is set.
func (v *Variables) Config(section, name string, password bool) (string, bool) {
	switch section {
	case SectionUsql:
		if _, ok := v.uset[name]; ok {
			return v.settingText(name), true
		}
	case SectionPrint:
		if _, ok := v.prnt[name]; ok {
			return v.printText(name), true
		}
	case SectionVars:
		if value, ok := v.vars[name]; ok {
			return Quote(value), true
		}
	case SectionCharts:
		if value, ok := v.charts[name]; ok {
			return strconv.QuoteToASCII(value), true
		}
	case SectionEnv:
		if value, ok := v.env[name]; ok {
			return Quote(value), true
		}
	case SectionConnections:
		if vals, ok := v.conn[name]; ok {
			s := strings.Join(vals, " ")
			if !password {
				s = strings.Join(hidePasswords(vals), " ")
			}
			return Quote(s), true
		}
	case SectionAliases:
		if value, ok := v.aliases[name]; ok {
			return Quote(value), true
		}
	}
	return "", false
}

// names returns the names of the settings of a section, in order.
func (v *Variables) names(section string) []string {
	switch section {
	case SectionUsql:
		return slices.Sorted(maps.Keys(v.uset))
	case SectionPrint:
		return slices.Sorted(maps.Keys(v.prnt))
	case SectionVars:
		return slices.Sorted(maps.Keys(v.vars))
	case SectionCharts:
		return slices.Sorted(maps.Keys(v.charts))
	case SectionEnv:
		return slices.Sorted(maps.Keys(v.env))
	case SectionConnections:
		return slices.Sorted(maps.Keys(v.conn))
	case SectionAliases:
		return slices.Sorted(maps.Keys(v.aliases))
	}
	return nil
}

// ConfigNames returns the full names of the settings that \config sets, and
// the sections, for completion.
func (v *Variables) ConfigNames() []string {
	names := slices.Clone(Sections)
	for _, section := range Sections[:len(Sections)-1] {
		for _, name := range v.names(section) {
			names = append(names, section+"."+name)
		}
	}
	return names
}

// DumpConfig writes the settings of the sections to w, as one table of full
// names and values. It writes the sections usql and print when no section is
// given. The password of a connection URL is hidden unless
// password is set.
func (v *Variables) DumpConfig(w io.Writer, password bool, sections ...string) error {
	if len(sections) == 0 {
		sections = []string{SectionUsql, SectionPrint}
	}
	return v.dump(w, true, password, sections...)
}

// dump writes the settings of the sections to w, as one table of names and
// values. A name has its section in front of it when full is set.
func (v *Variables) dump(w io.Writer, full, password bool, sections ...string) error {
	var rows [][2]string
	width := 0
	for _, section := range sections {
		for _, name := range v.names(section) {
			value, _ := v.Config(section, name, password)
			if full {
				name = section + "." + name
			}
			rows = append(rows, [2]string{name, value})
			width = max(width, len(name))
		}
	}
	for _, row := range rows {
		if _, err := fmt.Fprintln(w, strings.TrimRight(fmt.Sprintf("%-*s %s", width, row[0], row[1]), " ")); err != nil {
			return err
		}
	}
	return nil
}

// hidePasswords returns the values of a connection with the password of a URL
// written as xxxxx, as \conninfo writes it.
func hidePasswords(vals []string) []string {
	vals = slices.Clone(vals)
	for i, s := range vals {
		u, err := url.Parse(s)
		if err != nil || u.User == nil {
			continue
		}
		if _, ok := u.User.Password(); ok {
			vals[i] = u.Redacted()
		}
	}
	return vals
}

// LoadConfig reads the configuration file name. Each setting goes through
// the setter of its section, and prints nothing. It returns an error for each
// setting that it cannot set, and goes on with the next one. An error names
// the file, the line and the setting. isCommand reports whether a meta command
// of usql has a name, which an alias cannot take.
func (v *Variables) LoadConfig(name string, isCommand func(string) bool) []error {
	buf, err := os.ReadFile(name)
	if err != nil {
		return []error{err}
	}
	f, err := parser.ParseBytes(buf, 0)
	if err != nil {
		return []error{fmt.Errorf("%s: %w", name, err)}
	}
	if len(f.Docs) == 0 || f.Docs[0].Body == nil {
		return nil
	}
	l := &loader{v: v, file: name, isCommand: isCommand}
	pairs, ok := mapping(f.Docs[0].Body)
	if !ok {
		l.fail(f.Docs[0].Body, "", text.ErrConfigNotAMap)
		return l.errs
	}
	for _, pair := range pairs {
		key := keyText(pair.Key)
		switch {
		case slices.Contains(Sections, key):
			l.section(key, pair.Value)
		case key == "init":
			dir, _ := ConfigDir()
			l.fail(pair.Key, key, fmt.Errorf(text.ConfigInitMoved, filepath.Join(dir, "init.sql")))
		case key == "charts_path":
			l.fail(pair.Key, key, fmt.Errorf(text.ConfigMoveTo, SectionCharts+".path"))
		case v.home(key) != "":
			l.fail(pair.Key, key, fmt.Errorf(text.ConfigMoveTo, v.home(key)+"."+key))
		default:
			l.fail(pair.Key, key, fmt.Errorf(text.ConfigUnknownSection, key))
		}
	}
	return l.errs
}

// loader reads the settings of a configuration file.
type loader struct {
	v         *Variables
	file      string
	isCommand func(string) bool
	errs      []error
}

// fail adds the error of the setting name, at the line of node.
func (l *loader) fail(node ast.Node, name string, err error) {
	line := 0
	if tok := node.GetToken(); tok != nil && tok.Position != nil {
		line = tok.Position.Line
	}
	l.errs = append(l.errs, &ConfigError{File: l.file, Line: line, Name: name, Err: err})
}

// section sets the settings of a section.
func (l *loader) section(section string, node ast.Node) {
	if _, ok := node.(*ast.NullNode); ok {
		return
	}
	pairs, ok := mapping(node)
	if !ok {
		l.fail(node, section, text.ErrConfigNotAMap)
		return
	}
	for _, pair := range pairs {
		key := keyText(pair.Key)
		name := section + "." + key
		var err error
		switch section {
		case SectionConnections:
			err = l.connection(key, pair.Value)
		case SectionAliases:
			err = l.alias(key, pair.Value)
		default:
			var value string
			if value, err = scalar(pair.Value); err == nil {
				err = l.v.SetConfig(name, value)
			}
		}
		if err != nil {
			l.fail(pair.Key, name, err)
		}
	}
}

// connection sets a named connection, from a URL, from a list of a driver and
// its parameters, or from a map of the parts of a URL.
func (l *loader) connection(name string, node ast.Node) error {
	switch n := unwrap(node).(type) {
	case *ast.SequenceNode:
		vals := make([]string, len(n.Values))
		for i, value := range n.Values {
			var err error
			if vals[i], err = scalar(value); err != nil {
				return err
			}
		}
		return l.v.SetConn(name, vals...)
	case *ast.MappingNode, *ast.MappingValueNode:
		var m map[string]any
		if err := yaml.NodeToValue(n, &m); err != nil {
			return err
		}
		urlstr, err := dburl.BuildURL(m)
		if err != nil {
			return err
		}
		return l.v.SetConn(name, urlstr)
	}
	value, err := scalar(node)
	if err != nil {
		return err
	}
	return l.v.SetConn(name, value)
}

// alias sets an alias, which cannot take the name of a meta command.
func (l *loader) alias(name string, node ast.Node) error {
	value, err := scalar(node)
	switch {
	case err != nil:
		return err
	case l.isCommand != nil && l.isCommand(name):
		return fmt.Errorf(text.AliasIsACommand, name)
	}
	return l.v.SetAlias(name, value)
}

// ConfigError is an error of a setting in the configuration file.
type ConfigError struct {
	File string
	Line int
	Name string
	Err  error
}

// Error satisfies the error interface.
func (e *ConfigError) Error() string {
	var b strings.Builder
	b.WriteString(e.File)
	if e.Line != 0 {
		b.WriteString(":" + strconv.Itoa(e.Line))
	}
	if e.Name != "" {
		b.WriteString(": " + e.Name)
	}
	return b.String() + ": " + e.Err.Error()
}

// Unwrap returns the error of the setting.
func (e *ConfigError) Unwrap() error {
	return e.Err
}

// mapping returns the pairs of a map.
func mapping(node ast.Node) ([]*ast.MappingValueNode, bool) {
	switch n := unwrap(node).(type) {
	case *ast.MappingNode:
		return n.Values, true
	case *ast.MappingValueNode:
		return []*ast.MappingValueNode{n}, true
	}
	return nil, false
}

// unwrap returns the node that an anchor or a tag holds.
func unwrap(node ast.Node) ast.Node {
	for {
		switch n := node.(type) {
		case *ast.AnchorNode:
			node = n.Value
		case *ast.TagNode:
			node = n.Value
		default:
			return node
		}
	}
}

// keyText returns the text of a key, as it is written: null is the name null
// and not a null.
func keyText(key ast.MapKeyNode) string {
	if s, ok := key.(*ast.StringNode); ok {
		return s.Value
	}
	if tok := key.GetToken(); tok != nil {
		return tok.Value
	}
	return key.String()
}

// scalar returns the text of a value as it is written, so that on stays on
// and 0123 keeps its zero. A null is empty.
func scalar(node ast.Node) (string, error) {
	switch n := unwrap(node).(type) {
	case nil, *ast.NullNode:
		return "", nil
	case *ast.StringNode:
		return n.Value, nil
	case *ast.LiteralNode:
		return n.Value.Value, nil
	case *ast.IntegerNode, *ast.FloatNode, *ast.BoolNode, *ast.InfinityNode, *ast.NanNode:
		return n.GetToken().Value, nil
	}
	return "", text.ErrConfigNotAValue
}
