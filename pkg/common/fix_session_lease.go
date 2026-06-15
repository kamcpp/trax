package common

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FIXREL2 Phase 7 — split-brain protection.
//
// Single-replica is policy on every FIX role daemon (Recreate strategy,
// replicas=1), but an accidental kubectl scale or a stuck-terminating pod
// can produce two writers on the same FIX session. Two writers corrupt the
// outgoing sequence number stream irrecoverably. The lease is cheap insurance:
// before the new pod takes any FIX action, it must own a row in public.fix_session_lock
// keyed on a canonical session identifier; another live owner means the new
// pod aborts startup with a clear error.

// DefaultFIXLeaseTTL bounds the time a crashed pod's lease blocks a healthy
// successor. The renewer fires every TTL/3.
const DefaultFIXLeaseTTL = 30 * time.Second

// FixSessionLease tracks ownership of a single fix_session_lock row.
//
// Lifecycle:
//
//	lease := NewFixSessionLease(db, "FIX.4.2:BROKER->EXCHANGE", "fixclient", DefaultFIXLeaseTTL)
//	if err := lease.Acquire(ctx); err != nil { Fatal }   // fail fast on conflict
//	lease.StartRenewer(ctx)                              // background renewal
//	... do FIX work ...
//	lease.Release(ctx)                                   // graceful shutdown
type FixSessionLease struct {
	db        *sql.DB
	sessionID string
	role      string
	ownerID   string
	ttl       time.Duration

	mu             sync.Mutex
	renewerCancel  context.CancelFunc
	renewerStopped chan struct{}
}

// NewFixSessionLease constructs a lease handle. ownerID is generated from the
// pod hostname + a random suffix so a pod that restarts in place produces a
// fresh owner that won't collide with its previous incarnation's row.
func NewFixSessionLease(db *sql.DB, sessionID, role string, ttl time.Duration) *FixSessionLease {
	if ttl <= 0 {
		ttl = DefaultFIXLeaseTTL
	}
	return &FixSessionLease{
		db:        db,
		sessionID: sessionID,
		role:      role,
		ownerID:   buildLeaseOwnerID(),
		ttl:       ttl,
	}
}

// Acquire takes the lock. Returns nil when this process owns it; a non-nil
// error when the row is held by someone else and not yet expired.
//
// SQL semantics: INSERT ... ON CONFLICT DO UPDATE only when the existing row
// has expired (expires_at < now()). RETURNING owner_id tells us who we ended
// up as — if it is not us, somebody else holds it.
//
// Logs INFO on success; renewals call acquireOnce() directly to avoid the
// every-ttl/3 log flood (~360 lines/hour/pod).
func (l *FixSessionLease) Acquire(ctx context.Context) error {
	if err := l.acquireOnce(ctx); err != nil {
		return err
	}
	L.Info("fix session lease acquired",
		zap.String("session_id", l.sessionID),
		zap.String("role", l.role),
		zap.String("owner_id", l.ownerID),
		zap.Duration("ttl", l.ttl))
	return nil
}

// acquireOnce runs the INSERT-or-update without logging. Used by both the
// public Acquire (which adds the INFO log) and the renewer (which logs only
// on failure).
func (l *FixSessionLease) acquireOnce(ctx context.Context) error {
	expiresAt := time.Now().Add(l.ttl)
	var ownerOut string
	err := l.db.QueryRowContext(ctx,
		`INSERT INTO public.fix_session_lock (session_id, role, owner_id, expires_at, updated_at)
		 VALUES ($1, $2, $3, $4, now())
		 ON CONFLICT (session_id) DO UPDATE
		   SET role       = EXCLUDED.role,
		       owner_id   = EXCLUDED.owner_id,
		       expires_at = EXCLUDED.expires_at,
		       updated_at = now()
		   WHERE public.fix_session_lock.expires_at < now()
		      OR public.fix_session_lock.owner_id   = EXCLUDED.owner_id
		 RETURNING owner_id`,
		l.sessionID, l.role, l.ownerID, expiresAt,
	).Scan(&ownerOut)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// CONFLICT row exists, expires_at is still in the future, and the
			// existing owner_id is not us — somebody else holds the lease.
			held, holdErr := l.queryHolder(ctx)
			if holdErr != nil {
				return fmt.Errorf("fix lease %q: held by another owner; lookup failed: %w", l.sessionID, holdErr)
			}
			return fmt.Errorf("fix lease %q: held by %q (role=%q, expires_at=%s); refusing to start",
				l.sessionID, held.OwnerID, held.Role, held.ExpiresAt.Format(time.RFC3339))
		}
		return fmt.Errorf("fix lease %q: acquire failed: %w", l.sessionID, err)
	}

	if ownerOut != l.ownerID {
		// Defensive — should not happen under the WHERE clause, but make
		// it explicit so the failure mode is obvious if SQL semantics shift.
		return fmt.Errorf("fix lease %q: acquire raced; current owner=%q (we are %q)", l.sessionID, ownerOut, l.ownerID)
	}
	return nil
}

