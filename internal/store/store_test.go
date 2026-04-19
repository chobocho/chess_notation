package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chobocho/chess_notation/internal/pgn"
)

const sample = `[Event "Test"]
[Site "?"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 4. Ba4 Nf6 5. O-O 1-0
`

func TestImportAndList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	games, err := pgn.ParseString(sample)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	id, err := s.ImportGame(ctx, games[0], sample)
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatalf("bad id")
	}

	metas, err := s.ListGames(ctx, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(metas) != 1 {
		t.Fatalf("got %d games", len(metas))
	}
	if metas[0].White != "Alice" || metas[0].Result != "1-0" {
		t.Fatalf("meta = %+v", metas[0])
	}
	if metas[0].PlyCount != 9 {
		t.Fatalf("ply count = %d, want 9", metas[0].PlyCount)
	}

	moves, err := s.GetMoves(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 10 { // 1 for ply 0 + 9 moves
		t.Fatalf("got %d move rows, want 10", len(moves))
	}
	if moves[0].Ply != 0 || moves[0].SAN.Valid {
		t.Fatalf("ply-0 row bad: %+v", moves[0])
	}
	if !moves[5].SAN.Valid || moves[5].SAN.String != "Bb5" {
		t.Fatalf("ply-5 san = %+v", moves[5].SAN)
	}

	// Bookmark CRUD.
	bmID, err := s.AddBookmark(ctx, id, 5, "interesting")
	if err != nil {
		t.Fatal(err)
	}
	bms, err := s.ListBookmarks(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(bms) != 1 || bms[0].ID != bmID || bms[0].Ply != 5 {
		t.Fatalf("bookmarks = %+v", bms)
	}
	if err := s.DeleteBookmark(ctx, bmID); err != nil {
		t.Fatal(err)
	}
	bms, _ = s.ListBookmarks(ctx, id)
	if len(bms) != 0 {
		t.Fatalf("bookmarks not deleted: %+v", bms)
	}
}

func TestGetFENAt(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	games, _ := pgn.ParseString(sample)
	ctx := context.Background()
	id, _ := s.ImportGame(ctx, games[0], sample)
	fen, err := s.GetFENAt(ctx, id, 0)
	if err != nil {
		t.Fatal(err)
	}
	if fen == "" {
		t.Fatalf("empty fen")
	}
}

const sampleAlt = `[Event "Alt"]
[Site "?"]
[White "Carol"]
[Black "Dan"]
[Result "0-1"]

1. d4 d5 2. c4 e6 0-1
`

const sampleDraw = `[Event "Draw"]
[Site "?"]
[White "Alice"]
[Black "Carol"]
[Result "1/2-1/2"]

1. e4 c5 1/2-1/2
`

func importAll(t *testing.T, s *Store) {
	t.Helper()
	ctx := context.Background()
	for _, raw := range []string{sample, sampleAlt, sampleDraw} {
		gs, err := pgn.ParseString(raw)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.ImportGame(ctx, gs[0], raw); err != nil {
			t.Fatal(err)
		}
	}
}

func TestListGamesFiltered(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "f.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	importAll(t, s)
	ctx := context.Background()

	// White filter (case-insensitive substring).
	got, err := s.ListGamesFiltered(ctx, ListFilter{White: "alice"}, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("white=alice: %d rows, want 2", len(got))
	}

	// Result filter (exact).
	got, err = s.ListGamesFiltered(ctx, ListFilter{Result: "1/2-1/2"}, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Result != "1/2-1/2" {
		t.Fatalf("result filter: %+v", got)
	}

	// Combined White+Result, no match.
	got, err = s.ListGamesFiltered(ctx, ListFilter{White: "alice", Result: "0-1"}, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("combined no-match: %+v", got)
	}

	// Black filter.
	got, err = s.ListGamesFiltered(ctx, ListFilter{Black: "Dan"}, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Black != "Dan" {
		t.Fatalf("black=Dan: %+v", got)
	}
}

func TestCountGames(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "c.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	importAll(t, s)
	ctx := context.Background()

	n, err := s.CountGames(ctx, ListFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("total = %d", n)
	}
	n, err = s.CountGames(ctx, ListFilter{White: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("white=alice count = %d", n)
	}
	n, err = s.CountGames(ctx, ListFilter{Result: "unknown"})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("no-match count = %d", n)
	}
}

func TestSchemaVersionOnFresh(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	v, err := s.SchemaVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != 1 {
		t.Fatalf("version = %d, want 1", v)
	}
	// schema_migrations has exactly one row.
	var n int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("schema_migrations rows = %d, want 1", n)
	}
	var name string
	if err := s.DB.QueryRow(`SELECT name FROM schema_migrations WHERE version = 1`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "initial" {
		t.Fatalf("name = %q", name)
	}
}

func TestSchemaReopenIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "r.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	// Write some data so we can confirm it survives reopen.
	importAll(t, s)
	s.Close()

	s2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()

	v, err := s2.SchemaVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != 1 {
		t.Fatalf("after reopen version = %d", v)
	}
	var rows int
	if err := s2.DB.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("reopen duplicated schema_migrations rows: %d", rows)
	}
	n, err := s2.CountGames(context.Background(), ListFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("games after reopen = %d, want 3", n)
	}
}

func TestSchemaAppliesV2Migration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "v2.db")
	// Open with just the baseline to seed the DB.
	s, err := openWithMigrations(path, migrations)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()

	// Reopen with an extra v2 migration that adds a new column.
	extra := append([]migration{}, migrations...)
	extra = append(extra, migration{
		Version: 2,
		Name:    "add_games_notes",
		SQL:     `ALTER TABLE games ADD COLUMN notes TEXT`,
	})
	s2, err := openWithMigrations(path, extra)
	if err != nil {
		t.Fatalf("apply v2: %v", err)
	}
	defer s2.Close()

	v, err := s2.SchemaVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != 2 {
		t.Fatalf("version after v2 = %d", v)
	}

	// Third reopen is a no-op.
	s2.Close()
	s3, err := openWithMigrations(path, extra)
	if err != nil {
		t.Fatal(err)
	}
	defer s3.Close()
	var rows int
	if err := s3.DB.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 2 {
		t.Fatalf("re-reopen rows = %d, want 2", rows)
	}

	// New column should be queryable.
	if _, err := s3.DB.Exec(`UPDATE games SET notes = 'hello' WHERE id = 1`); err != nil {
		t.Fatalf("update notes: %v", err)
	}
}

func TestSchemaRejectsOutOfOrderMigrations(t *testing.T) {
	bad := []migration{
		{Version: 2, Name: "b", SQL: `SELECT 1`},
		{Version: 1, Name: "a", SQL: `SELECT 1`},
	}
	_, err := openWithMigrations(filepath.Join(t.TempDir(), "bad.db"), bad)
	if err == nil {
		t.Fatalf("expected error on out-of-order migrations")
	}
}

func TestPagination(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "p.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	importAll(t, s)
	ctx := context.Background()

	page1, err := s.ListGames(ctx, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 = %d", len(page1))
	}
	page2, err := s.ListGames(ctx, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2) != 1 {
		t.Fatalf("page2 = %d", len(page2))
	}
	if page1[0].ID == page2[0].ID {
		t.Fatalf("pages overlap")
	}
}
