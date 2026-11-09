package handler

import (
	"context"
	"database/sql"
	"fmt"
	"io"

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
func (h *Handler) Dump(ctx context.Context, w io.Writer, opts dump.Options) error {
	if h.db == nil {
		return text.ErrNotConnected
	}
	m, err := h.metadata(ctx)
	if err != nil {
		return err
	}
	var q dbmeta.Queryer = h.tx
	if h.tx == nil {
		db, ok := h.DB().(interface {
			BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
		})
		if !ok {
			return text.ErrNotConnected
		}
		tx, err := dump.Begin(ctx, db)
		if err != nil {
			return drivers.WrapErr(h.u.SchemeName, err)
		}
		// the transaction only reads
		defer func() { _ = tx.Rollback() }()
		q = tx
	}
	opts.Version = text.CommandVersion
	d, err := dump.New(ctx, m, q, opts)
	if err != nil {
		return drivers.WrapErr(h.u.SchemeName, err)
	}
	for _, line := range d.Warnings() {
		fmt.Fprintln(h.l.Stderr(), "warning:", line)
	}
	if err := d.Write(ctx, w); err != nil {
		return drivers.WrapErr(h.u.SchemeName, err)
	}
	return nil
}
