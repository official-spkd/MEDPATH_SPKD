package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Request memberi setiap request snapshot terpisah. Semua perubahan dan audit
// commit bersama. Lock PostgreSQL menyerialkan penulis lintas instance, termasuk
// penentuan ID dan validasi transisi; pembaca memakai snapshot repeatable-read.
func (s *Store) Request(ctx context.Context, write bool, run func(*Store) error) error {
	if s.pool == nil {
		return run(s)
	}
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	opts := pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}
	if write {
		opts = pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	}
	tx, err := s.pool.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if write {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(71423001)"); err != nil {
			return err
		}
	}
	current := newBase()
	current.db, current.ctx = tx, ctx
	if err := current.loadDB(); err != nil {
		return err
	}
	runErr := run(current)
	if current.dbErr != nil {
		return current.dbErr
	}
	if runErr != nil {
		return runErr
	}
	return tx.Commit(ctx)
}
