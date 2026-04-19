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
		webCtx, webCancel := context.WithCancel(ctx)
		defer webCancel()
		go web.RunConsole(webCtx, os.Stdin, os.Stdout, webCancel)
		return srv.Serve(webCtx, addr)
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
	chunks := pgn.SplitChunks(string(raw), len(games))

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
