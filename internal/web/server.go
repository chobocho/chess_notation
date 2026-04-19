// Package web serves the SQLite game library over HTTP with server-side-rendered boards.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chobocho/chess_notation/internal/chess"
	"github.com/chobocho/chess_notation/internal/pgn"
	"github.com/chobocho/chess_notation/internal/store"
)

// maxImportBytes caps both uploaded files and the textarea paste path.
const maxImportBytes = 10 * 1024 * 1024 // 10 MiB

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

// Server wraps the store and compiled templates.
type Server struct {
	Store    *store.Store
	indexT   *template.Template
	gameT    *template.Template
	boardT   *template.Template
	importT  *template.Template
}

// NewServer compiles templates and returns a ready Server.
func NewServer(s *store.Store) (*Server, error) {
	idx, err := template.ParseFS(templatesFS, "templates/layout.html", "templates/index.html")
	if err != nil {
		return nil, fmt.Errorf("web: parse index templates: %w", err)
	}
	game, err := template.ParseFS(templatesFS, "templates/layout.html", "templates/game.html", "templates/board.html")
	if err != nil {
		return nil, fmt.Errorf("web: parse game templates: %w", err)
	}
	board, err := template.ParseFS(templatesFS, "templates/board.html")
	if err != nil {
		return nil, fmt.Errorf("web: parse board template: %w", err)
	}
	imp, err := template.ParseFS(templatesFS, "templates/layout.html", "templates/import.html")
	if err != nil {
		return nil, fmt.Errorf("web: parse import templates: %w", err)
	}
	return &Server{Store: s, indexT: idx, gameT: game, boardT: board, importT: imp}, nil
}

// Handler returns the ServeMux for the web UI.
func (srv *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleIndex)
	mux.HandleFunc("/import", srv.handleImport)
	mux.HandleFunc("/game/", srv.handleGameRoutes)
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))
	return mux
}

// defaultPerPage is the default number of rows on /.
const defaultPerPage = 50

func (srv *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	q := r.URL.Query()
	filter := store.ListFilter{
		White:  strings.TrimSpace(q.Get("white")),
		Black:  strings.TrimSpace(q.Get("black")),
		Result: strings.TrimSpace(q.Get("result")),
	}
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	per, _ := strconv.Atoi(q.Get("per"))
	if per <= 0 || per > 200 {
		per = defaultPerPage
	}
	offset := (page - 1) * per

	total, err := srv.Store.CountGames(r.Context(), filter)
	if err != nil {
		httpErr(w, err)
		return
	}
	games, err := srv.Store.ListGamesFiltered(r.Context(), filter, per, offset)
	if err != nil {
		httpErr(w, err)
		return
	}

	totalPages := (total + per - 1) / per
	if totalPages == 0 {
		totalPages = 1
	}

	importedN := 0
	if v := q.Get("imported"); v != "" {
		importedN, _ = strconv.Atoi(v)
	}

	gamesJ := make([]gameRowJSON, len(games))
	for i, g := range games {
		gamesJ[i] = gameRowJSON{ID: g.ID, White: g.White, Black: g.Black, Event: g.Event, Date: g.Date, Result: g.Result, PlyCount: g.PlyCount}
	}
	pd := indexPageJSON{
		Page: "index", Imported: importedN, Total: total, Games: gamesJ,
		Filter:     filterJSON{White: filter.White, Black: filter.Black, Result: filter.Result},
		PageNum:    page, TotalPages: totalPages,
		HasPrev:    page > 1, HasNext: page < totalPages,
		PrevPage:   page - 1, NextPage: page + 1, PerPage: per,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := srv.indexT.ExecuteTemplate(w, "layout", map[string]any{"Title": "Games", "PageDataJSON": pageJSON(pd)}); err != nil {
		log.Printf("index template: %v", err)
	}
}

// indexQueryBase returns a query string (no leading "?") encoding the filter and per-page,
// ready to have "&page=N" appended.
func indexQueryBase(f store.ListFilter, per int) string {
	v := make([]string, 0, 4)
	add := func(k, s string) {
		if s == "" {
			return
		}
		v = append(v, k+"="+urlEscape(s))
	}
	add("white", f.White)
	add("black", f.Black)
	add("result", f.Result)
	if per != defaultPerPage {
		v = append(v, "per="+strconv.Itoa(per))
	}
	return strings.Join(v, "&")
}

// urlEscape is a tiny replacement for url.QueryEscape so we don't pull net/url here just for it.
func urlEscape(s string) string { return queryEscape(s) }

func queryEscape(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'),
			c == '-', c == '_', c == '.', c == '~':
			b.WriteByte(c)
		case c == ' ':
			b.WriteByte('+')
		default:
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0xF])
		}
	}
	return b.String()
}

