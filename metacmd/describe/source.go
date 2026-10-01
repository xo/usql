package describe

import (
	"context"
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

// Objects satisfies complete.Source.
func (s source) Objects(kind complete.Kind, f complete.Filter) ([]complete.Object, error) {
	// a completion must not hang on a slow database
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	a := args{Catalog: f.Catalog, Schema: f.Schema, Name: f.Name, WithSystem: f.WithSystem}
	var objs []complete.Object
	switch kind {
	case complete.Catalogs:
		dbs, err := all(ctx, s.f, dbmeta.Databases, args{Name: f.Name})
		for _, d := range dbs {
			objs = append(objs, complete.Object{Name: d.Name})
		}
		return objs, err
	case complete.Schemas:
		// a qualified name completes the schemas of a catalog
		a := args{Name: f.Name, WithSystem: true}
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
		columns, err := all(ctx, s.f, dbmeta.Columns, args{Catalog: f.Catalog, Schema: f.Schema, Parent: f.Parent, WithSystem: f.WithSystem})
		for _, c := range columns {
			objs = append(objs, complete.Object{Catalog: c.Catalog, Schema: c.Schema, Name: c.Name})
		}
		return objs, err
	}
	return nil, ErrNotSupported
}
