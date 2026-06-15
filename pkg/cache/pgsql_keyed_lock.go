// Postgres-backed implementation of a distributed keyed lock —
// companion to the existing cache.AcquireMultiLock (Redis) and a
// general facility for serializing operations across pods when:
//
//   - the key is operation-specific (composite of identifiers the
//     caller already has), and
//   - the lock must survive a Redis outage / be backed by the same
//     durable store that holds the saga state (Postgres).
//
// The TABT (`treasury_asset_balance_transfer`) saga is the first
// consumer — its `tabt_acquire_vault_lock` and `tabt_release_vault_lock`
// steps compose a key from
// `(account_iid, treasury_deploy_id, exec_runtime, vault_address,
//
//	source_stash, destination_stash)` and call into this store. Future
//
// callers can compose keys however they like; the store knows nothing
// about the inputs.
//
// # Liveness — what if Release never runs?
//
// Acquire can be lost in three ways: (1) the holding pod crashes
// between Acquire and Release; (2) Release itself fails (bad DB
// connection during the DELETE); (3) the saga errors out before
// reaching the release step. Without a self-healing path the row
// would block legitimate callers forever.
//
// Mitigation: every row carries an `expires_at` column. Acquire uses
// `INSERT … ON CONFLICT (lock_key) DO UPDATE …` gated by
// `WHERE expires_at <= NOW()`, so a new holder can atomically steal
// an expired lock. If the existing row is still alive the DO UPDATE
// is filtered out (RowsAffected = 0) and Acquire returns
// [ErrKeyedLockHeld].
//
// # TTL — who applies it
//
// The CALLER picks the TTL on each Acquire. There is no daemon-side
// "default for all locks" — different operations have different
// runtimes and different acceptable starvation windows. Recommended
// callsites:
//
//   - The saga step that acquires the lock passes a TTL that bounds
//     the WHOLE remaining saga runtime + a safety margin. For TABT
//     today that's [DefaultKeyedLockTTL] (5 minutes) — the saga
//     does at most one TVB transfer (typically <30s) followed by 3
//     short query/verify steps.
//   - Long-running steps that risk crossing the original TTL
//     boundary call [PgsqlKeyedLockStore.Refresh] inside their body
//     to push `expires_at` forward. TABT doesn't need this today;
//     the pattern is documented for future callers.
//   - The release step calls [PgsqlKeyedLockStore.Release] on BOTH
//     the commit path AND the compensation path. Idempotent — zero-
//     row delete is fine.
//   - A background reaper (optional) calls
//     [PgsqlKeyedLockStore.ReapExpired] periodically to keep the
//     table small. The steal-on-acquire path already reclaims
//     expired rows for correctness; this is hygiene only. Wire
//     from any long-lived daemon (treassvc, accmgr, traxctrl — any
//     pod that already has a *sql.DB) at startup with a 1-minute
//     ticker.
package cache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// DefaultKeyedLockTTL is the recommended TTL when the caller has no
// reason to pick something else. 5 minutes covers most saga
// runtimes; callers with longer operations should pass a larger
// duration explicitly.
const DefaultKeyedLockTTL = 5 * time.Minute

// ErrKeyedLockHeld is returned by [KeyedLockStore.Acquire] when
// another holder already holds a LIVE lock for the same key (i.e.
// `expires_at > NOW()`). Distinct from a real DB error — callers
// fail-fast and surface a "resource locked" message to the operator.
var ErrKeyedLockHeld = errors.New("cache.KeyedLockStore: another holder owns the lock for this key")

