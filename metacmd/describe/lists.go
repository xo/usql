package describe

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/xo/dbmeta"
	"github.com/xo/usql/text"
)

// Call is one describe command as typed, such as \dT+ public.*.
type Call struct {
	// Name is the name of the command without S and +, such as dT.
	Name string
	// Patterns are the patterns that follow the name.
	Patterns []string
	// Verbose is set by +.
	Verbose bool
	// System is set by S, and lists the objects of the system too.
	System bool
}

// pattern returns the i'th pattern, or an empty string.
func (c Call) pattern(i int) string {
	if i < len(c.Patterns) {
		return c.Patterns[i]
	}
	return ""
}

// command runs one describe command.
type command func(context.Context, *Formatter, Call) error

// commands are the describe commands that print one list, keyed by name.
// Commands that print more than a list, such as \d NAME, are methods of
// Writer.
var commands = map[string]command{
	"dA":      accessMethods,
	"dAc":     byAccessMethod(operatorClasses),
	"dAf":     byAccessMethod(operatorFamilies),
	"dAo":     byAccessMethod(familyOperators),
	"dAp":     byAccessMethod(familyFunctions),
	"db":      tablespaces,
	"dc":      conversions,
	"dconfig": settings,
	"dC":      casts,
	"dd":      comments,
	"dD":      domains,
	"ddp":     defaultACLs,
	"des":     foreignServers,
	"det":     foreignTables,
	"deu":     userMappings,
	"dew":     foreignDataWrappers,
	"dF":      textSearchConfigs,
	"dFd":     textSearchDictionaries,
	"dFp":     textSearchParsers,
	"dFt":     textSearchTemplates,
	"dg":      roles,
	"dl":      largeObjects,
	"dL":      languages,
	"do":      operators,
	"dO":      collations,
	"dP":      partitionedTables,
	"drds":    roleSettings,
	"drg":     roleGrants,
	"dRp":     publications,
	"dRs":     subscriptions,
	"dT":      types,
	"du":      roles,
	"dx":      extensions,
	"dX":      extendedStats,
	"dy":      eventTriggers,
}