// /game/{id}
// /game/{id}/ply/{n}
// /game/{id}/fragment/{n}
// POST /game/{id}/bookmark
func (srv *Server) handleGameRoutes(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/game/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 {
		srv.renderGame(w, r, id, 0, false)
		return
	}
	switch parts[1] {
	case "ply":
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		n, err := strconv.Atoi(parts[2])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		srv.renderGame(w, r, id, n, false)
	case "fragment":
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		n, err := strconv.Atoi(parts[2])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		srv.renderGame(w, r, id, n, true)
	case "bookmark":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ply, _ := strconv.Atoi(r.FormValue("ply"))
		note := r.FormValue("note")
		if _, err := srv.Store.AddBookmark(r.Context(), id, ply, note); err != nil {
			httpErr(w, err)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/game/%d/ply/%d", id, ply), http.StatusSeeOther)
	default:
		http.NotFound(w, r)
	}
}

func (srv *Server) renderGame(w http.ResponseWriter, r *http.Request, id int64, ply int, fragmentOnly bool) {
	ctx := r.Context()
	meta, err := srv.Store.GetGame(ctx, id)
	if err != nil {
		if errors.Is(err, chess.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		httpErr(w, err)
		return
	}
	moves, err := srv.Store.GetMoves(ctx, id)
	if err != nil {
		httpErr(w, err)
		return
	}
	max := 0
	for _, m := range moves {
		if m.Ply > max {
			max = m.Ply
		}
	}
	if ply < 0 {
		ply = 0
	}
	if ply > max {
		ply = max
	}

	var fen string
	for _, m := range moves {
		if m.Ply == ply {
			fen = m.FENAfter
			break
		}
	}
	if _, err := chess.ParseFEN(fen); err != nil {
		httpErr(w, err)
		return
	}

	if fragmentOnly {
		// Clients redraw the canvas from FEN; return plain text so the request is cheap.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(fen))
		return
	}

	bms, _ := srv.Store.ListBookmarks(ctx, id)

	movesJ := make([]moveJSON, 0, max)
	for _, m := range moves {
		if m.Ply == 0 {
			continue
		}
		san := ""
		if m.SAN.Valid {
			san = m.SAN.String
		}
		movesJ = append(movesJ, moveJSON{Ply: m.Ply, Number: (m.Ply + 1) / 2, SAN: san})
	}
	bmsJ := make([]bmJSON, len(bms))
	for i, bm := range bms {
		bmsJ[i] = bmJSON{Ply: bm.Ply, Note: bm.Note}
	}
	pd := gamePageJSON{
		Page: "game", GameID: id,
		White: meta.White, Black: meta.Black, Event: meta.Event, Site: meta.Site, Date: meta.Date, Result: meta.Result,
		Ply: ply, MaxPly: max, FEN: fen,
		Moves: movesJ, Bookmarks: bmsJ,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := srv.gameT.ExecuteTemplate(w, "layout", map[string]any{
		"Title":        fmt.Sprintf("%s vs %s", meta.White, meta.Black),
		"PageDataJSON": pageJSON(pd),
		"GameID":       id,
		"Ply":          ply,
	}); err != nil {
		log.Printf("game template: %v", err)
	}
}

func max0(x int) int {
	if x < 0 {
		return 0
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func httpErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// importView drives the /import template.
type importView struct {
	Title    string
	Error    string
	Message  string
	Text     string
	Imported int
}

// handleImport serves the upload + paste form and processes submissions.
func (srv *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		srv.renderImport(w, importView{Title: "Import PGN"})
	case http.MethodPost:
		srv.processImport(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (srv *Server) renderImport(w http.ResponseWriter, v importView) {
	pd := importPageJSON{Page: "import", Error: v.Error, Text: v.Text}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := srv.importT.ExecuteTemplate(w, "layout", map[string]any{"Title": "Import PGN", "PageDataJSON": pageJSON(pd)}); err != nil {
		log.Printf("import template: %v", err)
	}
}

func (srv *Server) processImport(w http.ResponseWriter, r *http.Request) {
	// Enforce an overall body cap covering both the file upload and the textarea.
	r.Body = http.MaxBytesReader(w, r.Body, maxImportBytes)
	if err := r.ParseMultipartForm(maxImportBytes); err != nil {
		// Fall back to urlencoded for plain-text posts without a file.
		if err := r.ParseForm(); err != nil {
			srv.renderImport(w, importView{Error: "could not parse form: " + err.Error()})
			return
		}
	}

	raw := ""
	// 1) File upload takes precedence.
	if f, _, err := r.FormFile("pgn"); err == nil {
		defer f.Close()
		b, err := io.ReadAll(io.LimitReader(f, maxImportBytes))
		if err != nil {
			srv.renderImport(w, importView{Error: "read upload: " + err.Error()})
			return
		}
		raw = string(b)
	}
	// 2) Textarea paste.
	pasted := r.FormValue("pgn_text")
	if raw == "" {
		raw = pasted
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		srv.renderImport(w, importView{
			Error: "please upload a .pgn file or paste PGN text",
			Text:  pasted,
		})
		return
	}

	games, err := pgn.ParseString(raw)
	if err != nil {
		srv.renderImport(w, importView{
			Error: "parse PGN: " + err.Error(),
			Text:  pasted,
		})
		return
	}
	if len(games) == 0 {
		srv.renderImport(w, importView{
			Error: "no games found in the PGN",
			Text:  pasted,
		})
		return
	}

	chunks := pgn.SplitChunks(raw, len(games))
	ctx := r.Context()
	var firstID int64
	for i, g := range games {
		chunk := chunks[i]
		if chunk == "" {
			chunk = raw
		}
		id, err := srv.Store.ImportGame(ctx, g, chunk)
		if err != nil {
			srv.renderImport(w, importView{
				Error:    fmt.Sprintf("import game %d: %v", i+1, err),
				Imported: i,
				Text:     pasted,
			})
			return
		}
		if i == 0 {
			firstID = id
		}
	}

	// If exactly one game was imported, jump straight to it.
	if len(games) == 1 && firstID > 0 {
		http.Redirect(w, r, fmt.Sprintf("/game/%d", firstID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/?imported="+strconv.Itoa(len(games)), http.StatusSeeOther)
}

// ── Canvas SPA: JSON page-data types ─────────────────────────────────────────

type indexPageJSON struct {
	Page       string        `json:"page"`
	Imported   int           `json:"imported"`
	Total      int           `json:"total"`
	Games      []gameRowJSON `json:"games"`
	Filter     filterJSON    `json:"filter"`
	PageNum    int           `json:"page_num"`
	TotalPages int           `json:"total_pages"`
	HasPrev    bool          `json:"has_prev"`
	HasNext    bool          `json:"has_next"`
	PrevPage   int           `json:"prev_page"`
	NextPage   int           `json:"next_page"`
	PerPage    int           `json:"per_page"`
}

type gameRowJSON struct {
	ID       int64  `json:"id"`
	White    string `json:"white"`
	Black    string `json:"black"`
	Event    string `json:"event"`
	Date     string `json:"date"`
	Result   string `json:"result"`
	PlyCount int    `json:"ply_count"`
}

type filterJSON struct {
	White  string `json:"white"`
	Black  string `json:"black"`
	Result string `json:"result"`
}

type gamePageJSON struct {
	Page      string     `json:"page"`
	GameID    int64      `json:"game_id"`
	White     string     `json:"white"`
	Black     string     `json:"black"`
	Event     string     `json:"event"`
	Site      string     `json:"site"`
	Date      string     `json:"date"`
	Result    string     `json:"result"`
	Ply       int        `json:"ply"`
	MaxPly    int        `json:"max_ply"`
	FEN       string     `json:"fen"`
	Moves     []moveJSON `json:"moves"`
	Bookmarks []bmJSON   `json:"bookmarks"`
}

type moveJSON struct {
	Ply    int    `json:"ply"`
	Number int    `json:"number"`
	SAN    string `json:"san"`
}

type bmJSON struct {
	Ply  int    `json:"ply"`
	Note string `json:"note"`
}

type importPageJSON struct {
	Page  string `json:"page"`
	Error string `json:"error"`
	Text  string `json:"text"`
}

// pageJSON marshals v to JSON safe for embedding in an HTML <script> tag.
func pageJSON(v any) template.JS {
	b, _ := json.Marshal(v)
	s := strings.ReplaceAll(string(b), "</", `<\/`)
	return template.JS(s)
}

// Serve runs the HTTP server until ctx is cancelled.
func (srv *Server) Serve(ctx context.Context, addr string) error {
	hs := &http.Server{Addr: addr, Handler: srv.Handler()}
	errCh := make(chan error, 1)
	go func() {
		log.Printf("chess_notation web: listening on http://%s", addr)
		errCh <- hs.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return hs.Shutdown(shutCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
