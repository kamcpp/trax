package common

import (
	"strings"
	"testing"
)

// TestSplitSQLStatements_SemicolonInsideSingleQuotedString — regression
// guard for a bug where the splitter mis-treated `;` inside a SQL string
// literal as a statement terminator. A description text containing
// `finalize_to_erc20=true; runs for Withdraw` ended up cleaved at the
// `;`, producing two halves with mid-string apostrophes that postgres
// rejected with "unterminated quoted string at or near ...".
func TestSplitSQLStatements_SemicolonInsideSingleQuotedString(t *testing.T) {
	const sql = `INSERT INTO t (description) VALUES ('first half; second half');
INSERT INTO t (description) VALUES ('next row');`

	got := splitSQLStatements(sql)
	if len(got) != 2 {
		t.Fatalf("expected 2 statements, got %d: %#v", len(got), got)
	}
	if !strings.Contains(got[0], "first half; second half") {
		t.Errorf("first statement should keep the in-literal `;`; got: %q", got[0])
	}
	if !strings.Contains(got[1], "next row") {
		t.Errorf("second statement should be the next INSERT; got: %q", got[1])
	}
}

// TestSplitSQLStatements_DoubledQuoteEscape — single-quoted string
// literals in postgres escape an apostrophe by doubling it (`it”s`).
// A doubled quote must NOT close the string and re-open it; the splitter
// must treat the whole token as still being inside the literal.
func TestSplitSQLStatements_DoubledQuoteEscape(t *testing.T) {
	const sql = `INSERT INTO t (name) VALUES ('it''s; fine');
INSERT INTO t (name) VALUES ('ok');`

	got := splitSQLStatements(sql)
	if len(got) != 2 {
		t.Fatalf("expected 2 statements, got %d: %#v", len(got), got)
	}
	if !strings.Contains(got[0], "it''s; fine") {
		t.Errorf("escaped-apostrophe literal must keep the `;`; got: %q", got[0])
	}
}

// TestSplitSQLStatements_DollarQuotedBlockStillWorks — the prior
// behavior must be preserved: $$-delimited plpgsql bodies count as one
// statement even with `;` inside them.
func TestSplitSQLStatements_DollarQuotedBlockStillWorks(t *testing.T) {
	const sql = `CREATE FUNCTION f() RETURNS void AS $$
BEGIN
  PERFORM 1;
  PERFORM 2;
END;
$$ LANGUAGE plpgsql;
SELECT f();`

	got := splitSQLStatements(sql)
	if len(got) != 2 {
		t.Fatalf("expected 2 statements (CREATE FUNCTION + SELECT), got %d: %#v", len(got), got)
	}
	if !strings.Contains(got[0], "PERFORM 1") || !strings.Contains(got[0], "PERFORM 2") {
		t.Errorf("CREATE FUNCTION body must stay intact; got: %q", got[0])
	}
}

// TestSplitSQLStatements_LineCommentWithSemicolon — `;` inside a `--`
// line comment must NOT terminate the surrounding statement.
func TestSplitSQLStatements_LineCommentWithSemicolon(t *testing.T) {
	const sql = `SELECT 1 -- ignored; still part of SELECT
;
SELECT 2;`

	got := splitSQLStatements(sql)
	if len(got) != 2 {
		t.Fatalf("expected 2 statements, got %d: %#v", len(got), got)
	}
}
