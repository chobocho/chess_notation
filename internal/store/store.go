// Package store is the persistence layer for games and bookmarks, backed by
// pure-Go SQLite (modernc.org/sqlite).
package store

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/chobocho/chess_notation/internal/chess"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Store wraps a sql.DB and exposes typed methods.
type Store struct {
	DB *sql.DB
}

// Open opens (or creates) the SQLite database at path and applies the schema.
// Pass ":memory:" for an ephemeral DB.
func Open(path string) (*Store, error) {
	dsn := path + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	if path == ":memory:" {
		dsn = ":memory:?_pragma=foreign_keys(1)"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // sqlite single-writer
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: apply schema: %w", err)
	}
	return &Store{DB: db}, nil
}

// Close closes the underlying DB.
func (s *Store) Close() error { return s.DB.Close() }

// GameMeta is a row in the games table without the raw PGN blob.
type GameMeta struct {
	ID         int64
	Event      string
	Site       string
	Date       string
	Round      string
	White      string
	Black      string
	Result     string
	WhiteElo   sql.NullInt64
	BlackElo   sql.NullInt64
	ECO        string
	Opening    string
	PlyCount   int
	ImportedAt time.Time
}

// Bookmark is a single bookmark row.
type Bookmark struct {
	ID        int64
	GameID    int64
	Ply       int
	Note      string
	CreatedAt time.Time
}

// ImportGame stores a parsed chess.Game together with its raw PGN text.
// Returns the new game ID.
func (s *Store) ImportGame(ctx context.Context, g *chess.Game, rawPGN string) (int64, error) {
	nodes := g.Mainline()
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	tag := func(k string) string { return g.Tags[k] }
	var wElo, bElo sql.NullInt64
	if v := tag("WhiteElo"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			wElo = sql.NullInt64{Int64: int64(n), Valid: true}
		}
	}
	if v := tag("BlackElo"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			bElo = sql.NullInt64{Int64: int64(n), Valid: true}
		}
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO games
		  (event, site, date, round, white, black, result,
		   white_elo, black_elo, eco, opening, pgn_raw, ply_count, imported_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tag("Event"), tag("Site"), tag("Date"), tag("Round"),
		tag("White"), tag("Black"), tag("Result"),
		wElo, bElo, tag("ECO"), tag("Opening"),
		rawPGN, len(nodes), time.Now().Unix(),
	)
	if err != nil {
		return 0, err
	}
	gameID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Ply 0 = starting position.
	startFEN := chess.StartFEN
	if g.Start != nil {
		startFEN = g.Start.FEN()
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO moves (game_id, ply, san, uci, fen_after) VALUES (?, 0, NULL, NULL, ?)`,
		gameID, startFEN); err != nil {
		return 0, err
	}

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO moves (game_id, ply, san, uci, fen_after) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	for _, n := range nodes {
		if _, err := stmt.ExecContext(ctx, gameID, n.Ply, n.SAN, n.UCI, n.FEN); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return gameID, nil
}

// ListFilter narrows results from ListGamesFiltered. Zero-valued fields are ignored.
// White and Black do substring (case-insensitive) matching; Result is exact.
type ListFilter struct {
	White  string
	Black  string
	Result string
}

// buildFilterSQL returns the WHERE clause (without the "WHERE") and the args.
func (f ListFilter) buildFilterSQL() (string, []any) {
	var clauses []string
	var args []any
	if f.White != "" {
		clauses = append(clauses, "LOWER(white) LIKE ?")
		args = append(args, "%"+strings.ToLower(f.White)+"%")
	}
	if f.Black != "" {
		clauses = append(clauses, "LOWER(black) LIKE ?")
		args = append(args, "%"+strings.ToLower(f.Black)+"%")
	}
	if f.Result != "" {
		clauses = append(clauses, "result = ?")
		args = append(args, f.Result)
	}
	return strings.Join(clauses, " AND "), args
}

// ListGames returns games ordered by ID descending, paginated.
func (s *Store) ListGames(ctx context.Context, limit, offset int) ([]GameMeta, error) {
	return s.ListGamesFiltered(ctx, ListFilter{}, limit, offset)
}

// ListGamesFiltered returns filtered, paginated games ordered by ID descending.
func (s *Store) ListGamesFiltered(ctx context.Context, f ListFilter, limit, offset int) ([]GameMeta, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT id, event, site, date, round, white, black, result,
	             white_elo, black_elo, eco, opening, ply_count, imported_at
	        FROM games`
	where, args := f.buildFilterSQL()
	if where != "" {
		q += " WHERE " + where
	}
	q += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GameMeta
	for rows.Next() {
		var m GameMeta
		var ts int64
		if err := rows.Scan(
			&m.ID, &m.Event, &m.Site, &m.Date, &m.Round,
			&m.White, &m.Black, &m.Result,
			&m.WhiteElo, &m.BlackElo, &m.ECO, &m.Opening,
			&m.PlyCount, &ts,
		); err != nil {
			return nil, err
		}
		m.ImportedAt = time.Unix(ts, 0)
		out = append(out, m)
	}
	return out, rows.Err()
}