// StartRenewer launches a background goroutine that re-Acquires the lease
// every ttl/3. If renewal ever fails because another owner has taken the
// row, the goroutine logs a fatal-class error and returns; the daemon's
// own health checks should then fail and the pod restart.
//
// Intentionally not a pure UPDATE — Acquire's INSERT-or-update semantics
// also recover the row if someone DELETEd it (e.g., manual ops cleanup).
func (l *FixSessionLease) StartRenewer(parentCtx context.Context) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.renewerCancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(parentCtx)
	l.renewerCancel = cancel
	l.renewerStopped = make(chan struct{})

	interval := l.ttl / 3
	if interval < time.Second {
		interval = time.Second
	}

	go func() {
		defer close(l.renewerStopped)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := l.acquireOnce(ctx); err != nil {
					L.Error("fix session lease renewal failed — another owner may have taken the lock",
						zap.String("session_id", l.sessionID),
						zap.String("role", l.role),
						zap.String("owner_id", l.ownerID),
						zap.Error(err))
					// Stop renewing; let liveness probes notice.
					return
				}
			}
		}
	}()
}

// Release deletes the lease row when (and only when) we still own it.
// Idempotent. Stops the renewer first.
func (l *FixSessionLease) Release(ctx context.Context) error {
	l.mu.Lock()
	if l.renewerCancel != nil {
		l.renewerCancel()
		l.renewerCancel = nil
	}
	stopped := l.renewerStopped
	l.mu.Unlock()
	if stopped != nil {
		<-stopped
	}

	_, err := l.db.ExecContext(ctx,
		`DELETE FROM public.fix_session_lock
		 WHERE session_id = $1 AND owner_id = $2`,
		l.sessionID, l.ownerID,
	)
	if err != nil {
		return fmt.Errorf("fix lease %q: release failed: %w", l.sessionID, err)
	}
	L.Info("fix session lease released",
		zap.String("session_id", l.sessionID),
		zap.String("owner_id", l.ownerID))
	return nil
}

// holder is what we return on a held-by-another acquire failure.
type holder struct {
	OwnerID   string
	Role      string
	ExpiresAt time.Time
}

func (l *FixSessionLease) queryHolder(ctx context.Context) (holder, error) {
	var h holder
	err := l.db.QueryRowContext(ctx,
		`SELECT owner_id, role, expires_at FROM public.fix_session_lock WHERE session_id = $1`,
		l.sessionID,
	).Scan(&h.OwnerID, &h.Role, &h.ExpiresAt)
	if err != nil {
		return h, err
	}
	return h, nil
}

// buildLeaseOwnerID returns "<hostname>-<8 hex chars>", e.g.
// "fixclient-abc123def456-7c4a9f00". The random suffix prevents collisions
// when a pod restarts in place before its previous lease expires (the
// previous incarnation's row would still be there with the previous random
// suffix, and the WHERE-expires_at clause keeps us from stomping it before
// TTL elapses).
func buildLeaseOwnerID() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknown-host"
	}
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Extremely unlikely, but keep an obvious marker.
		return host + "-norand"
	}
	return host + "-" + hex.EncodeToString(b[:])
}
