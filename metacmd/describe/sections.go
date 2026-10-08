package describe

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/xo/dbmeta"
)

// section is a block of lines that psql prints below the columns of a
// relation.
type section struct {
	header string
	lines  []string
}

// indented returns the lines of a section the way psql writes them, with four
// spaces before each line below the header.
func (s section) indented() []string {
	if len(s.lines) == 0 && s.header == "" {
		return nil
	}
	out := []string{s.header}
	for _, l := range s.lines {
		out = append(out, "    "+l)
	}
	return out
}

// listed returns the lines of one "Header: a, b, c" entry, as psql breaks a
// list of names over rows, with each row after the first lined up under the
// first name.
func listed(header string, names []string) []string {
	var lines []string
	pad := strings.Repeat(" ", len(header)+1)
	for i, n := range names {
		sep := ","
		if i == len(names)-1 {
			sep = ""
		}
		if i == 0 {
			lines = append(lines, header+" "+n+sep)
		} else {
			lines = append(lines, pad+n+sep)
		}
	}
	return lines
}

// partitionSections returns the lines about the partitions of a relation: the
// bound and constraint of a partition, and the key and the partitions of a
// partitioned table.
func (f *Formatter) partitionSections(ctx context.Context, t dbmeta.Table, verbose bool) (before, after []string, err error) {
	if !f.supports(dbmeta.Partitions) {
		return nil, nil, nil
	}
	// the parent of a partition
	parents, err := optional(all(ctx, f, dbmeta.Partitions, dbmeta.Args{Name: t.Name, WithSystem: true}))
	if err != nil {
		return nil, nil, err
	}
	for _, p := range parents {
		if p.PartitionSchema != t.Schema || p.Partition != t.Name {
			continue
		}
		before = append(before, "Partition of: "+f.refName(ctx, p.Schema, p.Table, p.TableVisible)+" "+null(p.Bound))
		if c := null(p.Constraint); c != "" {
			before = append(before, "Partition constraint: "+c)
		}
	}
	// the partitions of a partitioned table
	if !strings.EqualFold(t.Type, "partitioned table") {
		return before, nil, nil
	}
	keys, err := optional(all(ctx, f, dbmeta.PartitionedTables, dbmeta.Args{Schema: t.Schema, Name: t.Name, WithSystem: true}))
	if err != nil {
		return nil, nil, err
	}
	for _, k := range keys {
		if k.Schema == t.Schema && k.Name == t.Name && k.Strategy != "" {
			before = append(before, "Partition key: "+k.Expression)
			break
		}
	}
	children, err := optional(all(ctx, f, dbmeta.Partitions, dbmeta.Args{Schema: t.Schema, Parent: t.Name, WithSystem: true}))
	if err != nil {
		return nil, nil, err
	}
	// psql lists the partitions by name, and the default one last
	slices.SortStableFunc(children, func(a, b dbmeta.Partition) int {
		ad, bd := null(a.Bound) == "DEFAULT", null(b.Bound) == "DEFAULT"
		switch {
		case ad != bd && ad:
			return 1
		case ad != bd:
			return -1
		}
		return strings.Compare(a.Partition, b.Partition)
	})
	var names []string
	for _, p := range children {
		if p.Schema != t.Schema || p.Table != t.Name {
			continue
		}
		s := f.refName(ctx, p.PartitionSchema, p.Partition, p.PartitionVisible) + " " + null(p.Bound)
		if p.Partitioned {
			s += ", PARTITIONED"
		}
		names = append(names, s)
	}
	switch {
	case len(names) == 0:
	case verbose:
		after = append(after, listed("Partitions:", names)...)
	default:
		after = append(after, fmt.Sprintf("Number of partitions: %d (Use \\d+ to list them.)", len(names)))
	}
	return before, after, nil
}

