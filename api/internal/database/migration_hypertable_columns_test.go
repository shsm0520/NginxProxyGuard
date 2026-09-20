package database

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// This is the test that should have existed before #135, #55 and #297.
//
// `logs_partitioned` is the one table in this schema with FOUR migration sites
// instead of the documented three. migrateToTimescaleDB() builds a brand-new
// `logs_hypertable` from a column list typed out by hand, copies the rows
// across a second hand-typed list, and then RENAMEs the result over
// `logs_partitioned`. It runs on fresh installs. So a column added only to
// 001_init.sql and the `upgrades` slice exists right up until that swap and
// then silently disappears, taking every subsequent log INSERT with it
// ("column does not exist") — which is exactly how #135 shipped.
//
// #297 is the same trap one turn further in: the column WAS in all four
// places, but the hypertable copy declared `rule_id integer` where
// 001_init.sql declares `bigint`. Names matching is not enough, because
// ADD COLUMN IF NOT EXISTS never repairs a type — a narrow type installed by
// the swap is permanent. So this test compares types too, which is what turns
// #297 from "fixed once" into "cannot recur".
//
// It reads the two source files as text rather than a live database on
// purpose: the failure it guards against is invisible on any box that already
// carries the `timescaledb_hypertable` marker, i.e. every developer machine.

const (
	initSQLPath      = "migrations/001_init.sql"
	migrationGoPath  = "migration.go"
	hypertableCreate = "CREATE TABLE IF NOT EXISTS logs_hypertable ("
	partitionedCreat = "CREATE TABLE IF NOT EXISTS public.logs_partitioned ("
	hypertableInsert = "INSERT INTO logs_hypertable ("
)

// column is one parsed column declaration: its name and its normalised type.
type column struct {
	name string
	typ  string
}

func TestLogsPartitionedColumnListsAgreeAcrossAllFourMigrationSites(t *testing.T) {
	goSrc := readSource(t, migrationGoPath)
	sqlSrc := readSource(t, initSQLPath)

	initCols := parseCreateTableColumns(t, sqlSrc, partitionedCreat)
	hyperCols := parseCreateTableColumns(t, goSrc, hypertableCreate)

	if len(initCols) == 0 || len(hyperCols) == 0 {
		t.Fatalf("parsed no columns (init=%d hypertable=%d) — the anchors in this test have drifted from the source",
			len(initCols), len(hyperCols))
	}

	// 1. The two CREATE lists must describe the same table, name for name.
	assertSameNames(t, "001_init.sql logs_partitioned", names(initCols),
		"migration.go logs_hypertable", names(hyperCols))

	// 2. ...and type for type. A name-only match still ships #297.
	initByName := byName(initCols)
	for _, c := range hyperCols {
		want, ok := initByName[c.name]
		if !ok {
			continue // already reported by assertSameNames
		}
		if want.typ != c.typ {
			t.Errorf("column %q type mismatch: 001_init.sql declares %q but migrateToTimescaleDB() creates %q.\n"+
				"The hypertable is RENAMEd over logs_partitioned on fresh installs, so the narrower\n"+
				"declaration is the one those installs keep forever (#297).",
				c.name, want.typ, c.typ)
		}
	}

	// 3. The copy step names every column explicitly, twice — the INSERT target
	//    list and the SELECT source list. Either one drifting loses a column's
	//    data during the swap without any error at all.
	insertCols := parseParenNameList(t, goSrc, hypertableInsert)
	assertSameNames(t, "migration.go logs_hypertable CREATE", names(hyperCols),
		"migration.go INSERT INTO logs_hypertable (...)", insertCols)

	selectCols := parseCopySelectColumns(t, goSrc)
	assertSameNames(t, "migration.go INSERT INTO logs_hypertable (...)", insertCols,
		"migration.go SELECT lp.* source list", selectCols)

	// 4. The two copy lists are positional: Postgres matches them by order, not
	//    by name, so agreeing as sets is not enough.
	for i := range insertCols {
		if i < len(selectCols) && insertCols[i] != selectCols[i] {
			t.Errorf("copy lists are positionally misaligned at index %d: INSERT names %q but SELECT supplies %q.\n"+
				"Postgres matches these by position, so this silently writes one column's values into another.",
				i, insertCols[i], selectCols[i])
		}
	}
}

