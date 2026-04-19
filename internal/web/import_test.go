package web

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chobocho/chess_notation/internal/store"
)

const singlePGN = `[Event "Upload"]
[White "Up"]
[Black "Load"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 1-0
`

const multiPGN = `[Event "One"]
[White "A"]
[Black "B"]
[Result "1-0"]

1. e4 e5 1-0

[Event "Two"]
[White "C"]
[Black "D"]
[Result "0-1"]

1. d4 d5 0-1
`

func newEmptyServer(t *testing.T) *Server {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "empty.db"))
	if err != nil {
		t.Fatal(err)
	}
	srv, err := NewServer(s)
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func TestImportGETShowsForm(t *testing.T) {
	srv := newEmptyServer(t)
	defer srv.Store.Close()
	req := httptest.NewRequest(http.MethodGet, "/import", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{
		`name="pgn"`,
		`name="pgn_text"`,
		`enctype="multipart/form-data"`,
		`>Import<`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("form missing %q", want)
		}
	}
}

func TestImportPOSTTextSingleGame(t *testing.T) {
	srv := newEmptyServer(t)
	defer srv.Store.Close()

	form := url.Values{"pgn_text": {singlePGN}}
	req := httptest.NewRequest(http.MethodPost, "/import", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status %d, want 303; body:\n%s", rr.Code, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	if !strings.HasPrefix(loc, "/game/") {
		t.Fatalf("redirect = %q, want /game/...", loc)
	}

	n, err := srv.Store.CountGames(context.Background(), store.ListFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("DB count = %d, want 1", n)
	}
}

func TestImportPOSTTextMultipleGames(t *testing.T) {
	srv := newEmptyServer(t)
	defer srv.Store.Close()

	form := url.Values{"pgn_text": {multiPGN}}
	req := httptest.NewRequest(http.MethodPost, "/import", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status %d, body:\n%s", rr.Code, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	if loc != "/?imported=2" {
		t.Fatalf("redirect = %q, want /?imported=2", loc)
	}

	n, err := srv.Store.CountGames(context.Background(), store.ListFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("DB count = %d, want 2", n)
	}
}

func TestImportPOSTFile(t *testing.T) {
	srv := newEmptyServer(t)
	defer srv.Store.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("pgn", "sample.pgn")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte(singlePGN)); err != nil {
		t.Fatal(err)
	}
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/import", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status %d, body:\n%s", rr.Code, rr.Body.String())
	}
	n, err := srv.Store.CountGames(context.Background(), store.ListFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("DB count = %d, want 1", n)
	}
}

func TestImportPOSTEmptyShowsError(t *testing.T) {
	srv := newEmptyServer(t)
	defer srv.Store.Close()

	form := url.Values{"pgn_text": {""}}
	req := httptest.NewRequest(http.MethodPost, "/import", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "upload a .pgn file or paste") {
		t.Fatalf("missing empty-error message: %s", rr.Body.String())
	}
}

func TestImportPOSTInvalidPreservesText(t *testing.T) {
	srv := newEmptyServer(t)
	defer srv.Store.Close()

	bad := `[Event "Bad"]

1. e9 not-a-move`
	form := url.Values{"pgn_text": {bad}}
	req := httptest.NewRequest(http.MethodPost, "/import", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "parse PGN") && !strings.Contains(body, "move ") {
		t.Fatalf("missing parse error: %s", body)
	}
	// The original text should still be in the form so the user can fix it.
	if !strings.Contains(body, "not-a-move") {
		t.Fatalf("expected pasted text to survive re-render")
	}
}

func TestIndexShowsImportedBanner(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Store.Close()
	req := httptest.NewRequest(http.MethodGet, "/?imported=3", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if !strings.Contains(rr.Body.String(), "Imported 3 games") {
		t.Fatalf("banner missing. body:\n%s", rr.Body.String())
	}
}

func TestIndexLinksToImport(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Store.Close()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if !strings.Contains(rr.Body.String(), `href="/import"`) {
		t.Fatalf("index missing /import link")
	}
}

// Uses fmt to quiet the unused-import warnings in some environments.
var _ = fmt.Sprintf