// inheritanceSections returns the Inherits and Child tables lines of a table.
func (f *Formatter) inheritanceSections(ctx context.Context, t dbmeta.Table, verbose bool) ([]string, error) {
	if !f.supports(dbmeta.Inherits) {
		return nil, nil
	}
	var lines []string
	parents, err := optional(all(ctx, f, dbmeta.Inherits, dbmeta.Args{Schema: t.Schema, Name: t.Name, WithSystem: true}))
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(parents, func(a, b dbmeta.Inherit) int { return int(a.Ordinal - b.Ordinal) })
	var names []string
	for _, p := range parents {
		if p.Schema == t.Schema && p.Name == t.Name && !p.Partition {
			names = append(names, f.refName(ctx, p.ParentSchema, p.Parent, p.ParentVisible))
		}
	}
	lines = append(lines, listed("Inherits:", names)...)
	children, err := optional(all(ctx, f, dbmeta.Inherits, dbmeta.Args{Parent: t.Name, WithSystem: true}))
	if err != nil {
		return nil, err
	}
	names = names[:0]
	for _, c := range children {
		if c.ParentSchema == t.Schema && c.Parent == t.Name && !c.Partition {
			names = append(names, f.refName(ctx, c.Schema, c.Name, c.Visible))
		}
	}
	slices.Sort(names)
	switch {
	case len(names) == 0:
	case verbose:
		lines = append(lines, listed("Child tables:", names)...)
	default:
		lines = append(lines, fmt.Sprintf("Number of child tables: %d (Use \\d+ to list them.)", len(names)))
	}
	return lines, nil
}

// policySection returns the Policies section of a table.
func (f *Formatter) policySection(ctx context.Context, t dbmeta.Table) (section, error) {
	if !f.supports(dbmeta.Policies) {
		return section{}, nil
	}
	policies, err := optional(all(ctx, f, dbmeta.Policies, dbmeta.Args{Schema: t.Schema, Parent: t.Name, WithSystem: true}))
	if err != nil {
		return section{}, err
	}
	policies = slices.DeleteFunc(policies, func(p dbmeta.Policy) bool {
		return p.Schema != t.Schema || p.Table != t.Name
	})
	on := t.RowSecurity.Valid && t.RowSecurity.V
	forced := t.RowSecurityForced.Valid && t.RowSecurityForced.V
	var header string
	switch {
	case len(policies) == 0 && !on:
		return section{}, nil
	case !on:
		header = "Policies (row security disabled):"
	case forced && len(policies) == 0:
		header = "Policies (forced row security enabled): (none)"
	case forced:
		header = "Policies (forced row security enabled):"
	case len(policies) == 0:
		header = "Policies (row security enabled): (none)"
	default:
		header = "Policies:"
	}
	s := section{header: header}
	for _, p := range policies {
		line := `POLICY "` + p.Name + `"`
		if !p.Permissive {
			line += " AS RESTRICTIVE"
		}
		if c := strings.ToUpper(p.Command); c != "" && c != "ALL" {
			line += " FOR " + c
		}
		s.lines = append(s.lines, line)
		if r := null(p.Roles); r != "" {
			s.lines = append(s.lines, "  TO "+r)
		}
		if u := null(p.Using); u != "" {
			s.lines = append(s.lines, "  USING ("+u+")")
		}
		if w := null(p.WithCheck); w != "" {
			s.lines = append(s.lines, "  WITH CHECK ("+w+")")
		}
	}
	return s, nil
}

// ruleSections returns the sections of the rules of a table, one for each
// way a rule is enabled.
func (f *Formatter) ruleSections(ctx context.Context, t dbmeta.Table) ([]section, error) {
	if !f.supports(dbmeta.Rules) {
		return nil, nil
	}
	rules, err := optional(all(ctx, f, dbmeta.Rules, dbmeta.Args{Schema: t.Schema, Parent: t.Name, WithSystem: true}))
	if err != nil {
		return nil, err
	}
	headers := []struct{ enabled, header string }{
		{"enabled", "Rules:"}, {"disabled", "Disabled rules:"}, {"always", "Rules firing always:"}, {"replica", "Rules firing on replica only:"},
	}
	var res []section
	for _, h := range headers {
		s := section{header: h.header}
		for _, r := range rules {
			if r.Schema != t.Schema || r.Table != t.Name || !strings.EqualFold(r.Enabled, h.enabled) {
				continue
			}
			def := strings.TrimSuffix(strings.TrimSpace(r.Definition), ";")
			def = strings.TrimPrefix(def, "CREATE RULE ")
			s.lines = append(s.lines, def)
		}
		if len(s.lines) != 0 {
			res = append(res, s)
		}
	}
	return res, nil
}

