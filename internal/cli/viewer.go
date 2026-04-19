package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/chobocho/chess_notation/internal/chess"
	"github.com/chobocho/chess_notation/internal/pgn"
	"github.com/chobocho/chess_notation/internal/store"
	"golang.org/x/term"
)

// Run launches the interactive CLI. If `game` is nil, it prompts the user to
// pick one from the store.
func Run(ctx context.Context, s *store.Store, game *chess.Game, gameID int64) error {
	if game == nil {
		picked, id, err := pickGame(ctx, s)
		if err != nil {
			return err
		}
		if picked == nil {
			return nil
		}
		game, gameID = picked, id
	}
	return viewGame(ctx, s, game, gameID)
}

func pickGame(ctx context.Context, s *store.Store) (*chess.Game, int64, error) {
	metas, err := s.ListGames(ctx, 50, 0)
	if err != nil {
		return nil, 0, err
	}
	if len(metas) == 0 {
		fmt.Println("No games in the database. Use --import <file.pgn> to add one.")
		return nil, 0, nil
	}
	fmt.Println("Select a game:")
	for i, m := range metas {
		fmt.Printf("  %2d) %s vs %s  (%s, %s) [%s]\n",
			i+1, m.White, m.Black, m.Event, m.Result, m.Date)
	}
	fmt.Print("Number (or q to quit): ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" || line == "q" || line == "Q" {
		return nil, 0, nil
	}
	n, err := strconv.Atoi(line)
	if err != nil || n < 1 || n > len(metas) {
		return nil, 0, fmt.Errorf("cli: bad selection %q", line)
	}
	meta := metas[n-1]
	raw, err := s.GetRawPGN(ctx, meta.ID)
	if err != nil {
		return nil, 0, err
	}
	games, err := pgn.ParseString(raw)
	if err != nil {
		return nil, 0, err
	}
	if len(games) == 0 {
		return nil, 0, fmt.Errorf("cli: empty PGN for game %d", meta.ID)
	}
	return games[0], meta.ID, nil
}

// viewGame runs the render + key loop. Falls back to line-buffered input when
// stdin is not a TTY.
func viewGame(ctx context.Context, s *store.Store, g *chess.Game, gameID int64) error {
	draw := func() {
		clearScreen()
		pos := g.PositionAt(g.Cur)
		title := fmt.Sprintf("%s vs %s", g.Tags["White"], g.Tags["Black"])
		if g.Tags["Event"] != "" {
			title += " — " + g.Tags["Event"]
		}
		fmt.Println(title)
		fmt.Println()
		fmt.Print(RenderBoard(pos))
		fmt.Println()
		total := len(g.Mainline())
		if g.Cur.SAN != "" {
			fmt.Printf("Ply %d/%d: %s\n", g.Cur.Ply, total, g.Cur.SAN)
		} else {
			fmt.Printf("Ply 0/%d (start)\n", total)
		}
		if g.Cur.Comment != "" {
			fmt.Printf("  // %s\n", g.Cur.Comment)
		}
		fmt.Println()
		fmt.Println("Keys: n=next  b=back  g=goto  m=bookmark  l=list  f=fen  q=quit")
	}

	fd := int(os.Stdin.Fd())
	raw := term.IsTerminal(fd)
	var restore func()
	if raw {
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			raw = false
		} else {
			restore = func() { _ = term.Restore(fd, oldState) }
			defer restore()
		}
	}

	for {
		draw()
		action, err := readAction(raw)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		switch action {
		case actionQuit:
			return nil
		case actionNext:
			g.Next()
		case actionBack:
			g.Back()
		case actionFEN:
			fmt.Printf("\r\nFEN: %s\r\n", g.Cur.FEN)
			waitEnter(raw)
		case actionGoto:
			n, err := promptInt(raw, "Goto ply: ")
			if err == nil {
				g.Goto(n)
			}
		case actionBookmark:
			if s == nil || gameID == 0 {
				break
			}
			note, _ := promptString(raw, "Note: ")
			if _, err := s.AddBookmark(ctx, gameID, g.Cur.Ply, note); err != nil {
				fmt.Printf("\r\nerror: %v\r\n", err)
				waitEnter(raw)
			}
		case actionListBookmarks:
			if s == nil || gameID == 0 {
				break
			}
			bms, err := s.ListBookmarks(ctx, gameID)
			if err != nil {
				fmt.Printf("\r\nerror: %v\r\n", err)
				waitEnter(raw)
				break
			}
			fmt.Print("\r\nBookmarks:\r\n")
			for _, b := range bms {
				fmt.Printf("  ply %3d — %s\r\n", b.Ply, b.Note)
			}
			waitEnter(raw)
		}
	}
}

type action int

const (
	actionNone action = iota
	actionNext
	actionBack
	actionQuit
	actionFEN
	actionGoto
	actionBookmark
	actionListBookmarks
)

func readAction(raw bool) (action, error) {
	if raw {
		buf := make([]byte, 4)
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return actionNone, err
		}
		if n == 0 {
			return actionNone, nil
		}
		b := buf[0]
		// Arrow keys come as "\x1b[C" (right) or "\x1b[D" (left).
		if b == 0x1b && n >= 3 && buf[1] == '[' {
			switch buf[2] {
			case 'C':
				return actionNext, nil
			case 'D':
				return actionBack, nil
			}
			return actionQuit, nil
		}
		return mapChar(b), nil
	}
	// Line-buffered fallback.
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return actionNone, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return actionNext, nil
	}
	return mapChar(line[0]), nil
}

func mapChar(b byte) action {
	switch b {
	case 'n', 'N', ' ':
		return actionNext
	case 'b', 'B':
		return actionBack
	case 'g', 'G':
		return actionGoto
	case 'm', 'M':
		return actionBookmark
	case 'l', 'L':
		return actionListBookmarks
	case 'f', 'F':
		return actionFEN
	case 'q', 'Q', 0x03, 0x04:
		return actionQuit
	}
	return actionNone
}

// promptInt leaves raw mode briefly to read a line.
func promptInt(raw bool, prompt string) (int, error) {
	s, err := promptString(raw, prompt)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, err
	}
	return n, nil
}

func promptString(raw bool, prompt string) (string, error) {
	fd := int(os.Stdin.Fd())
	var oldState *term.State
	if raw {
		var err error
		oldState, err = term.GetState(fd)
		if err == nil {
			_ = term.Restore(fd, oldState)
		}
	}
	defer func() {
		if raw && oldState != nil {
			_, _ = term.MakeRaw(fd)
		}
	}()
	fmt.Print("\r\n" + prompt)
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func waitEnter(raw bool) {
	if !raw {
		return
	}
	buf := make([]byte, 1)
	_, _ = os.Stdin.Read(buf)
}

func clearScreen() {
	fmt.Print("\x1b[2J\x1b[H")
}