// KeyedLockStore is a single-key distributed lock backed by a row
// with `expires_at` semantics. Compose multi-resource locks by
// acquiring several keys in a deterministic order (sorted) — same
// pattern [Cache.AcquireMultiLock] uses on Redis.
type KeyedLockStore interface {
	// Acquire takes the lock for `key` on behalf of `holderID`. The
	// lock auto-expires after `ttl` — past which a future Acquire
	// can atomically steal it. Pass [DefaultKeyedLockTTL] when
	// unsure.
	//
	// Returns [ErrKeyedLockHeld] when an existing lock for the same
	// key is still live; any other error is a real backend failure.
	Acquire(ctx context.Context, key, holderID string, ttl time.Duration) error

	// Release deletes EVERY row tagged with `holderID`. Idempotent —
	// safe to call even if Acquire didn't run (zero rows deleted).
	// Use a single holderID per logical operation (e.g. a saga
	// instance id) so this single call cleans up every lock the
	// operation held.
	Release(ctx context.Context, holderID string) (int, error)

	// Refresh extends `expires_at` on every row tagged with
	// `holderID` to `NOW() + ttl`. Use when a long-running operation
	// risks crossing the original TTL boundary while still working.
	Refresh(ctx context.Context, holderID string, ttl time.Duration) (int, error)

	// IsHeldBy reports whether `holderID` currently holds any LIVE
	// lock. Diagnostic only — callers MUST NOT rely on it for
	// correctness (Acquire's atomic INSERT-ON-CONFLICT is the
	// source of truth).
	IsHeldBy(ctx context.Context, holderID string) (bool, error)

	// ReapExpired DELETEs every row whose lock has expired. Returns
	// the number of rows deleted. Safe to run concurrently with
	// Acquire / Release. Optional — the steal-on-acquire path
	// already reclaims expired rows for correctness.
	ReapExpired(ctx context.Context) (int, error)
}

// PgsqlKeyedLockStore is the durable, multi-pod-safe implementation
// backed by `shared.distributed_locks`. Schema DDL is owned by this
// file's [PgsqlKeyedLockStore.EnsureSchema] (called at daemon
// startup); init_shared_pgsql.sql owns only the `shared` schema
// namespace.
type PgsqlKeyedLockStore struct {
	db *sql.DB
}

// NewPgsqlKeyedLockStore wraps a *sql.DB. Caller owns the connection
// lifetime.
func NewPgsqlKeyedLockStore(db *sql.DB) *PgsqlKeyedLockStore {
	return &PgsqlKeyedLockStore{db: db}
}

