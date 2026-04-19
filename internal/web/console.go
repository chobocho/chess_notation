package web

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

// WebCmd is the kind of command entered on the admin stdin console.
type WebCmd int

const (
	CmdUnknown WebCmd = iota
	CmdNoop           // empty line
	CmdHelp
	CmdExit
)

// ParseWebCommand maps a raw line to a WebCmd. Case-insensitive.
func ParseWebCommand(line string) WebCmd {
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "":
		return CmdNoop
	case "help", "?", "h":
		return CmdHelp
	case "exit", "quit", "q":
		return CmdExit
	}
	return CmdUnknown
}

// ConsoleHelp is the help text printed at startup and on "help".
const ConsoleHelp = `chess_notation web console commands:
  help, ?, h    show this message
  exit, quit, q shut down the server`

// RunConsole reads commands from r and writes prompts/responses to w.
// When the user types "exit", onExit is called and the loop returns.
// The loop also returns when r reaches EOF or ctx is cancelled.
func RunConsole(ctx context.Context, r io.Reader, w io.Writer, onExit func()) {
	fmt.Fprintln(w, ConsoleHelp)
	fmt.Fprint(w, "> ")

	lines := make(chan string, 1)
	go func() {
		defer close(lines)
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			lines <- scan.Text()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-lines:
			if !ok {
				return
			}
			switch ParseWebCommand(line) {
			case CmdExit:
				fmt.Fprintln(w, "shutting down...")
				if onExit != nil {
					onExit()
				}
				return
			case CmdHelp:
				fmt.Fprintln(w, ConsoleHelp)
			case CmdNoop:
				// Just re-prompt.
			default:
				fmt.Fprintf(w, "unknown command %q; type 'help' for options\n", strings.TrimSpace(line))
			}
			fmt.Fprint(w, "> ")
		}
	}
}
