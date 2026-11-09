package dump

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// Flag is an option of a dump. The command line takes it as --NAME, after
// --dump, and the meta command \dump takes it as -NAME. Both read this table,
// so that the two have the same options.
type Flag struct {
	// Name is the name of the option, such as schema-only.
	Name string
	// Value names the value of the option, such as PATTERN. It is empty for a
	// switch, which takes no value.
	Value string
	// Repeat is set when the option can be given more than once.
	Repeat bool
	// Usage is the description of the option.
	Usage string

	// set sets the option in o.
	set func(o *Options, value string)
}

// Set sets the option in o. A switch reads value as a Boolean, and an empty
// value turns it on.
func (f Flag) Set(o *Options, value string) error {
	if f.Value == "" {
		if value == "" {
			value = "true"
		}
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("%s: %w: %q", f.Name, ErrNotBoolean, value)
		}
		value = ""
		if b {
			value = "true"
		}
	}
	f.set(o, value)
	return nil
}

// Flags are the options of a dump, in the order that the help lists them.
var Flags = []Flag{
	{
		Name:  "target",
		Value: "DRIVER",
		Usage: "write the dump for another database, such as duckdb",
		set:   func(o *Options, v string) { o.Target = v },
	},
	{
		Name:  "schema-only",
		Usage: "dump the schema and no rows",
		set:   func(o *Options, v string) { o.SchemaOnly = v != "" },
	},
	{
		Name:  "data-only",
		Usage: "dump the rows and no schema",
		set:   func(o *Options, v string) { o.DataOnly = v != "" },
	},
	{
		Name:   "schema",
		Value:  "PATTERN",
		Repeat: true,
		Usage:  "dump the schemas that match PATTERN",
		set:    func(o *Options, v string) { o.Schemas = append(o.Schemas, v) },
	},
	{
		Name:   "exclude-schema",
		Value:  "PATTERN",
		Repeat: true,
		Usage:  "do not dump the schemas that match PATTERN",
		set:    func(o *Options, v string) { o.ExcludeSchemas = append(o.ExcludeSchemas, v) },
	},
	{
		Name:   "table",
		Value:  "PATTERN",
		Repeat: true,
		Usage:  "dump the tables that match PATTERN",
		set:    func(o *Options, v string) { o.Tables = append(o.Tables, v) },
	},
	{
		Name:   "exclude-table",
		Value:  "PATTERN",
		Repeat: true,
		Usage:  "do not dump the tables that match PATTERN",
		set:    func(o *Options, v string) { o.ExcludeTables = append(o.ExcludeTables, v) },
	},
	{
		Name:  "allow-lossy",
		Usage: "allow a type of the target that holds less than the source type",
		set:   func(o *Options, v string) { o.Lossy = v != "" },
	},
	{
		Name:  "allow-text",
		Usage: "allow the string type of the target for a type that it does not have",
		set:   func(o *Options, v string) { o.Text = v != "" },
	},
	{
		Name:  "restrict-key",
		Value: "KEY",
		Usage: `the key of the \restrict line of the dump`,
		set:   func(o *Options, v string) { o.RestrictKey = v },
	},
}

// LookupFlag returns the option of a dump that has the name.
func LookupFlag(name string) (Flag, bool) {
	for _, f := range Flags {
		if f.Name == name {
			return f, true
		}
	}
	return Flag{}, false
}

// WriteFile writes a dump to the file name with write. The file is written
// whole or not at all: the dump goes to a temporary file in the same
// directory, which replaces name only when write returns no error. The file
// can be read by its owner only, because it holds the rows of the database.
func WriteFile(name string, write func(io.Writer) error) error {
	f, err := os.CreateTemp(filepath.Dir(name), "."+filepath.Base(name)+".*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	err = write(f)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp, name)
	}
	if err != nil {
		_ = os.Remove(tmp)
	}
	return err
}