// Commands returns the names of the describe commands that Run answers.
func Commands() []string {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Run runs the describe command c. It returns false when no such command
// exists.
func (f *Formatter) Run(ctx context.Context, c Call) (bool, error) {
	cmd, ok := commands[c.Name]
	if !ok {
		return false, nil
	}
	return true, cmd(ctx, f, c)
}

// field is one column that a list prints, and how to read it from a row.
type field[T any] struct {
	name string
	// verbose is a column that only the + form of the command prints.
	verbose bool
	value   func(T) any
}

// spec is a describe command that prints one list: the rows of one query,
// under psql's title, in psql's columns.
type spec[T any] struct {
	cmd    string
	title  string
	q      *dbmeta.Query[T]
	fields []field[T]
}

// named returns the arguments for pattern. It is split into a schema and a
// name where the query takes a schema.
func (s spec[T]) named(f *Formatter, pattern string) dbmeta.Args {
	var a dbmeta.Args
	if f.takes(s.q, "schema") {
		a.Schema, a.Name = f.meta.Dialect().Pattern(pattern)
	} else {
		a.Name = f.fold(pattern)
	}
	return a
}

// read returns the rows of the query for a.
func (s spec[T]) read(ctx context.Context, f *Formatter, c Call, a dbmeta.Args) ([]T, error) {
	a.WithSystem = c.System
	res, err := all(ctx, f, s.q, a)
	if err != nil {
		return nil, f.check(s.cmd, err)
	}
	return res, nil
}

// print prints res.
func (s spec[T]) print(out io.Writer, res []T, verbose bool) error {
	r := &rows{}
	for _, f := range s.fields {
		if verbose || !f.verbose {
			r.columns = append(r.columns, f.name)
		}
	}
	for _, v := range res {
		var row []any
		for _, f := range s.fields {
			if verbose || !f.verbose {
				row = append(row, f.value(v))
			}
		}
		r.add(row...)
	}
	// psql orders each list by its leading columns
	slices.SortStableFunc(r.values, compareRows)
	return encode(out, r, s.title, nil, nil)
}

// list returns the command that prints s.
func list[T any](s spec[T]) command {
	return func(ctx context.Context, f *Formatter, c Call) error {
		res, err := s.read(ctx, f, c, s.named(f, c.pattern(0)))
		if err != nil {
			return err
		}
		return s.print(f.out, res, c.Verbose)
	}
}

// byAccessMethod returns the command that prints s, for the \dA commands,
// whose first pattern matches the access method and whose second matches
// the name.
func byAccessMethod[T any](s spec[T]) command {
	return func(ctx context.Context, f *Formatter, c Call) error {
		a := dbmeta.Args{AccessMethod: f.fold(c.pattern(0)), Name: f.fold(c.pattern(1))}
		res, err := s.read(ctx, f, c, a)
		if err != nil {
			return err
		}
		return s.print(f.out, res, c.Verbose)
	}
}

// compareRows orders two rows by their values in turn, numbers as numbers.
func compareRows(a, b []any) int {
	for i := range min(len(a), len(b)) {
		x, xok := a[i].(int64)
		y, yok := b[i].(int64)
		var n int
		if xok && yok {
			n = cmp.Compare(x, y)
		} else {
			n = cmp.Compare(fmt.Sprint(a[i]), fmt.Sprint(b[i]))
		}
		if n != 0 {
			return n
		}
	}
	return cmp.Compare(len(a), len(b))
}

// takes reports whether q takes the parameter named name.
func (f *Formatter) takes(q dbmeta.AnyQuery, name string) bool {
	params, err := q.Params(f.meta)
	if err != nil {
		return false
	}
	return slices.ContainsFunc(params, func(p dbmeta.Param) bool {
		return p.Name == name
	})
}

// fold returns a pattern for an object that no schema holds, such as a role,
// folded as the database folds a name that is not quoted.
func (f *Formatter) fold(pattern string) string {
	return like(f.meta.Dialect().FoldIdentifier(pattern))
}

// tf returns a boolean the way psql prints one.
func tf(b bool) string {
	if b {
		return "t"
	}
	return "f"
}

// yesNo returns a boolean the way psql prints one it computes, such as
// Default?.
func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// nullYesNo is yesNo, or an empty string for NULL.
func nullYesNo(b sql.Null[bool]) string {
	if !b.Valid {
		return ""
	}
	return yesNo(b.V)
}

// The columns that most lists share.
func description[T any](f func(T) sql.Null[string]) field[T] {
	return field[T]{name: "Description", value: func(v T) any { return null(f(v)) }}
}

func verboseDescription[T any](f func(T) sql.Null[string]) field[T] {
	return field[T]{name: "Description", verbose: true, value: func(v T) any { return null(f(v)) }}
}

func accessPrivileges[T any](f func(T) sql.Null[string]) field[T] {
	return field[T]{name: "Access privileges", verbose: true, value: func(v T) any { return null(f(v)) }}
}

var types = list(spec[dbmeta.Type]{
	cmd: `\dT`, title: text.ListOfTypes, q: dbmeta.Types,
	fields: []field[dbmeta.Type]{
		{name: "Schema", value: func(t dbmeta.Type) any { return t.Schema }},
		{name: "Name", value: func(t dbmeta.Type) any { return t.Name }},
		{name: "Internal name", verbose: true, value: func(t dbmeta.Type) any { return t.Internal }},
		{name: "Size", verbose: true, value: func(t dbmeta.Type) any { return null(t.Size) }},
		{name: "Elements", verbose: true, value: func(t dbmeta.Type) any { return t.Elements }},
		{name: "Owner", verbose: true, value: func(t dbmeta.Type) any { return null(t.Owner) }},
		accessPrivileges(func(t dbmeta.Type) sql.Null[string] { return t.Access }),
		description(func(t dbmeta.Type) sql.Null[string] { return t.Comment }),
	},
})

var domains = list(spec[dbmeta.Domain]{
	cmd: `\dD`, title: text.ListOfDomains, q: dbmeta.Domains,
	fields: []field[dbmeta.Domain]{
		{name: "Schema", value: func(d dbmeta.Domain) any { return d.Schema }},
		{name: "Name", value: func(d dbmeta.Domain) any { return d.Name }},
		{name: "Type", value: func(d dbmeta.Domain) any { return d.DataType }},
		{name: "Collation", value: func(d dbmeta.Domain) any { return d.Collation }},
		{name: "Nullable", value: func(d dbmeta.Domain) any { return notNull(d.Nullable) }},
		{name: "Default", value: func(d dbmeta.Domain) any { return null(d.Default) }},
		{name: "Check", value: func(d dbmeta.Domain) any { return d.Constraints }},
		accessPrivileges(func(d dbmeta.Domain) sql.Null[string] { return d.Access }),
		verboseDescription(func(d dbmeta.Domain) sql.Null[string] { return d.Comment }),
	},
})

// notNull returns what psql prints in the Nullable column.
func notNull(nullable bool) string {
	if nullable {
		return ""
	}
	return "not null"
}

var operators = list(spec[dbmeta.Operator]{
	cmd: `\do`, title: text.ListOfOperators, q: dbmeta.Operators,
	fields: []field[dbmeta.Operator]{
		{name: "Schema", value: func(o dbmeta.Operator) any { return o.Schema }},
		{name: "Name", value: func(o dbmeta.Operator) any { return o.Name }},
		{name: "Left arg type", value: func(o dbmeta.Operator) any { return o.LeftType }},
		{name: "Right arg type", value: func(o dbmeta.Operator) any { return o.RightType }},
		{name: "Result type", value: func(o dbmeta.Operator) any { return o.ResultType }},
		{name: "Function", verbose: true, value: func(o dbmeta.Operator) any { return null(o.Function) }},
		description(func(o dbmeta.Operator) sql.Null[string] { return o.Comment }),
	},
})

var casts = list(spec[dbmeta.Cast]{
	cmd: `\dC`, title: text.ListOfCasts, q: dbmeta.Casts,
	fields: []field[dbmeta.Cast]{
		{name: "Source type", value: func(c dbmeta.Cast) any { return c.Source }},
		{name: "Target type", value: func(c dbmeta.Cast) any { return c.Target }},
		{name: "Function", value: func(c dbmeta.Cast) any { return null(c.Function) }},
		{name: "Implicit?", value: func(c dbmeta.Cast) any { return c.Implicit }},
		{name: "Leakproof?", verbose: true, value: func(c dbmeta.Cast) any { return nullYesNo(c.LeakProof) }},
		verboseDescription(func(c dbmeta.Cast) sql.Null[string] { return c.Comment }),
	},
})

var languages = list(spec[dbmeta.Language]{
	cmd: `\dL`, title: text.ListOfLanguages, q: dbmeta.Languages,
	fields: []field[dbmeta.Language]{
		{name: "Name", value: func(l dbmeta.Language) any { return l.Name }},
		{name: "Owner", value: func(l dbmeta.Language) any { return l.Owner }},
		{name: "Trusted", value: func(l dbmeta.Language) any { return tf(l.Trusted) }},
		{name: "Internal language", verbose: true, value: func(l dbmeta.Language) any { return tf(l.Internal) }},
		{name: "Call handler", verbose: true, value: func(l dbmeta.Language) any { return l.Handler }},
		{name: "Validator", verbose: true, value: func(l dbmeta.Language) any { return l.Validator }},
		{name: "Inline handler", verbose: true, value: func(l dbmeta.Language) any { return l.Inline }},
		accessPrivileges(func(l dbmeta.Language) sql.Null[string] { return l.Access }),
		description(func(l dbmeta.Language) sql.Null[string] { return l.Comment }),
	},
})

var roles = list(spec[dbmeta.Role]{
	cmd: `\du`, title: text.ListOfRoles, q: dbmeta.Roles,
	fields: []field[dbmeta.Role]{
		{name: "Role name", value: func(r dbmeta.Role) any { return r.Name }},
		{name: "Attributes", value: func(r dbmeta.Role) any { return attributes(r) }},
		verboseDescription(func(r dbmeta.Role) sql.Null[string] { return r.Comment }),
	},
})

// attributes returns the attributes of a role, the way psql 16 and later
// prints them.
func attributes(r dbmeta.Role) string {
	var v []string
	add := func(ok bool, s string) {
		if ok {
			v = append(v, s)
		}
	}
	add(r.Superuser, "Superuser")
	add(!r.Inherit, "No inheritance")
	add(r.CreateRole, "Create role")
	add(r.CreateDB, "Create DB")
	add(!r.CanLogin, "Cannot login")
	add(r.Replication, "Replication")
	add(r.BypassRLS, "Bypass RLS")
	switch {
	case r.ConnLimit == 1:
		v = append(v, "1 connection")
	case r.ConnLimit >= 0:
		v = append(v, strconv.FormatInt(r.ConnLimit, 10)+" connections")
	}
	add(r.ValidUntil.Valid, "Password valid until "+r.ValidUntil.V)
	return strings.Join(v, ", ")
}

var roleSettings = func(ctx context.Context, f *Formatter, c Call) error {
	s := spec[dbmeta.RoleSetting]{
		cmd: `\drds`, title: text.ListOfRoleSettings, q: dbmeta.RoleSettings,
		fields: []field[dbmeta.RoleSetting]{
			{name: "Role", value: func(r dbmeta.RoleSetting) any { return r.Role }},
			{name: "Database", value: func(r dbmeta.RoleSetting) any { return r.Database }},
			{name: "Settings", value: func(r dbmeta.RoleSetting) any { return null(r.Settings) }},
		},
	}
	a := s.named(f, c.pattern(0))
	a.Database = f.fold(c.pattern(1))
	res, err := s.read(ctx, f, c, a)
	if err != nil {
		return err
	}
	return s.print(f.out, res, c.Verbose)
}

var roleGrants = list(spec[dbmeta.RoleGrant]{
	cmd: `\drg`, title: text.ListOfRoleGrants, q: dbmeta.RoleGrants,
	fields: []field[dbmeta.RoleGrant]{
		{name: "Role name", value: func(r dbmeta.RoleGrant) any { return r.Role }},
		{name: "Member of", value: func(r dbmeta.RoleGrant) any { return r.MemberOf }},
		{name: "Options", value: func(r dbmeta.RoleGrant) any { return grantOptions(r) }},
		{name: "Grantor", value: func(r dbmeta.RoleGrant) any { return null(r.Grantor) }},
	},
})

// grantOptions returns the options of a role grant, the way psql prints them.
func grantOptions(r dbmeta.RoleGrant) string {
	var v []string
	if r.Admin {
		v = append(v, "ADMIN")
	}
	if r.Inherit {
		v = append(v, "INHERIT")
	}
	if r.Set {
		v = append(v, "SET")
	}
	return strings.Join(v, ", ")
}

var defaultACLs = list(spec[dbmeta.DefaultACL]{
	cmd: `\ddp`, title: text.DefaultAccessPrivileges, q: dbmeta.DefaultACLs,
	fields: []field[dbmeta.DefaultACL]{
		{name: "Owner", value: func(d dbmeta.DefaultACL) any { return d.Owner }},
		{name: "Schema", value: func(d dbmeta.DefaultACL) any { return d.Schema }},
		{name: "Type", value: func(d dbmeta.DefaultACL) any { return d.Type }},
		{name: "Access privileges", value: func(d dbmeta.DefaultACL) any { return null(d.Access) }},
	},
})

var tablespaces = list(spec[dbmeta.Tablespace]{
	cmd: `\db`, title: text.ListOfTablespaces, q: dbmeta.Tablespaces,
	fields: []field[dbmeta.Tablespace]{
		{name: "Name", value: func(t dbmeta.Tablespace) any { return t.Name }},
		{name: "Owner", value: func(t dbmeta.Tablespace) any { return null(t.Owner) }},
		{name: "Location", value: func(t dbmeta.Tablespace) any { return null(t.Location) }},
		accessPrivileges(func(t dbmeta.Tablespace) sql.Null[string] { return t.Access }),
		{name: "Options", verbose: true, value: func(t dbmeta.Tablespace) any { return null(t.Options) }},
		{name: "Size", verbose: true, value: func(t dbmeta.Tablespace) any { return null(t.Size) }},
		verboseDescription(func(t dbmeta.Tablespace) sql.Null[string] { return t.Comment }),
	},
})

var accessMethods = list(spec[dbmeta.AccessMethod]{
	cmd: `\dA`, title: text.ListOfAccessMethods, q: dbmeta.AccessMethods,
	fields: []field[dbmeta.AccessMethod]{
		{name: "Name", value: func(a dbmeta.AccessMethod) any { return a.Name }},
		{name: "Type", value: func(a dbmeta.AccessMethod) any { return a.Type }},
		{name: "Handler", verbose: true, value: func(a dbmeta.AccessMethod) any { return null(a.Handler) }},
		verboseDescription(func(a dbmeta.AccessMethod) sql.Null[string] { return a.Comment }),
	},
})

var collations = list(spec[dbmeta.Collation]{
	cmd: `\dO`, title: text.ListOfCollations, q: dbmeta.Collations,
	fields: []field[dbmeta.Collation]{
		{name: "Schema", value: func(c dbmeta.Collation) any { return c.Schema }},
		{name: "Name", value: func(c dbmeta.Collation) any { return c.Name }},
		{name: "Provider", value: func(c dbmeta.Collation) any { return null(c.Provider) }},
		{name: "Collate", value: func(c dbmeta.Collation) any { return null(c.Collate) }},
		{name: "Ctype", value: func(c dbmeta.Collation) any { return null(c.CType) }},
		{name: "Locale", value: func(c dbmeta.Collation) any { return null(c.Locale) }},
		{name: "ICU Rules", value: func(c dbmeta.Collation) any { return null(c.Rules) }},
		{name: "Deterministic?", value: func(c dbmeta.Collation) any { return nullYesNo(c.Deterministic) }},
		verboseDescription(func(c dbmeta.Collation) sql.Null[string] { return c.Comment }),
	},
})

var conversions = list(spec[dbmeta.Conversion]{
	cmd: `\dc`, title: text.ListOfConversions, q: dbmeta.Conversions,
	fields: []field[dbmeta.Conversion]{
		{name: "Schema", value: func(c dbmeta.Conversion) any { return c.Schema }},
		{name: "Name", value: func(c dbmeta.Conversion) any { return c.Name }},
		{name: "Source", value: func(c dbmeta.Conversion) any { return c.Source }},
		{name: "Destination", value: func(c dbmeta.Conversion) any { return c.Target }},
		{name: "Default?", value: func(c dbmeta.Conversion) any { return yesNo(c.Default) }},
		verboseDescription(func(c dbmeta.Conversion) sql.Null[string] { return c.Comment }),
	},
})

var largeObjects = list(spec[dbmeta.LargeObject]{
	cmd: `\dl`, title: text.LargeObjects, q: dbmeta.LargeObjects,
	fields: []field[dbmeta.LargeObject]{
		{name: "ID", value: func(l dbmeta.LargeObject) any { return l.OID }},
		{name: "Owner", value: func(l dbmeta.LargeObject) any { return l.Owner }},
		accessPrivileges(func(l dbmeta.LargeObject) sql.Null[string] { return l.Access }),
		description(func(l dbmeta.LargeObject) sql.Null[string] { return l.Comment }),
	},
})

var settings = list(spec[dbmeta.Setting]{
	cmd: `\dconfig`, title: text.ListOfSettings, q: dbmeta.Settings,
	fields: []field[dbmeta.Setting]{
		{name: "Parameter", value: func(s dbmeta.Setting) any { return s.Name }},
		{name: "Value", value: func(s dbmeta.Setting) any { return cmp.Or(null(s.Display), null(s.Value)) }},
		{name: "Type", verbose: true, value: func(s dbmeta.Setting) any { return null(s.Type) }},
		{name: "Context", verbose: true, value: func(s dbmeta.Setting) any { return null(s.Context) }},
		accessPrivileges(func(s dbmeta.Setting) sql.Null[string] { return s.Access }),
	},
})

var comments = list(spec[dbmeta.Comment]{
	cmd: `\dd`, title: text.ObjectDescriptions, q: dbmeta.Comments,
	fields: []field[dbmeta.Comment]{
		{name: "Schema", value: func(c dbmeta.Comment) any { return c.Schema }},
		{name: "Name", value: func(c dbmeta.Comment) any { return c.Name }},
		{name: "Object", value: func(c dbmeta.Comment) any { return c.Type }},
		{name: "Description", value: func(c dbmeta.Comment) any { return c.Comment }},
	},
})

var eventTriggers = list(spec[dbmeta.EventTrigger]{
	cmd: `\dy`, title: text.ListOfEventTriggers, q: dbmeta.EventTriggers,
	fields: []field[dbmeta.EventTrigger]{
		{name: "Name", value: func(e dbmeta.EventTrigger) any { return e.Name }},
		{name: "Event", value: func(e dbmeta.EventTrigger) any { return e.Event }},
		{name: "Owner", value: func(e dbmeta.EventTrigger) any { return e.Owner }},
		{name: "Enabled", value: func(e dbmeta.EventTrigger) any { return e.Enabled }},
		{name: "Function", value: func(e dbmeta.EventTrigger) any { return e.Function }},
		{name: "Tags", value: func(e dbmeta.EventTrigger) any { return null(e.Tags) }},
		verboseDescription(func(e dbmeta.EventTrigger) sql.Null[string] { return e.Comment }),
	},
})

var extendedStats = list(spec[dbmeta.ExtendedStat]{
	cmd: `\dX`, title: text.ListOfExtendedStats, q: dbmeta.ExtendedStats,
	fields: []field[dbmeta.ExtendedStat]{
		{name: "Schema", value: func(e dbmeta.ExtendedStat) any { return e.Schema }},
		{name: "Name", value: func(e dbmeta.ExtendedStat) any { return null(e.Name) }},
		{name: "Definition", value: func(e dbmeta.ExtendedStat) any { return null(e.Definition) }},
		{name: "Ndistinct", value: func(e dbmeta.ExtendedStat) any { return defined(e.Ndistinct) }},
		{name: "Dependencies", value: func(e dbmeta.ExtendedStat) any { return defined(e.Dependencies) }},
		{name: "MCV", value: func(e dbmeta.ExtendedStat) any { return defined(e.MCV) }},
	},
})

// defined returns what psql prints for a kind of extended statistics.
func defined(b bool) string {
	if b {
		return "defined"
	}
	return ""
}

var foreignDataWrappers = list(spec[dbmeta.ForeignDataWrapper]{
	cmd: `\dew`, title: text.ListOfForeignDataWrappers, q: dbmeta.ForeignDataWrappers,
	fields: []field[dbmeta.ForeignDataWrapper]{
		{name: "Name", value: func(f dbmeta.ForeignDataWrapper) any { return f.Name }},
		{name: "Owner", value: func(f dbmeta.ForeignDataWrapper) any { return f.Owner }},
		{name: "Handler", value: func(f dbmeta.ForeignDataWrapper) any { return null(f.Handler) }},
		{name: "Validator", value: func(f dbmeta.ForeignDataWrapper) any { return null(f.Validator) }},
		accessPrivileges(func(f dbmeta.ForeignDataWrapper) sql.Null[string] { return f.Access }),
		{name: "FDW options", verbose: true, value: func(f dbmeta.ForeignDataWrapper) any { return null(f.Options) }},
		verboseDescription(func(f dbmeta.ForeignDataWrapper) sql.Null[string] { return f.Comment }),
	},
})

var foreignServers = list(spec[dbmeta.ForeignServer]{
	cmd: `\des`, title: text.ListOfForeignServers, q: dbmeta.ForeignServers,
	fields: []field[dbmeta.ForeignServer]{
		{name: "Name", value: func(f dbmeta.ForeignServer) any { return f.Name }},
		{name: "Owner", value: func(f dbmeta.ForeignServer) any { return f.Owner }},
		{name: "Foreign-data wrapper", value: func(f dbmeta.ForeignServer) any { return f.Wrapper }},
		accessPrivileges(func(f dbmeta.ForeignServer) sql.Null[string] { return f.Access }),
		{name: "Type", verbose: true, value: func(f dbmeta.ForeignServer) any { return null(f.Type) }},
		{name: "Version", verbose: true, value: func(f dbmeta.ForeignServer) any { return null(f.Version) }},
		{name: "FDW options", verbose: true, value: func(f dbmeta.ForeignServer) any { return null(f.Options) }},
		verboseDescription(func(f dbmeta.ForeignServer) sql.Null[string] { return f.Comment }),
	},
})

var userMappings = list(spec[dbmeta.UserMapping]{
	cmd: `\deu`, title: text.ListOfUserMappings, q: dbmeta.UserMappings,
	fields: []field[dbmeta.UserMapping]{
		{name: "Server", value: func(u dbmeta.UserMapping) any { return u.Server }},
		{name: "User name", value: func(u dbmeta.UserMapping) any { return null(u.Name) }},
		{name: "FDW options", verbose: true, value: func(u dbmeta.UserMapping) any { return null(u.Options) }},
	},
})

var foreignTables = list(spec[dbmeta.ForeignTable]{
	cmd: `\det`, title: text.ListOfForeignTables, q: dbmeta.ForeignTables,
	fields: []field[dbmeta.ForeignTable]{
		{name: "Schema", value: func(f dbmeta.ForeignTable) any { return f.Schema }},
		{name: "Table", value: func(f dbmeta.ForeignTable) any { return f.Name }},
		{name: "Server", value: func(f dbmeta.ForeignTable) any { return f.Server }},
		{name: "FDW options", verbose: true, value: func(f dbmeta.ForeignTable) any { return null(f.Options) }},
		verboseDescription(func(f dbmeta.ForeignTable) sql.Null[string] { return f.Comment }),
	},
})

var partitionedTables = list(spec[dbmeta.PartitionedTable]{
	cmd: `\dP`, title: text.ListOfPartitionedRels, q: dbmeta.PartitionedTables,
	fields: []field[dbmeta.PartitionedTable]{
		{name: "Schema", value: func(p dbmeta.PartitionedTable) any { return p.Schema }},
		{name: "Name", value: func(p dbmeta.PartitionedTable) any { return p.Name }},
		{name: "Owner", value: func(p dbmeta.PartitionedTable) any { return p.Owner }},
		{name: "Type", value: func(p dbmeta.PartitionedTable) any { return p.Type }},
		{name: "Parent name", value: func(p dbmeta.PartitionedTable) any { return p.Parent }},
		verboseDescription(func(p dbmeta.PartitionedTable) sql.Null[string] { return p.Comment }),
	},
})

// publicationList is the list that \dRp prints. \dRp+ prints each
// publication with its tables instead.
var publicationList = spec[dbmeta.Publication]{
	cmd: `\dRp`, title: text.ListOfPublications, q: dbmeta.Publications,
	fields: []field[dbmeta.Publication]{
		{name: "Name", value: func(p dbmeta.Publication) any { return p.Name }},
		{name: "Owner", value: func(p dbmeta.Publication) any { return p.Owner }},
		{name: "All tables", value: func(p dbmeta.Publication) any { return tf(p.AllTables) }},
		{name: "Inserts", value: func(p dbmeta.Publication) any { return tf(p.Insert) }},
		{name: "Updates", value: func(p dbmeta.Publication) any { return tf(p.Update) }},
		{name: "Deletes", value: func(p dbmeta.Publication) any { return tf(p.Delete) }},
		{name: "Truncates", value: func(p dbmeta.Publication) any { return tf(p.Truncate) }},
		{name: "Via root", value: func(p dbmeta.Publication) any { return tf(p.ViaRoot) }},
	},
}

func publications(ctx context.Context, f *Formatter, c Call) error {
	pubs, err := publicationList.read(ctx, f, c, publicationList.named(f, c.pattern(0)))
	switch {
	case err != nil:
		return err
	case !c.Verbose:
		return publicationList.print(f.out, pubs, false)
	}
	slices.SortStableFunc(pubs, func(a, b dbmeta.Publication) int {
		return cmp.Compare(a.Name, b.Name)
	})
	for _, p := range pubs {
		tables, err := optional(all(ctx, f, dbmeta.PublicationTables, dbmeta.Args{Name: p.Name}))
		if err != nil {
			return err
		}
		var lines []string
		for _, t := range tables {
			if t.Publication != p.Name {
				continue
			}
			s := qualified(t.Schema, t.Name)
			if t.Columns != "" {
				s += " (" + t.Columns + ")"
			}
			if t.Where.Valid {
				s += " WHERE " + t.Where.V
			}
			lines = append(lines, "    "+s)
		}
		slices.Sort(lines)
		footer := noFooter
		if len(lines) != 0 {
			footer = func(out io.Writer) error {
				_, err := fmt.Fprintln(out, "Tables:\n"+strings.Join(lines, "\n"))
				return err
			}
		}
		// the title names the publication, so the columns leave it out
		r := &rows{}
		var row []any
		for _, f := range publicationList.fields[1:] {
			r.columns = append(r.columns, f.name)
			row = append(row, f.value(p))
		}
		r.add(row...)
		if err := encode(f.out, r, fmt.Sprintf(text.Publication, p.Name), footer, nil); err != nil {
			return err
		}
	}
	return nil
}

var subscriptions = list(spec[dbmeta.Subscription]{
	cmd: `\dRs`, title: text.ListOfSubscriptions, q: dbmeta.Subscriptions,
	fields: []field[dbmeta.Subscription]{
		{name: "Name", value: func(s dbmeta.Subscription) any { return s.Name }},
		{name: "Owner", value: func(s dbmeta.Subscription) any { return s.Owner }},
		{name: "Enabled", value: func(s dbmeta.Subscription) any { return tf(s.Enabled) }},
		{name: "Publication", value: func(s dbmeta.Subscription) any { return s.Publications }},
		{name: "Synchronous commit", verbose: true, value: func(s dbmeta.Subscription) any { return null(s.Synchronous) }},
		{name: "Slot name", verbose: true, value: func(s dbmeta.Subscription) any { return null(s.Slot) }},
		verboseDescription(func(s dbmeta.Subscription) sql.Null[string] { return s.Comment }),
	},
})

// configList is the list that \dF prints. \dF+ prints the token map of each
// configuration instead.
var configList = spec[dbmeta.TextSearchConfig]{
	cmd: `\dF`, title: text.ListOfTSConfigs, q: dbmeta.TextSearchConfigs,
	fields: []field[dbmeta.TextSearchConfig]{
		{name: "Schema", value: func(t dbmeta.TextSearchConfig) any { return t.Schema }},
		{name: "Name", value: func(t dbmeta.TextSearchConfig) any { return t.Name }},
		description(func(t dbmeta.TextSearchConfig) sql.Null[string] { return t.Comment }),
	},
}

func textSearchConfigs(ctx context.Context, f *Formatter, c Call) error {
	configs, err := configList.read(ctx, f, c, configList.named(f, c.pattern(0)))
	switch {
	case err != nil:
		return err
	case !c.Verbose:
		return configList.print(f.out, configs, false)
	}
	slices.SortStableFunc(configs, func(a, b dbmeta.TextSearchConfig) int {
		return cmp.Or(cmp.Compare(a.Schema, b.Schema), cmp.Compare(a.Name, b.Name))
	})
	for _, t := range configs {
		maps, err := all(ctx, f, dbmeta.TextSearchConfigMaps, dbmeta.Args{Schema: t.Schema, Name: t.Name, WithSystem: true})
		if err != nil {
			return f.check(`\dF+`, err)
		}
		slices.SortStableFunc(maps, func(a, b dbmeta.TextSearchConfigMap) int {
			return cmp.Or(cmp.Compare(a.Token, b.Token), cmp.Compare(a.Position, b.Position))
		})
		// one row for each dictionary, joined into one line for each token
		r := &rows{columns: []string{"Token", "Dictionaries"}}
		for _, m := range maps {
			if m.Config != t.Name || m.Schema != t.Schema {
				continue
			}
			if n := len(r.values); n != 0 && r.values[n-1][0] == m.Token {
				r.values[n-1][1] = r.values[n-1][1].(string) + "," + m.Dictionary
				continue
			}
			r.add(m.Token, m.Dictionary)
		}
		title := fmt.Sprintf(text.TSConfig, t.Schema+"."+t.Name) + "\n" + fmt.Sprintf(text.TSConfigParser, qualify(null(t.ParserSchema), t.Parser))
		if err := encode(f.out, r, title, noFooter, nil); err != nil {
			return err
		}
	}
	return nil
}

var textSearchParsers = list(spec[dbmeta.TextSearchParser]{
	cmd: `\dFp`, title: text.ListOfTSParsers, q: dbmeta.TextSearchParsers,
	fields: []field[dbmeta.TextSearchParser]{
		{name: "Schema", value: func(t dbmeta.TextSearchParser) any { return t.Schema }},
		{name: "Name", value: func(t dbmeta.TextSearchParser) any { return t.Name }},
		description(func(t dbmeta.TextSearchParser) sql.Null[string] { return t.Comment }),
	},
})

var textSearchDictionaries = list(spec[dbmeta.TextSearchDictionary]{
	cmd: `\dFd`, title: text.ListOfTSDictionaries, q: dbmeta.TextSearchDictionaries,
	fields: []field[dbmeta.TextSearchDictionary]{
		{name: "Schema", value: func(t dbmeta.TextSearchDictionary) any { return t.Schema }},
		{name: "Name", value: func(t dbmeta.TextSearchDictionary) any { return t.Name }},
		{name: "Template", verbose: true, value: func(t dbmeta.TextSearchDictionary) any { return t.Template }},
		{name: "Init options", verbose: true, value: func(t dbmeta.TextSearchDictionary) any { return null(t.Options) }},
		description(func(t dbmeta.TextSearchDictionary) sql.Null[string] { return t.Comment }),
	},
})

var textSearchTemplates = list(spec[dbmeta.TextSearchTemplate]{
	cmd: `\dFt`, title: text.ListOfTSTemplates, q: dbmeta.TextSearchTemplates,
	fields: []field[dbmeta.TextSearchTemplate]{
		{name: "Schema", value: func(t dbmeta.TextSearchTemplate) any { return t.Schema }},
		{name: "Name", value: func(t dbmeta.TextSearchTemplate) any { return t.Name }},
		{name: "Init", verbose: true, value: func(t dbmeta.TextSearchTemplate) any { return null(t.Init) }},
		{name: "Lexize", verbose: true, value: func(t dbmeta.TextSearchTemplate) any { return t.Lexize }},
		description(func(t dbmeta.TextSearchTemplate) sql.Null[string] { return t.Comment }),
	},
})

var operatorClasses = spec[dbmeta.OperatorClass]{
	cmd: `\dAc`, title: text.ListOfOperatorClasses, q: dbmeta.OperatorClasses,
	fields: []field[dbmeta.OperatorClass]{
		{name: "AM", value: func(o dbmeta.OperatorClass) any { return o.AccessMethod }},
		{name: "Input type", value: func(o dbmeta.OperatorClass) any { return o.InputType }},
		{name: "Storage type", value: func(o dbmeta.OperatorClass) any { return null(o.StorageType) }},
		{name: "Operator class", value: func(o dbmeta.OperatorClass) any { return o.Name }},
		{name: "Default?", value: func(o dbmeta.OperatorClass) any { return yesNo(o.Default) }},
		{name: "Operator family", verbose: true, value: func(o dbmeta.OperatorClass) any { return o.Family }},
		{name: "Owner", verbose: true, value: func(o dbmeta.OperatorClass) any { return o.Owner }},
	},
}

var operatorFamilies = spec[dbmeta.OperatorFamily]{
	cmd: `\dAf`, title: text.ListOfOperatorFamilies, q: dbmeta.OperatorFamilies,
	fields: []field[dbmeta.OperatorFamily]{
		{name: "AM", value: func(o dbmeta.OperatorFamily) any { return o.AccessMethod }},
		{name: "Operator family", value: func(o dbmeta.OperatorFamily) any { return o.Name }},
		{name: "Applicable types", value: func(o dbmeta.OperatorFamily) any { return null(o.AppliesTo) }},
		{name: "Owner", verbose: true, value: func(o dbmeta.OperatorFamily) any { return o.Owner }},
	},
}

var familyOperators = spec[dbmeta.OperatorFamilyOperator]{
	cmd: `\dAo`, title: text.ListOfFamilyOperators, q: dbmeta.OperatorFamilyOperators,
	fields: []field[dbmeta.OperatorFamilyOperator]{
		{name: "AM", value: func(o dbmeta.OperatorFamilyOperator) any { return o.AccessMethod }},
		{name: "Operator family", value: func(o dbmeta.OperatorFamilyOperator) any { return o.Family }},
		{name: "Operator", value: func(o dbmeta.OperatorFamilyOperator) any { return o.Operator }},
		{name: "Strategy", value: func(o dbmeta.OperatorFamilyOperator) any { return o.Strategy }},
		{name: "Purpose", value: func(o dbmeta.OperatorFamilyOperator) any { return o.Purpose }},
	},
}

var familyFunctions = spec[dbmeta.OperatorFamilyFunction]{
	cmd: `\dAp`, title: text.ListOfFamilyFunctions, q: dbmeta.OperatorFamilyFunctions,
	fields: []field[dbmeta.OperatorFamilyFunction]{
		{name: "AM", value: func(o dbmeta.OperatorFamilyFunction) any { return o.AccessMethod }},
		{name: "Operator family", value: func(o dbmeta.OperatorFamilyFunction) any { return o.Family }},
		{name: "Registered left type", value: func(o dbmeta.OperatorFamilyFunction) any { return o.LeftType }},
		{name: "Registered right type", value: func(o dbmeta.OperatorFamilyFunction) any { return o.RightType }},
		{name: "Number", value: func(o dbmeta.OperatorFamilyFunction) any { return o.Number }},
		{name: "Function", value: func(o dbmeta.OperatorFamilyFunction) any { return o.Function }},
	},
}

// extensionList is the list that \dx prints. \dx+ prints the objects of each
// extension instead.
var extensionList = spec[dbmeta.Extension]{
	cmd: `\dx`, title: text.ListOfExtensions, q: dbmeta.Extensions,
	fields: []field[dbmeta.Extension]{
		{name: "Name", value: func(e dbmeta.Extension) any { return e.Name }},
		{name: "Version", value: func(e dbmeta.Extension) any { return e.Version }},
		{name: "Schema", value: func(e dbmeta.Extension) any { return e.Schema }},
		description(func(e dbmeta.Extension) sql.Null[string] { return e.Comment }),
	},
}

func extensions(ctx context.Context, f *Formatter, c Call) error {
	exts, err := extensionList.read(ctx, f, c, extensionList.named(f, c.pattern(0)))
	switch {
	case err != nil:
		return err
	case !c.Verbose:
		return extensionList.print(f.out, exts, false)
	}
	slices.SortStableFunc(exts, func(a, b dbmeta.Extension) int {
		return cmp.Compare(a.Name, b.Name)
	})
	for _, e := range exts {
		objects, err := all(ctx, f, dbmeta.ExtensionObjects, dbmeta.Args{Name: e.Name})
		if err != nil {
			return f.check(`\dx+`, err)
		}
		r := &rows{columns: []string{"Object description"}}
		for _, o := range objects {
			if o.Extension == e.Name {
				r.add(o.Description)
			}
		}
		slices.SortStableFunc(r.values, compareRows)
		if err := encode(f.out, r, fmt.Sprintf(text.ObjectsInExtension, `"`+e.Name+`"`), noFooter, nil); err != nil {
			return err
		}
	}
	return nil
}

// qualify returns schema.name, or name when there is no schema.
func qualify(schema, name string) string {
	if schema == "" {
		return name
	}
	return schema + "." + name
}

// ShowFunction prints the definition of the function that name names, for
// \sf. name may carry the argument types, as in f(integer), to pick one of
// several functions with the name. verbose numbers the lines of the body.
func (f *Formatter) ShowFunction(ctx context.Context, name string, verbose bool) error {
	pattern, argTypes, hasArgs := strings.Cut(name, "(")
	argTypes = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(argTypes), ")"))
	schema, n := f.meta.Dialect().Pattern(strings.TrimSpace(pattern))
	funcs, err := all(ctx, f, dbmeta.Functions, dbmeta.Args{Schema: schema, Name: n, WithSystem: true})
	if err != nil {
		return f.check(`\sf`, err)
	}
	var match []dbmeta.Function
	if hasArgs {
		want := typeList(strings.Split(argTypes, ","))
		params, err := f.routineTypes(ctx, schema, n)
		if err != nil {
			return err
		}
		for _, f := range funcs {
			// the types alone, or the signature as \df prints it
			if params[null(f.ID)] == want || typeList(strings.Split(null(f.ArgTypes), ",")) == want {
				match = append(match, f)
			}
		}
	} else {
		match = funcs
	}
	switch {
	case len(match) == 0:
		return fmt.Errorf(text.FunctionNotFound, name)
	case len(match) > 1:
		return fmt.Errorf(text.MoreThanOneFunction, name)
	case !match[0].Definition.Valid:
		return fmt.Errorf(text.FunctionHasNoDefinition, name)
	}
	lines := strings.Split(strings.TrimRight(match[0].Definition.V, "\n"), "\n")
	// psql numbers the body, which starts at the line that begins with AS,
	// and indents the header above it
	body := slices.IndexFunc(lines, func(s string) bool {
		_, ok := cutPrefixFold(s, "AS ")
		return ok
	})
	if body == -1 {
		body = 0
	}
	for i, line := range lines {
		switch {
		case verbose && i < body:
			line = "        " + line
		case verbose:
			line = fmt.Sprintf("%-7d %s", i-body+1, line)
		}
		if _, err := fmt.Fprintln(f.out, line); err != nil {
			return err
		}
	}
	return nil
}

