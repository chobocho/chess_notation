// chess_notation is a CLI and web viewer for PGN chess games.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/chobocho/chess_notation/internal/chess"
	"github.com/chobocho/chess_notation/internal/cli"
	"github.com/chobocho/chess_notation/internal/pgn"
	"github.com/chobocho/chess_notation/internal/store"
	"github.com/chobocho/chess_notation/internal/web"
)

func main() {
	var (
		webMode    = flag.Bool("web", false, "run the web UI server")
		port       = flag.Int("port", 8080, "port for --web")
		importPath = flag.String("import", "", "import a PGN file into the DB and exit")
		pgnPath    = flag.String("pgn", "", "view a single PGN file without touching the DB")
		dbPath     = flag.String("db", "", "SQLite path (default: $XDG_DATA_HOME/chess_notation/games.db)")
	)
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, *webMode, *port, *importPath, *pgnPath, *dbPath); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, webMode bool, port int, importPath, pgnPath, dbPath string) error {
	// --pgn mode: one-off viewer, no DB.
	if pgnPath != "" {
		game, err := loadSinglePGN(pgnPath)
		if err != nil {
			return err
		}
		if webMode {
			return errors.New("--pgn is incompatible with --web; import the file first")
		}
		return cli.Run(ctx, nil, game, 0)
	}

	resolved, err := resolveDBPath(dbPath)
	if err != nil {
		return err
	}
	s, err := store.Open(resolved)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer s.Close()

	if importPath != "" {
		return importFile(ctx, s, importPath)
	}

	if webMode {
		srv, err := web.NewServer(s)
		if err != nil {
			return err
		}
		addr := net.JoinHostPort("", strconv.Itoa(port))
		return srv.Serve(ctx, addr)
	}

	return cli.Run(ctx, s, nil, 0)
}

func resolveDBPath(dbPath string) (string, error) {
	if dbPath != "" {
		return dbPath, nil
	}
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(base, "chess_notation")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "games.db"), nil
}

func loadSinglePGN(path string) (*chess.Game, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	games, err := pgn.Parse(f)
	if err != nil {
		return nil, err
	}
	if len(games) == 0 {
		return nil, fmt.Errorf("no games in %s", path)
	}
	return games[0], nil
}

func importFile(ctx context.Context, s *store.Store, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	games, err := pgn.ParseString(string(raw))
	if err != nil {
		return fmt.Errorf("parse pgn: %w", err)
	}
	if len(games) == 0 {
		return fmt.Errorf("no games in %s", path)
	}

	// Re-split the raw text into per-game slices so each stored pgn_raw is self-contained.
	chunks := splitPGNChunks(string(raw), len(games))

	for i, g := range games {
		chunk := chunks[i]
		if chunk == "" {
			chunk = string(raw)
		}
		id, err := s.ImportGame(ctx, g, chunk)
		if err != nil {
			return fmt.Errorf("import game %d: %w", i+1, err)
		}
		fmt.Printf("imported game %d (%s vs %s) as id=%d\n", i+1, g.Tags["White"], g.Tags["Black"], id)
	}
	return nil
}

// splitPGNChunks attempts to split a multi-game PGN text by blank-line boundaries
// before [Event ... tag blocks. Falls back to a single chunk on failure.
func splitPGNChunks(text string, expected int) []string {
	const sep = "\n[Event "
	chunks := []string{}
	idx := 0
	for {
		next := -1
		if idx == 0 {
			// First game may start at 0 or after some leading whitespace.
			p := indexAt(text, idx, "[Event ")
			if p < 0 {
				break
			}
			idx = p
		}
		after := idx + 1
		p := indexAt(text, after, sep)
		if p < 0 {
			next = -1
		} else {
			next = p + 1 // skip the leading '\n'
		}
		if next < 0 {
			chunks = append(chunks, text[idx:])
			break
		}
		chunks = append(chunks, text[idx:next])
		idx = next
	}
	if len(chunks) != expected {
		return make([]string, expected)
	}
	return chunks
}

func indexAt(s string, from int, sub string) int {
	if from >= len(s) {
		return -1
	}
	i := indexOf(s[from:], sub)
	if i < 0 {
		return -1
	}
	return from + i
}

// indexOf is a tiny substring index to avoid importing strings in this small helper.
func indexOf(s, sub string) int {
	n, m := len(s), len(sub)
	if m == 0 {
		return 0
	}
	for i := 0; i+m <= n; i++ {
		if s[i:i+m] == sub {
			return i
		}
	}
	return -1
}