// readSource loads a file the test compares as text. A missing file is a
// failure, not a skip: a silent skip here would restore the exact blind spot
// this test exists to remove.
func readSource(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	return string(b)
}

// stripSQLComments removes `-- ...` line comments. Both column lists carry
// explanatory comments, and the #297 one contains "(2147483647)" — unbalanced
// parens in a comment would derail the scanner below.
func stripSQLComments(s string) string {
	var out strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i]
		}
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String()
}

// balancedBody returns the text between the parenthesis that ends `anchor` and
// its matching close.
func balancedBody(t *testing.T, src, anchor string) string {
	t.Helper()
	i := strings.Index(src, anchor)
	if i < 0 {
		t.Fatalf("anchor %q not found — this test's anchors have drifted from the source", anchor)
	}
	rest := src[i+len(anchor):]
	depth := 1
	for j, r := range rest {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return rest[:j]
			}
		}
	}
	t.Fatalf("unbalanced parentheses after anchor %q", anchor)
	return ""
}

// splitTopLevel splits on commas that are not inside parentheses, so
// `varchar(2)` and `numeric(10,6)` survive intact.
func splitTopLevel(body string) []string {
	var parts []string
	depth, start := 0, 0
	for i, r := range body {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, body[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, body[start:])
	return parts
}

// parseCreateTableColumns pulls (name, type) out of a CREATE TABLE body.
func parseCreateTableColumns(t *testing.T, src, anchor string) []column {
	t.Helper()
	body := stripSQLComments(balancedBody(t, src, anchor))

	var cols []column
	for _, part := range splitTopLevel(body) {
		field := strings.TrimSpace(strings.Join(strings.Fields(part), " "))
		if field == "" {
			continue
		}
		name, rest, ok := strings.Cut(field, " ")
		if !ok {
			continue
		}
		cols = append(cols, column{name: normalizeName(name), typ: normalizeType(rest)})
	}
	return cols
}

// parseParenNameList pulls a bare comma-separated column list — the INSERT
// target list — out of the parentheses following `anchor`.
func parseParenNameList(t *testing.T, src, anchor string) []string {
	t.Helper()
	body := stripSQLComments(balancedBody(t, src, anchor))

	var out []string
	for _, part := range splitTopLevel(body) {
		if name := normalizeName(strings.TrimSpace(part)); name != "" {
			out = append(out, name)
		}
	}
	return out
}

var lpColumnRef = regexp.MustCompile(`\blp\.([a-zA-Z_][a-zA-Z0-9_]*)`)

// parseCopySelectColumns pulls the `lp.<column>` source list out of the copy
// step's SELECT, stopping at FROM so the WHERE clause's own lp.* references
// are not mistaken for projected columns.
func parseCopySelectColumns(t *testing.T, src string) []string {
	t.Helper()
	i := strings.Index(src, hypertableInsert)
	if i < 0 {
		t.Fatalf("anchor %q not found", hypertableInsert)
	}
	rest := src[i:]
	sel := strings.Index(rest, "SELECT")
	from := strings.Index(rest, "FROM logs_partitioned")
	if sel < 0 || from < 0 || from < sel {
		t.Fatalf("could not locate the SELECT ... FROM logs_partitioned copy step")
	}

	var out []string
	for _, m := range lpColumnRef.FindAllStringSubmatch(rest[sel:from], -1) {
		out = append(out, normalizeName(m[1]))
	}
	return out
}

// normalizeName lowercases and unquotes an identifier. `"timestamp"` has to be
// quoted in 001_init.sql because it is a reserved word, and is unquoted in the
// Go source.
func normalizeName(s string) string {
	return strings.ToLower(strings.Trim(strings.TrimSpace(s), `"`))
}

var typeAliases = strings.NewReplacer(
	"character varying", "varchar",
	"timestamp with time zone", "timestamptz",
	"public.", "",
)

// normalizeType reduces a column definition's tail to just its type, so that
// `uuid DEFAULT gen_random_uuid() NOT NULL` and `uuid NOT NULL DEFAULT
// gen_random_uuid()` compare equal. Defaults and nullability genuinely differ
// between the two sites (the hypertable copy supplies values explicitly), and
// neither can strand an install the way a type can.
func normalizeType(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, clause := range []string{" default ", " not null", " null", " collate ", " generated "} {
		if i := strings.Index(s, clause); i >= 0 {
			s = s[:i]
		}
	}
	s = typeAliases.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func names(cols []column) []string {
	out := make([]string, 0, len(cols))
	for _, c := range cols {
		out = append(out, c.name)
	}
	return out
}

func byName(cols []column) map[string]column {
	out := make(map[string]column, len(cols))
	for _, c := range cols {
		out[c.name] = c
	}
	return out
}

// assertSameNames reports the two directions separately, because which side a
// column is missing from decides what breaks: absent from the hypertable list
// means fresh installs lose the column at the swap; absent from 001_init.sql
// means the upgrade path never gains it.
func assertSameNames(t *testing.T, leftLabel string, left []string, rightLabel string, right []string) {
	t.Helper()
	if missing := difference(left, right); len(missing) > 0 {
		t.Errorf("%s has column(s) %s that %s does not.\n%s",
			leftLabel, strings.Join(missing, ", "), rightLabel, fourSiteHint)
	}
	if extra := difference(right, left); len(extra) > 0 {
		t.Errorf("%s has column(s) %s that %s does not.\n%s",
			rightLabel, strings.Join(extra, ", "), leftLabel, fourSiteHint)
	}
}

const fourSiteHint = "  logs_partitioned has FOUR migration sites, not three — see api/CLAUDE.md step 3.\n" +
	"  A column must appear in: 001_init.sql CREATE, the `upgrades` slice, and BOTH hand-typed\n" +
	"  lists inside migrateToTimescaleDB() (the logs_hypertable CREATE and the INSERT...SELECT)."

func difference(a, b []string) []string {
	inB := make(map[string]struct{}, len(b))
	for _, s := range b {
		inB[s] = struct{}{}
	}
	var out []string
	for _, s := range a {
		if _, ok := inB[s]; !ok {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// TestHypertableColumnParserActuallyParses is a canary for the test above: if
// the anchors drift or the parser silently degrades to zero columns, the
// comparison would pass vacuously and the guard would be gone without a single
// red test. So assert the shape of what was parsed, not just that it matched.
func TestHypertableColumnParserActuallyParses(t *testing.T) {
	goSrc := readSource(t, migrationGoPath)
	cols := parseCreateTableColumns(t, goSrc, hypertableCreate)

	const minColumns = 30
	if len(cols) < minColumns {
		t.Fatalf("parsed only %d columns from the logs_hypertable CREATE; expected at least %d — the parser or the anchor is broken",
			len(cols), minColumns)
	}

	// Spot-check the two columns whose declarations carry the traps this file
	// guards: rule_id is the #297 type, and geo_country_code is the only
	// parenthesised type in the list.
	got := byName(cols)
	for name, wantType := range map[string]string{
		"rule_id":          "bigint",
		"geo_country_code": "varchar(2)",
		"id":               "uuid",
		"timestamp":        "timestamptz",
	} {
		c, ok := got[name]
		if !ok {
			t.Errorf("expected column %q in the parsed logs_hypertable list", name)
			continue
		}
		if c.typ != wantType {
			t.Errorf("column %q parsed as type %q, want %q", name, c.typ, wantType)
		}
	}

	if t.Failed() {
		t.Log(fmt.Sprintf("parsed %d columns: %v", len(cols), names(cols)))
	}
}
