package sqlite

import (
	"context"
	"errors"
	"fmt"
)

// WithinTransaction führt einen Anwendungsfall atomar aus.
func (d *Database) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if work == nil {
		return errors.New("Transaktionsarbeit fehlt")
	}

	if ctx.Value(transactionKey{}) != nil {
		return errors.New("verschachtelte Transaktion ist nicht unterstützt")
	}

	return d.within(ctx, work)
}

func (d *Database) within(ctx context.Context, work func(context.Context) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Transaktion starten: %w", err)
	}

	defer tx.Rollback()
	if err = work(context.WithValue(ctx, transactionKey{}, tx)); err != nil {
		return err
	}

	return tx.Commit()
}

// executor bindet alle SQLite-Repository-Aufrufe im Transaktionskontext an denselben Tx.
func (d *Database) executor(ctx context.Context) executor {
	if tx, ok := ctx.Value(transactionKey{}).(executor); ok {
		return tx
	}

	return d.db
}