// CountGames returns the total row count matching f.
func (s *Store) CountGames(ctx context.Context, f ListFilter) (int, error) {
	q := `SELECT COUNT(*) FROM games`
	where, args := f.buildFilterSQL()
	if where != "" {
		q += " WHERE " + where
	}
	var n int
	if err := s.DB.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// GetGame fetches metadata for a single game.
func (s *Store) GetGame(ctx context.Context, id int64) (*GameMeta, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, event, site, date, round, white, black, result,
		       white_elo, black_elo, eco, opening, ply_count, imported_at
		  FROM games WHERE id = ?`, id)
	var m GameMeta
	var ts int64
	err := row.Scan(
		&m.ID, &m.Event, &m.Site, &m.Date, &m.Round,
		&m.White, &m.Black, &m.Result,
		&m.WhiteElo, &m.BlackElo, &m.ECO, &m.Opening,
		&m.PlyCount, &ts,
	)
	if err == sql.ErrNoRows {
		return nil, chess.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	m.ImportedAt = time.Unix(ts, 0)
	return &m, nil
}

// GetRawPGN returns the stored raw PGN for a game.
func (s *Store) GetRawPGN(ctx context.Context, id int64) (string, error) {
	var s1 string
	err := s.DB.QueryRowContext(ctx, `SELECT pgn_raw FROM games WHERE id = ?`, id).Scan(&s1)
	if err == sql.ErrNoRows {
		return "", chess.ErrNotFound
	}
	return s1, err
}

// MoveRow is a (ply, san, fen_after) tuple from the moves table.
type MoveRow struct {
	Ply      int
	SAN      sql.NullString
	UCI      sql.NullString
	FENAfter string
}

// GetMoves returns all move rows for a game ordered by ply.
func (s *Store) GetMoves(ctx context.Context, gameID int64) ([]MoveRow, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT ply, san, uci, fen_after FROM moves WHERE game_id = ? ORDER BY ply`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MoveRow
	for rows.Next() {
		var m MoveRow
		if err := rows.Scan(&m.Ply, &m.SAN, &m.UCI, &m.FENAfter); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// GetFENAt returns the FEN stored at (gameID, ply).
func (s *Store) GetFENAt(ctx context.Context, gameID int64, ply int) (string, error) {
	var fen string
	err := s.DB.QueryRowContext(ctx,
		`SELECT fen_after FROM moves WHERE game_id = ? AND ply = ?`, gameID, ply).Scan(&fen)
	if err == sql.ErrNoRows {
		return "", chess.ErrNotFound
	}
	return fen, err
}

// AddBookmark inserts a new bookmark and returns its ID.
func (s *Store) AddBookmark(ctx context.Context, gameID int64, ply int, note string) (int64, error) {
	res, err := s.DB.ExecContext(ctx,
		`INSERT INTO bookmarks (game_id, ply, note, created_at) VALUES (?, ?, ?, ?)`,
		gameID, ply, note, time.Now().Unix())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListBookmarks returns bookmarks for a game (most recent first).
func (s *Store) ListBookmarks(ctx context.Context, gameID int64) ([]Bookmark, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, game_id, ply, note, created_at FROM bookmarks
		 WHERE game_id = ? ORDER BY id DESC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Bookmark
	for rows.Next() {
		var b Bookmark
		var ts int64
		if err := rows.Scan(&b.ID, &b.GameID, &b.Ply, &b.Note, &ts); err != nil {
			return nil, err
		}
		b.CreatedAt = time.Unix(ts, 0)
		out = append(out, b)
	}
	return out, rows.Err()
}

// DeleteBookmark removes a bookmark by ID.
func (s *Store) DeleteBookmark(ctx context.Context, id int64) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM bookmarks WHERE id = ?`, id)
	return err
}

// DeleteGame removes a game (and cascading moves/bookmarks).
func (s *Store) DeleteGame(ctx context.Context, id int64) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM games WHERE id = ?`, id)
	return err
}
