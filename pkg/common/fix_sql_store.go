//go:build ignore

package common

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/quickfixgo/quickfix"
	"go.uber.org/zap"
)

// IdempotentSQLStore wraps a quickfix.MessageStore (backed by SQL) and overrides
// SaveMessage / SaveMessageAndIncrNextSenderMsgSeqNum to use INSERT ... ON CONFLICT
// DO UPDATE, preventing primary key violations on reconnect.
//
// This follows the same pattern as QuickFIX/J's JdbcStore which catches duplicate
// key errors and falls back to UPDATE. The FIX protocol spec is silent on store
// implementation details; tolerating duplicates is the industry best practice
// (QuickFIX/J, QuickFIX C++ FileStore, fix8 BDB all silently overwrite).
type IdempotentSQLStore struct {
	inner     quickfix.MessageStore
	db        *sql.DB
	sessionID quickfix.SessionID
}

// IdempotentSQLStoreFactory wraps a quickfix.MessageStoreFactory and produces
// IdempotentSQLStore instances that use UPSERT for message persistence.
type IdempotentSQLStoreFactory struct {
	inner    quickfix.MessageStoreFactory
	db       *sql.DB
	logLabel string
}

// NewIdempotentSQLStoreFactory creates a factory that wraps the given SQL store factory.
// The db connection is used for the UPSERT queries (bypassing the inner store's INSERT).
func NewIdempotentSQLStoreFactory(inner quickfix.MessageStoreFactory, db *sql.DB, logLabel string) *IdempotentSQLStoreFactory {
	return &IdempotentSQLStoreFactory{inner: inner, db: db, logLabel: logLabel}
}

// Create delegates to the inner factory, then wraps the result.
func (f *IdempotentSQLStoreFactory) Create(sessionID quickfix.SessionID) (quickfix.MessageStore, error) {
	innerStore, err := f.inner.Create(sessionID)
	if err != nil {
		return nil, err
	}
	L.Info(fmt.Sprintf("%s: using idempotent SQL store wrapper for session %s", f.logLabel, sessionID))
	return &IdempotentSQLStore{
		inner:     innerStore,
		db:        f.db,
		sessionID: sessionID,
	}, nil
}

// SaveMessage persists a message using INSERT ... ON CONFLICT DO UPDATE.
// Logs a warning when a duplicate sequence number is detected (xmax != 0 means row was updated, not inserted).
func (s *IdempotentSQLStore) SaveMessage(seqNum int, msg []byte) error {
	var xmax uint32
	err := s.db.QueryRow(
		`INSERT INTO messages (
			msgseqnum, message,
			beginstring, session_qualifier,
			sendercompid, sendersubid, senderlocid,
			targetcompid, targetsubid, targetlocid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT ON CONSTRAINT messages_pkey
		DO UPDATE SET message = EXCLUDED.message
		RETURNING xmax`,
		seqNum, string(msg),
		s.sessionID.BeginString, s.sessionID.Qualifier,
		s.sessionID.SenderCompID, s.sessionID.SenderSubID, s.sessionID.SenderLocationID,
		s.sessionID.TargetCompID, s.sessionID.TargetSubID, s.sessionID.TargetLocationID,
	).Scan(&xmax)
	if err != nil {
		return err
	}
	if xmax != 0 {
		L.Warn("duplicate seq number detected in message store (overwritten)",
			zap.Int("seq_num", seqNum),
			zap.String("session_id", s.sessionID.String()))
	}
	return nil
}

