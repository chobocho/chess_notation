# chess_notation

A small chess PGN viewer written in Go. Stores a game library in SQLite, steps
through moves in the terminal with an ANSI board, or serves a minimal web UI.

Originally a half-finished C# console app (preserved under `legacy-csharp/`);
this repository is the Go rewrite.

## Features

- Full PGN parsing: tag pairs, SAN movetext, comments (`{ ... }` and `;`), NAGs
  (`$N`), recursive variations (`( ... )`), multi-game files.
- Full chess engine: every piece (incl. castling, en passant, promotion,
  disambiguation like `Nbd2`/`R1e2`), check/mate detection, legal move
  filtering. Validated with perft to depth 4 from the starting position and
  depth 3 on standard test positions (Kiwipete, etc.).
- SQLite-backed library (pure Go, no CGO) with a normalized `moves` table for
  fast ply seek and a `bookmarks` table.
- Terminal viewer: ANSI-colored board, raw-mode key input with graceful
  fallback to line-buffered stdin.
- Web viewer: server-rendered board + move list, keyboard arrow-key
  navigation, bookmark form.
- Single static binary. Cross-compiles for Android/arm64 (termux) without
  CGO.

## Build

```sh
go build -o chess_notation .
```

Cross-build for termux:

```sh
GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build -o chess_notation-termux .
```

## Usage

```sh
# Import a PGN file into the default database.
chess_notation --import game.pgn

# Open the interactive CLI viewer on the stored library.
chess_notation

# View a single PGN file without touching the database.
chess_notation --pgn game.pgn

# Serve the web UI on http://localhost:8080.
chess_notation --web --port 8080

# Use an alternate database path.
chess_notation --db /path/to/games.db
```

The default database lives at `$XDG_DATA_HOME/chess_notation/games.db`
(falling back to `~/.local/share/chess_notation/games.db`).

### CLI keys

| Key | Action |
| --- | --- |
| `n`, `→`, space | next ply |
| `b`, `←` | previous ply |
| `g` | jump to ply (prompts for a number) |
| `f` | print current FEN |
| `m` | bookmark current ply (prompts for a note) |
| `l` | list bookmarks |
| `q`, Esc | quit |

### Web UI

`/` — game list. `/game/{id}` — viewer; `/game/{id}/ply/{n}` — direct link to
a ply. Arrow keys (or `j`/`k`) step forward and backward without reloading.

## Layout

```
main.go                       flag routing
internal/chess/               board, move generation, SAN, game tree
internal/pgn/                 PGN parser
internal/store/               SQLite schema + CRUD (modernc.org/sqlite)
internal/cli/                 ANSI board renderer + key loop
internal/web/                 net/http server + embedded templates/static
testdata/                     sample PGN used by tests and docs
legacy-csharp/                original C# prototype (reference only)
```

## Tests

```sh
go test ./...
go vet ./...
```

Perft and SAN roundtrip live in `internal/chess`; PGN parsing in
`internal/pgn`; store CRUD in `internal/store`.
