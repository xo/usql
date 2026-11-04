package describe

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/xo/dbmeta"
	"github.com/xo/usql/metacmd/complete"
)

// Source returns the source of the names that the completer offers, read
// from the database db through m.
func Source(m *dbmeta.Meta, db dbmeta.Queryer) complete.Source {
	return source{f: NewFormatter(m, db, nil, "")}
}

// source answers the objects of a database for the complete.
type source struct {
	f *Formatter
}

// Objects satisfies complete.Source. The text typed so far is folded as the
// database folds a name, as \d folds a pattern, so that ord and ORD both find
// orders on PostgreSQL. A name that the database would fold differently, or
// that has a character a plain name cannot hold, comes back quoted.
func (s source) Objects(kind complete.Kind, f complete.Filter) ([]complete.Object, error) {
	f.Catalog, f.Schema, f.Parent, f.Name = s.fold(f.Catalog), s.fold(f.Schema), s.fold(f.Parent), s.fold(f.Name)
	objs, err := s.objects(kind, f)
	for i := range objs {
		objs[i].Catalog, objs[i].Schema, objs[i].Name = s.quote(objs[i].Catalog), s.quote(objs[i].Schema), s.quote(objs[i].Name)
	}
	return objs, err
}

// fold folds a part of a name as the database folds a name that is not
// quoted. A part that opens a quote and does not close it yet, as while it is
// typed, keeps its case.
func (s source) fold(part string) string {
	if part == "" {
		return part
	}
	if q := part[0]; q == '"' || q == '`' {
		if !strings.ContainsRune(strings.TrimSuffix(part[1:], "%"), rune(q)) {
			return part[1:]
		}
	}
	return s.f.meta.Dialect().FoldIdentifier(part)
}

// plainName is a name that needs no quotes, unless the database folds it.
var plainName = regexp.MustCompile(`^[\pL_][\pL\pN_$]*$`)

// quote returns name between the quotes of the database when it would not
// stay the same name without them, such as MyTable on PostgreSQL, which
// folds it to mytable, or Order Details on SQL Server.
func (s source) quote(name string) string {
	d := s.f.meta.Dialect()
	if name == "" || plainName.MatchString(name) && d.FoldIdentifier(name) == name {
		return name
	}
	q := `"`
	if syn, ok := d.Syntax(); ok && syn.Backticks {
		q = "`"
	}
	return q + strings.ReplaceAll(name, q, q+q) + q
}

// objects returns the objects of kind that f selects.
func (s source) objects(kind complete.Kind, f complete.Filter) ([]complete.Object, error) {
	// a completion must not hang on a slow database
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	a := dbmeta.Args{Catalog: f.Catalog, Schema: f.Schema, Name: f.Name, WithSystem: f.WithSystem}
	var objs []complete.Object
	switch kind {
	case complete.Catalogs:
		dbs, err := all(ctx, s.f, dbmeta.Databases, dbmeta.Args{Name: f.Name})
		for _, d := range dbs {
			objs = append(objs, complete.Object{Name: d.Name})
		}
		return objs, err
	case complete.Schemas:
		// a qualified name completes the schemas of a catalog
		a := dbmeta.Args{Name: f.Name, WithSystem: true}
		if f.Schema != "" {
			a.Catalog = f.Schema
		}
		schemas, err := all(ctx, s.f, dbmeta.Schemas, a)
		for _, sc := range schemas {
			objs = append(objs, complete.Object{Catalog: sc.Catalog, Name: sc.Name})
		}
		return objs, err
	case complete.Tables:
		want := classes(f.Types)
		tables, err := all(ctx, s.f, dbmeta.Tables, a)
		for _, t := range tables {
			if want[classOf(t.Type)] {
				objs = append(objs, complete.Object{Catalog: t.Catalog, Schema: t.Schema, Name: t.Name})
			}
		}
		return objs, err
	case complete.Functions:
		q := dbmeta.Functions
		if f.Aggregates && s.f.supports(dbmeta.Aggregates) {
			q = dbmeta.Aggregates
		}
		funcs, err := all(ctx, s.f, q, a)
		for _, fn := range funcs {
			if f.Aggregates && !strings.HasPrefix(strings.ToLower(fn.Kind), "agg") {
				continue
			}
			objs = append(objs, complete.Object{Catalog: fn.Catalog, Schema: fn.Schema, Name: fn.Name})
		}
		return objs, err
	case complete.Sequences:
		seqs, err := all(ctx, s.f, dbmeta.Sequences, a)
		for _, sq := range seqs {
			objs = append(objs, complete.Object{Schema: sq.Schema, Name: sq.Name})
		}
		return objs, err
	case complete.Indexes:
		indexes, err := all(ctx, s.f, dbmeta.Indexes, a)
		for _, i := range indexes {
			objs = append(objs, complete.Object{Catalog: i.Catalog, Schema: i.Schema, Name: i.Name})
		}
		return objs, err
	case complete.Columns:
		columns, err := all(ctx, s.f, dbmeta.Columns, dbmeta.Args{Catalog: f.Catalog, Schema: f.Schema, Parent: f.Parent, WithSystem: f.WithSystem})
		for _, c := range columns {
			objs = append(objs, complete.Object{Catalog: c.Catalog, Schema: c.Schema, Name: c.Name})
		}
		return objs, err
	}
	return nil, ErrNotSupported
}