// SaveMessageAndIncrNextSenderMsgSeqNum atomically saves the message (with UPSERT)
// and increments the outgoing sequence number, matching the transactional behavior
// of the original SQL store (PR #525).
func (s *IdempotentSQLStore) SaveMessageAndIncrNextSenderMsgSeqNum(seqNum int, msg []byte) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var xmax uint32
	err = tx.QueryRow(
		`INSERT INTO messages (
			msgseqnum, message,
			beginstring, session_qualifier,
			sendercompid, sendersubid, senderlocid,
			targetcompid, targetsubid, targetlocid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT ON CONSTRAINT messages_pkey
		DO UPDATE SET message = EXCLUDED.message
		RETURNING xmax`,
		seqNum, string(msg),
		s.sessionID.BeginString, s.sessionID.Qualifier,
		s.sessionID.SenderCompID, s.sessionID.SenderSubID, s.sessionID.SenderLocationID,
		s.sessionID.TargetCompID, s.sessionID.TargetSubID, s.sessionID.TargetLocationID,
	).Scan(&xmax)
	if err != nil {
		return err
	}
	if xmax != 0 {
		L.Warn("duplicate seq number detected in message store (overwritten)",
			zap.Int("seq_num", seqNum),
			zap.String("session_id", s.sessionID.String()))
	}

	next := s.inner.NextSenderMsgSeqNum() + 1
	_, err = tx.Exec(
		`UPDATE sessions SET outgoing_seqnum = $1
		WHERE beginstring=$2 AND session_qualifier=$3
		AND sendercompid=$4 AND sendersubid=$5 AND senderlocid=$6
		AND targetcompid=$7 AND targetsubid=$8 AND targetlocid=$9`,
		next, s.sessionID.BeginString, s.sessionID.Qualifier,
		s.sessionID.SenderCompID, s.sessionID.SenderSubID, s.sessionID.SenderLocationID,
		s.sessionID.TargetCompID, s.sessionID.TargetSubID, s.sessionID.TargetLocationID,
	)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return s.inner.SetNextSenderMsgSeqNum(next)
}

// All remaining methods delegate to the inner store.

func (s *IdempotentSQLStore) NextSenderMsgSeqNum() int {
	return s.inner.NextSenderMsgSeqNum()
}

func (s *IdempotentSQLStore) NextTargetMsgSeqNum() int {
	return s.inner.NextTargetMsgSeqNum()
}

func (s *IdempotentSQLStore) IncrNextSenderMsgSeqNum() error {
	return s.inner.IncrNextSenderMsgSeqNum()
}

func (s *IdempotentSQLStore) IncrNextTargetMsgSeqNum() error {
	return s.inner.IncrNextTargetMsgSeqNum()
}

func (s *IdempotentSQLStore) SetNextSenderMsgSeqNum(next int) error {
	return s.inner.SetNextSenderMsgSeqNum(next)
}

func (s *IdempotentSQLStore) SetNextTargetMsgSeqNum(next int) error {
	return s.inner.SetNextTargetMsgSeqNum(next)
}

func (s *IdempotentSQLStore) CreationTime() time.Time {
	return s.inner.CreationTime()
}

func (s *IdempotentSQLStore) SetCreationTime(t time.Time) {
	s.inner.SetCreationTime(t)
}

func (s *IdempotentSQLStore) GetMessages(beginSeqNum, endSeqNum int) ([][]byte, error) {
	return s.inner.GetMessages(beginSeqNum, endSeqNum)
}

func (s *IdempotentSQLStore) Refresh() error {
	return s.inner.Refresh()
}

func (s *IdempotentSQLStore) Reset() error {
	return s.inner.Reset()
}

func (s *IdempotentSQLStore) Close() error {
	return s.inner.Close()
}

// OpenFixStoreDB opens a PostgreSQL connection for the idempotent store wrapper.
// Returns nil DB (no error) if pgsqlURL is empty.
//
// FIXREL2 Phase 7 — explicit connection caps so multiple FIX daemons
// (5 fixreceiver versions × N namespaces) cannot exhaust Postgres
// max_connections. Without these the pool defaulted to "unlimited" which
// scaled badly under load.
func OpenFixStoreDB(pgsqlURL, label string) *sql.DB {
	if pgsqlURL == "" {
		return nil
	}
	db, err := sql.Open("postgres", pgsqlURL)
	if err != nil {
		L.Warn("failed to open DB for idempotent store",
			zap.String("label", label), zap.Error(err))
		return nil
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)
	return db
}