// notNullSection returns the Not-null constraints section, which PostgreSQL
// 18 gives.
func (f *Formatter) notNullSection(ctx context.Context, t dbmeta.Table) (section, error) {
	if !f.supports(dbmeta.NotNulls) {
		return section{}, nil
	}
	nns, err := optional(all(ctx, f, dbmeta.NotNulls, dbmeta.Args{Schema: t.Schema, Parent: t.Name, WithSystem: true}))
	if err != nil {
		return section{}, err
	}
	s := section{header: "Not-null constraints:"}
	for _, n := range nns {
		if n.Schema != t.Schema || n.Table != t.Name {
			continue
		}
		line := `"` + n.Name + `" NOT NULL "` + n.Column + `"`
		if n.NoInherit {
			line += " NO INHERIT"
		}
		if !n.Validated {
			line += " NOT VALID"
		}
		switch {
		case n.Inherited && !n.Local:
			line += " (inherited)"
		}
		s.lines = append(s.lines, line)
	}
	if len(s.lines) == 0 {
		return section{}, nil
	}
	return s, nil
}

// publicationSection returns the Publications section of a table.
func (f *Formatter) publicationSection(ctx context.Context, t dbmeta.Table) (section, error) {
	if !f.supports(dbmeta.PublicationTables) {
		return section{}, nil
	}
	pts, err := optional(all(ctx, f, dbmeta.PublicationTables, dbmeta.Args{Schema: t.Schema, Parent: t.Name, WithSystem: true}))
	if err != nil {
		return section{}, err
	}
	s := section{header: "Publications:"}
	for _, p := range pts {
		if p.Schema != t.Schema || p.Name != t.Name {
			continue
		}
		line := `"` + p.Publication + `"`
		if p.Columns != "" {
			line += " (" + p.Columns + ")"
		}
		if w := null(p.Where); w != "" {
			line += " WHERE (" + w + ")"
		}
		s.lines = append(s.lines, line)
	}
	if len(s.lines) == 0 {
		return section{}, nil
	}
	return s, nil
}

// statisticsSection returns the Statistics objects section of a table.
func (f *Formatter) statisticsSection(ctx context.Context, t dbmeta.Table) (section, error) {
	if !f.supports(dbmeta.ExtendedStats) {
		return section{}, nil
	}
	stats, err := optional(all(ctx, f, dbmeta.ExtendedStats, dbmeta.Args{Schema: t.Schema, Parent: t.Name, WithSystem: true}))
	if err != nil {
		return section{}, err
	}
	s := section{header: "Statistics objects:"}
	for _, st := range stats {
		if st.Schema != t.Schema || st.Table != t.Name {
			continue
		}
		name := st.Name.V
		if !st.Name.Valid {
			name = ""
		}
		var kinds []string
		for _, k := range []struct {
			on   bool
			name string
		}{{st.Ndistinct, "ndistinct"}, {st.Dependencies, "dependencies"}, {st.MCV, "mcv"}} {
			if k.on {
				kinds = append(kinds, k.name)
			}
		}
		s.lines = append(s.lines, fmt.Sprintf(`"%s.%s" (%s) ON %s`, st.Schema, name, strings.Join(kinds, ", "), null(st.Definition)))
	}
	if len(s.lines) == 0 {
		return section{}, nil
	}
	return s, nil
}

// viewDefinition returns the View definition section of a view or a
// materialized view.
func (f *Formatter) viewDefinition(ctx context.Context, t dbmeta.Table) ([]string, error) {
	class := classOf(t.Type)
	if class != classView && class != classMaterialized || !f.supports(dbmeta.Views) {
		return nil, nil
	}
	views, err := optional(all(ctx, f, dbmeta.Views, dbmeta.Args{Schema: t.Schema, Name: t.Name, WithSystem: true}))
	if err != nil {
		return nil, err
	}
	for _, v := range views {
		if v.Schema == t.Schema && v.Name == t.Name && v.Definition.Valid {
			return []string{"View definition:", strings.TrimRight(v.Definition.V, "\n")}, nil
		}
	}
	return nil, nil
}

