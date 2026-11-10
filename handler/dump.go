package handler

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/xo/dbmeta"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/dump"
	"github.com/xo/usql/text"
)

// Dump writes the database of the connection to w as a usql script, which
// usql runs with -f to make the database again. It reads the schema and the
// rows in one transaction, so that they agree. In an open transaction it reads
// in that one. Each warning of the dump goes to standard error before the
// script is written.
//
// Outside an open transaction, the script goes to a temporary file first, and
// w gets it only when it is complete. When reading the rows fails, the dump
// starts again in a new transaction after a pause, as [dump.Spool] does,
// because Oracle refuses a read-only transaction that reads a table made a
// few seconds before it started. In an open transaction it does not start
// again, because only that transaction sees its rows.
func (h *Handler) Dump(ctx context.Context, w io.Writer, opts dump.Options) error {
	if h.db == nil {
		return text.ErrNotConnected
	}
	m, err := h.metadata(ctx)
	if err != nil {
		return err
	}
	opts.Version = text.CommandVersion
	if h.tx != nil {
		return h.dumpIn(ctx, m, h.tx, w, opts, true)
	}
	db, ok := h.DB().(interface {
		BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	})
	if !ok {
		return text.ErrNotConnected
	}
	warn := true
	return dump.Spool(ctx, w, dump.Waits, func(w io.Writer) error {
		tx, err := dump.Begin(ctx, db)
		if err != nil {
			return drivers.WrapErr(h.u.SchemeName, err)
		}
		// the transaction only reads
		defer func() { _ = tx.Rollback() }()
		err = h.dumpIn(ctx, m, tx, w, opts, warn)
		warn = false
		return err
	}, func(err error, wait time.Duration) {
		fmt.Fprintf(h.l.Stderr(), "warning: the dump failed, and starts again in %v: %v\n", wait, err)
	})
}

// dumpIn writes the dump that q reads to w. warn says whether the warnings
// of the dump go to standard error, so that a new try does not write them
// again. The dump marks an error of the database while it reads the rows as
// one that a new try can avoid.
func (h *Handler) dumpIn(ctx context.Context, m *dbmeta.Meta, q dbmeta.Queryer, w io.Writer, opts dump.Options, warn bool) error {
	d, err := dump.New(ctx, m, q, opts)
	if err != nil {
		return drivers.WrapErr(h.u.SchemeName, err)
	}
	if warn {
		for _, line := range d.Warnings() {
			fmt.Fprintln(h.l.Stderr(), "warning:", line)
		}
	}
	if err := d.Write(ctx, w); err != nil {
		return drivers.WrapErr(h.u.SchemeName, err)
	}
	return nil
}