// routineTypes returns the types of the arguments of each routine that
// matches schema and name, keyed by the ID of the routine and written as
// typeList writes them.
func (f *Formatter) routineTypes(ctx context.Context, schema, name string) (map[string]string, error) {
	params, err := optional(all(ctx, f, dbmeta.RoutineParameters, dbmeta.Args{Schema: schema, Parent: name, WithSystem: true}))
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(params, func(a, b dbmeta.RoutineParameter) int {
		return cmp.Compare(a.Ordinal, b.Ordinal)
	})
	types := make(map[string][]string)
	for _, p := range params {
		switch strings.ToLower(p.Mode) {
		case "in", "inout", "variadic":
			id := null(p.RoutineID)
			types[id] = append(types[id], p.DataType)
		}
	}
	m := make(map[string]string, len(types))
	for id, v := range types {
		m[id] = typeList(v)
	}
	return m, nil
}

// typeList returns a list of types in one form, so that two can be compared.
func typeList(types []string) string {
	var v []string
	for _, t := range types {
		if t = strings.ToLower(strings.Join(strings.Fields(t), " ")); t != "" {
			v = append(v, t)
		}
	}
	return strings.Join(v, ",")
}

// ShowView prints the definition of each view that matches pattern, for \sv.
// verbose numbers the lines.
func (f *Formatter) ShowView(ctx context.Context, pattern string, verbose bool) error {
	schema, name := f.meta.Dialect().Pattern(pattern)
	views, err := all(ctx, f, dbmeta.Views, dbmeta.Args{Schema: schema, Name: name, WithSystem: true})
	if err != nil {
		return f.check(`\sv`, err)
	}
	if len(views) == 0 {
		_, err := fmt.Fprintf(f.out, text.ViewNotFound+"\n", pattern)
		return err
	}
	for _, v := range views {
		n := v.Name
		if v.Schema != "" {
			n = v.Schema + "." + v.Name
		}
		def := null(v.Definition)
		// SQLite stores the whole statement, and the others the query
		if _, ok := cutPrefixFold(strings.TrimSpace(def), "CREATE"); !ok {
			def = "CREATE OR REPLACE VIEW " + n + " AS\n" + def
		}
		lines := strings.Split(strings.TrimRight(def, "\n"), "\n")
		for i, line := range lines {
			if verbose {
				line = fmt.Sprintf("%-7d %s", i+1, line)
			}
			if _, err := fmt.Fprintln(f.out, line); err != nil {
				return err
			}
		}
	}
	return nil
}