// foreignTableLines returns the Server and FDW options lines of a foreign
// table.
func (f *Formatter) foreignTableLines(ctx context.Context, t dbmeta.Table) ([]string, error) {
	if !f.supports(dbmeta.ForeignTables) {
		return nil, nil
	}
	tables, err := optional(all(ctx, f, dbmeta.ForeignTables, dbmeta.Args{Schema: t.Schema, Name: t.Name, WithSystem: true}))
	if err != nil {
		return nil, err
	}
	for _, ft := range tables {
		if ft.Schema != t.Schema || ft.Name != t.Name {
			continue
		}
		lines := []string{"Server: " + ft.Server}
		opts, err := f.foreignOptions(ctx, "foreign table")
		if err != nil {
			return nil, err
		}
		o := opts[t.Schema+"\x00"+t.Name]
		if o == "" && opts == nil {
			o = null(ft.Options)
		}
		if o != "" {
			lines = append(lines, "FDW options: "+o)
		}
		return lines, nil
	}
	return nil, nil
}

// columnPrivileges returns the column privileges of each relation, written
// as psql writes them: the column, a colon, and an access item on each line
// below. The key is the schema and the name of the relation. It returns nil
// when the database cannot say.
func (f *Formatter) columnPrivileges(ctx context.Context) (map[string]string, error) {
	if !f.supports(dbmeta.ColumnPrivileges) {
		return nil, nil
	}
	privs, err := optional(all(ctx, f, dbmeta.ColumnPrivileges, dbmeta.Args{WithSystem: true}))
	if err != nil || privs == nil {
		return nil, err
	}
	slices.SortStableFunc(privs, func(a, b dbmeta.ColumnPrivilege) int {
		return cmp.Or(cmp.Compare(a.Schema, b.Schema), cmp.Compare(a.Table, b.Table), cmp.Compare(a.Ordinal, b.Ordinal))
	})
	lines := make(map[string][]string)
	last := make(map[string]string)
	for _, p := range privs {
		key := p.Schema + "\x00" + p.Table
		if last[key] != p.Column {
			lines[key] = append(lines[key], p.Column+":")
			last[key] = p.Column
		}
		lines[key] = append(lines[key], "  "+p.Access)
	}
	m := make(map[string]string, len(lines))
	for k, v := range lines {
		m[k] = strings.Join(v, "\n")
	}
	return m, nil
}

// policySummaries returns the policies of each relation, written as psql
// writes them in \dp: the name with its command, then the expressions and the
// roles. The key is the schema and the name of the relation. It returns nil
// when the database cannot say.
func (f *Formatter) policySummaries(ctx context.Context) (map[string]string, error) {
	if !f.supports(dbmeta.Policies) {
		return nil, nil
	}
	policies, err := optional(all(ctx, f, dbmeta.Policies, dbmeta.Args{WithSystem: true}))
	if err != nil || policies == nil {
		return nil, err
	}
	slices.SortStableFunc(policies, func(a, b dbmeta.Policy) int {
		return cmp.Or(cmp.Compare(a.Schema, b.Schema), cmp.Compare(a.Table, b.Table), cmp.Compare(a.Name, b.Name))
	})
	letters := map[string]string{"select": "r", "insert": "a", "update": "w", "delete": "d"}
	lines := make(map[string][]string)
	for _, p := range policies {
		key := p.Schema + "\x00" + p.Table
		head := p.Name
		if !p.Permissive {
			head += " (RESTRICTIVE)"
		}
		if l, ok := letters[strings.ToLower(p.Command)]; ok {
			head += " (" + l + "):"
		} else {
			head += ":"
		}
		lines[key] = append(lines[key], head)
		if u := null(p.Using); u != "" {
			lines[key] = append(lines[key], "  (u): "+u)
		}
		if w := null(p.WithCheck); w != "" {
			lines[key] = append(lines[key], "  (c): "+w)
		}
		if r := null(p.Roles); r != "" {
			lines[key] = append(lines[key], "  to: "+r)
		}
	}
	m := make(map[string]string, len(lines))
	for k, v := range lines {
		m[k] = strings.Join(v, "\n")
	}
	return m, nil
}
