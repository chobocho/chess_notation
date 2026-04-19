package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chobocho/chess_notation/internal/chess"
	"github.com/chobocho/chess_notation/internal/pgn"
	"github.com/chobocho/chess_notation/internal/store"
)

func parseFENForTest(t *testing.T, fen string) (*chess.Position, error) {
	t.Helper()
	return chess.ParseFEN(fen)
}

const sampleA = `[Event "A"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 1-0
`

const sampleB = `[Event "B"]
[White "Carol"]
[Black "Dan"]
[Result "0-1"]

1. d4 d5 0-1
`

const sampleC = `[Event "C"]
[White "Eve"]
[Black "Frank"]
[Result "1/2-1/2"]

1. c4 c5 1/2-1/2
`

func newTestServer(t *testing.T) (*Server, *store.Store) {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "web.db"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, raw := range []string{sampleA, sampleB, sampleC} {
		gs, err := pgn.ParseString(raw)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.ImportGame(ctx, gs[0], raw); err != nil {
			t.Fatal(err)
		}
	}
	srv, err := NewServer(s)
	if err != nil {
		t.Fatal(err)
	}
	return srv, s
}

func TestIndexListsAllGames(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Store.Close()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	body := rr.Body.String()
	for _, name := range []string{"Alice", "Carol", "Eve", "Games (3)"} {
		if !strings.Contains(body, name) {
			t.Errorf("body missing %q", name)
		}
	}
}

func TestIndexWhiteFilter(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Store.Close()
	req := httptest.NewRequest(http.MethodGet, "/?white=alice", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	if !strings.Contains(body, "Alice") {
		t.Fatalf("Alice missing")
	}
	if strings.Contains(body, ">Carol<") || strings.Contains(body, ">Eve<") {
		t.Fatalf("Carol/Eve should be filtered out. Body:\n%s", body)
	}
	if !strings.Contains(body, "Games (1)") {
		t.Fatalf("count header wrong")
	}
}

func TestIndexResultFilter(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Store.Close()
	req := httptest.NewRequest(http.MethodGet, "/?result=1%2F2-1%2F2", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	if !strings.Contains(body, "Eve") {
		t.Fatalf("Eve missing")
	}
	if strings.Contains(body, ">Alice<") {
		t.Fatalf("Alice should be filtered out")
	}
}

func TestIndexPagination(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Store.Close()

	req := httptest.NewRequest(http.MethodGet, "/?per=2&page=1", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	if !strings.Contains(body, "Page 1 / 2") {
		t.Fatalf("page indicator: body=\n%s", body)
	}
	if !strings.Contains(body, "page=2") {
		t.Fatalf("next link missing")
	}

	req = httptest.NewRequest(http.MethodGet, "/?per=2&page=2", nil)
	rr = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body = rr.Body.String()
	if !strings.Contains(body, "Page 2 / 2") {
		t.Fatalf("page 2 indicator: body=\n%s", body)
	}
}

func TestIndexNoMatch(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Store.Close()
	req := httptest.NewRequest(http.MethodGet, "/?white=zzz", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	if !strings.Contains(body, "No games match") {
		t.Fatalf("expected empty-state message")
	}
	if !strings.Contains(body, "clearing filters") {
		t.Fatalf("expected clear-filters link")
	}
}

func TestGamePageRenders(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Store.Close()
	req := httptest.NewRequest(http.MethodGet, "/game/1/ply/2", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "class=\"sq") {
		t.Fatalf("board cells missing")
	}
	if !strings.Contains(body, "ply-indicator") {
		t.Fatalf("ply indicator missing")
	}
}

func TestPieceSVGsServed(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Store.Close()
	pieces := []string{"wK", "wQ", "wR", "wB", "wN", "wP", "bK", "bQ", "bR", "bB", "bN", "bP"}
	for _, p := range pieces {
		req := httptest.NewRequest(http.MethodGet, "/static/pieces/"+p+".svg", nil)
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("%s: status %d", p, rr.Code)
			continue
		}
		body := rr.Body.String()
		if !strings.Contains(body, "<svg") {
			t.Errorf("%s: body missing <svg>: %q", p, body[:min(80, len(body))])
		}
	}
}

func TestBoardTemplateUsesImg(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Store.Close()
	req := httptest.NewRequest(http.MethodGet, "/game/1/fragment/0", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	// Starting position has all the pieces we expect.
	for _, p := range []string{"wK", "wQ", "wR", "wB", "wN", "wP", "bK", "bQ", "bR", "bB", "bN", "bP"} {
		if !strings.Contains(body, "/static/pieces/"+p+".svg") {
			t.Errorf("fragment missing img for %s", p)
		}
	}
}

func TestPieceCodeMapping(t *testing.T) {
	cases := []struct {
		piece string
		want  string
	}{
		{"wK", "wK"}, {"wQ", "wQ"}, {"wR", "wR"},
		{"wB", "wB"}, {"wN", "wN"}, {"wP", "wP"},
		{"bK", "bK"}, {"bQ", "bQ"}, {"bR", "bR"},
		{"bB", "bB"}, {"bN", "bN"}, {"bP", "bP"},
	}
	// Parse the start position and collect codes we've seen.
	startFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	pos, err := parseFENForTest(t, startFEN)
	if err != nil {
		t.Fatal(err)
	}
	view := buildBoardView(pos, startFEN)
	seen := map[string]bool{}
	for _, row := range view.Rows {
		for _, cell := range row.Cells {
			if cell.Piece != "" {
				seen[cell.Piece] = true
			}
		}
	}
	for _, c := range cases {
		if !seen[c.want] {
			t.Errorf("start position missing code %s", c.want)
		}
	}
}

func TestGameFragment(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Store.Close()
	req := httptest.NewRequest(http.MethodGet, "/game/1/fragment/2", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "class=\"chessboard\"") {
		t.Fatalf("fragment missing board: %s", body)
	}
	if strings.Contains(body, "<html") {
		t.Fatalf("fragment should not be a full page")
	}
}
