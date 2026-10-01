package describe

import (
	"context"
	"strings"
	"time"

	"github.com/xo/dbmeta"
	"github.com/xo/usql/drivers/completer"
)

// Source returns the source of the names that the completer offers, read
// from the database db through m.
func Source(m *dbmeta.Meta, db dbmeta.Queryer) completer.Source {
	return source{f: NewFormatter(m, db, nil, "")}
}

// source answers the objects of a database for the completer.
type source struct {
	f *Formatter
}

// Objects satisfies completer.Source.
func (s source) Objects(kind completer.Kind, f completer.Filter) ([]completer.Object, error) {
	// a completion must not hang on a slow database
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	a := args{Catalog: f.Catalog, Schema: f.Schema, Name: f.Name, WithSystem: f.WithSystem}
	var objs []completer.Object
	switch kind {
	case completer.Catalogs:
		dbs, err := all(ctx, s.f, dbmeta.Databases, args{Name: f.Name})
		for _, d := range dbs {
			objs = append(objs, completer.Object{Name: d.Name})
		}
		return objs, err
	case completer.Schemas:
		// a qualified name completes the schemas of a catalog
		a := args{Name: f.Name, WithSystem: true}
		if f.Schema != "" {
			a.Catalog = f.Schema
		}
		schemas, err := all(ctx, s.f, dbmeta.Schemas, a)
		for _, sc := range schemas {
			objs = append(objs, completer.Object{Catalog: sc.Catalog, Name: sc.Name})
		}
		return objs, err
	case completer.Tables:
		want := classes(f.Types)
		tables, err := all(ctx, s.f, dbmeta.Tables, a)
		for _, t := range tables {
			if want[classOf(t.Type)] {
				objs = append(objs, completer.Object{Catalog: t.Catalog, Schema: t.Schema, Name: t.Name})
			}
		}
		return objs, err
	case completer.Functions:
		q := dbmeta.Functions
		if f.Aggregates && s.f.supports(dbmeta.Aggregates) {
			q = dbmeta.Aggregates
		}
		funcs, err := all(ctx, s.f, q, a)
		for _, fn := range funcs {
			if f.Aggregates && !strings.HasPrefix(strings.ToLower(fn.Kind), "agg") {
				continue
			}
			objs = append(objs, completer.Object{Catalog: fn.Catalog, Schema: fn.Schema, Name: fn.Name})
		}
		return objs, err
	case completer.Sequences:
		seqs, err := all(ctx, s.f, dbmeta.Sequences, a)
		for _, sq := range seqs {
			objs = append(objs, completer.Object{Schema: sq.Schema, Name: sq.Name})
		}
		return objs, err
	case completer.Indexes:
		indexes, err := all(ctx, s.f, dbmeta.Indexes, a)
		for _, i := range indexes {
			objs = append(objs, completer.Object{Catalog: i.Catalog, Schema: i.Schema, Name: i.Name})
		}
		return objs, err
	case completer.Columns:
		columns, err := all(ctx, s.f, dbmeta.Columns, args{Catalog: f.Catalog, Schema: f.Schema, Parent: f.Parent, WithSystem: f.WithSystem})
		for _, c := range columns {
			objs = append(objs, completer.Object{Catalog: c.Catalog, Schema: c.Schema, Name: c.Name})
		}
		return objs, err
	}
	return nil, ErrNotSupported
}