// EnsureSchema is the fail-safe CREATE TABLE IF NOT EXISTS, called
// once at daemon startup. Idempotent.
func (s *PgsqlKeyedLockStore) EnsureSchema(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS shared.distributed_locks (
			lock_key    TEXT        NOT NULL PRIMARY KEY,
			holder_id   TEXT        NOT NULL,
			acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at  TIMESTAMPTZ NOT NULL
		);
		CREATE INDEX IF NOT EXISTS distributed_locks_holder_id_idx
		    ON shared.distributed_locks (holder_id);
		CREATE INDEX IF NOT EXISTS distributed_locks_expires_at_idx
		    ON shared.distributed_locks (expires_at);
	`)
	if err != nil {
		return fmt.Errorf("cache.PgsqlKeyedLockStore.EnsureSchema: %w", err)
	}
	return nil
}

// sqlKeyedLockAcquire — INSERT with ON CONFLICT DO UPDATE gated on the
// existing row being expired. Outcomes:
//   - No existing row → INSERT, RowsAffected = 1.
//   - Existing row, expired (`expires_at <= NOW()`) → DO UPDATE
//     overwrites with the new holder + a fresh TTL. RowsAffected = 1.
//   - Existing row, still alive → ON CONFLICT triggers DO UPDATE but
//     the WHERE-clause filters it out. RowsAffected = 0 → caller
//     translates to ErrKeyedLockHeld.
const sqlKeyedLockAcquire = `
INSERT INTO shared.distributed_locks (lock_key, holder_id, acquired_at, expires_at)
VALUES ($1, $2, NOW(), NOW() + ($3::TEXT || ' milliseconds')::INTERVAL)
ON CONFLICT (lock_key)
DO UPDATE SET
    holder_id   = EXCLUDED.holder_id,
    acquired_at = EXCLUDED.acquired_at,
    expires_at  = EXCLUDED.expires_at
 WHERE shared.distributed_locks.expires_at <= NOW()
`

// Acquire INSERTs a fresh lock row, or steals an expired one via
// ON CONFLICT DO UPDATE. RowsAffected = 0 means a live lock blocked
// us → [ErrKeyedLockHeld].
func (s *PgsqlKeyedLockStore) Acquire(ctx context.Context, key, holderID string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = DefaultKeyedLockTTL
	}
	res, err := s.db.ExecContext(ctx, sqlKeyedLockAcquire,
		key, holderID,
		fmt.Sprintf("%d", ttl.Milliseconds()),
	)
	if err != nil {
		// pq.Error code "23505" only fires if a uniqueness
		// constraint OTHER than the primary key trips — none today,
		// kept for defence-in-depth.
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return fmt.Errorf("%w (key=%q)", ErrKeyedLockHeld, key)
		}
		return fmt.Errorf("cache.PgsqlKeyedLockStore.Acquire(%q): %w", key, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("cache.PgsqlKeyedLockStore.Acquire(%q): rows-affected: %w", key, err)
	}
	if n == 0 {
		// Existing live lock blocked both INSERT and DO UPDATE.
		return fmt.Errorf("%w (key=%q)", ErrKeyedLockHeld, key)
	}
	return nil
}

const sqlKeyedLockRelease = `
DELETE FROM shared.distributed_locks
 WHERE holder_id = $1
`

// Release deletes every row tagged with `holderID`. Returns the number
// of rows deleted (0 is fine — Release is idempotent).
func (s *PgsqlKeyedLockStore) Release(ctx context.Context, holderID string) (int, error) {
	res, err := s.db.ExecContext(ctx, sqlKeyedLockRelease, holderID)
	if err != nil {
		return 0, fmt.Errorf("cache.PgsqlKeyedLockStore.Release(%q): %w", holderID, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("cache.PgsqlKeyedLockStore.Release(%q): rows-affected: %w", holderID, err)
	}
	return int(n), nil
}

const sqlKeyedLockRefresh = `
UPDATE shared.distributed_locks
   SET expires_at = NOW() + ($2::TEXT || ' milliseconds')::INTERVAL
 WHERE holder_id = $1
`

// Refresh extends `expires_at` on every row tagged with `holderID`
// to `NOW() + ttl`. Use when a long-running operation risks crossing
// the original TTL boundary while still working.
func (s *PgsqlKeyedLockStore) Refresh(ctx context.Context, holderID string, ttl time.Duration) (int, error) {
	if ttl <= 0 {
		ttl = DefaultKeyedLockTTL
	}
	res, err := s.db.ExecContext(ctx, sqlKeyedLockRefresh,
		holderID,
		fmt.Sprintf("%d", ttl.Milliseconds()),
	)
	if err != nil {
		return 0, fmt.Errorf("cache.PgsqlKeyedLockStore.Refresh(%q): %w", holderID, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("cache.PgsqlKeyedLockStore.Refresh(%q): rows-affected: %w", holderID, err)
	}
	return int(n), nil
}

const sqlKeyedLockIsHeldBy = `
SELECT 1 FROM shared.distributed_locks
 WHERE holder_id = $1
   AND expires_at > NOW()
 LIMIT 1
`

// IsHeldBy reports whether `holderID` currently holds any LIVE lock
// (expired rows tagged with the same id don't count).
func (s *PgsqlKeyedLockStore) IsHeldBy(ctx context.Context, holderID string) (bool, error) {
	var x int
	err := s.db.QueryRowContext(ctx, sqlKeyedLockIsHeldBy, holderID).Scan(&x)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cache.PgsqlKeyedLockStore.IsHeldBy(%q): %w", holderID, err)
	}
	return true, nil
}

const sqlKeyedLockReapExpired = `
DELETE FROM shared.distributed_locks
 WHERE expires_at <= NOW()
`

// ReapExpired DELETEs every row whose lock has expired. Returns the
// number of rows deleted. Safe to run concurrently with Acquire /
// Release — the steal-on-acquire path uses the same `expires_at`
// predicate.
func (s *PgsqlKeyedLockStore) ReapExpired(ctx context.Context) (int, error) {
	res, err := s.db.ExecContext(ctx, sqlKeyedLockReapExpired)
	if err != nil {
		return 0, fmt.Errorf("cache.PgsqlKeyedLockStore.ReapExpired: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("cache.PgsqlKeyedLockStore.ReapExpired: rows-affected: %w", err)
	}
	return int(n), nil
}
