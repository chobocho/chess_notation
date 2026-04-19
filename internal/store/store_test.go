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
