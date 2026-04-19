// Package cli renders boards and drives the terminal viewer.
package cli

import (
	"fmt"
	"strings"

	"github.com/chobocho/chess_notation/internal/chess"
)

const (
	ansiReset    = "\x1b[0m"
	ansiDarkSq   = "\x1b[48;5;94m"  // brown
	ansiLightSq  = "\x1b[48;5;180m" // tan
	ansiWhiteFg  = "\x1b[38;5;231m"
	ansiBlackFg  = "\x1b[38;5;16m"
)

// RenderBoard returns an ANSI-colored board string.
// Files are labeled a..h, ranks 1..8, viewed from white's perspective.
func RenderBoard(pos *chess.Position) string {
	var sb strings.Builder
	sb.WriteString("   a  b  c  d  e  f  g  h\n")
	for rank := 7; rank >= 0; rank-- {
		fmt.Fprintf(&sb, "%d ", rank+1)
		for file := 0; file < 8; file++ {
			sq := chess.Sq(file, rank)
			dark := (file+rank)%2 == 0
			if dark {
				sb.WriteString(ansiDarkSq)
			} else {
				sb.WriteString(ansiLightSq)
			}
			p := pos.Board[sq]
			if p == chess.Empty {
				sb.WriteString("   ")
			} else {
				if p.Color() == chess.White {
					sb.WriteString(ansiWhiteFg)
				} else {
					sb.WriteString(ansiBlackFg)
				}
				fmt.Fprintf(&sb, " %c ", p.Glyph())
			}
			sb.WriteString(ansiReset)
		}
		fmt.Fprintf(&sb, " %d\n", rank+1)
	}
	sb.WriteString("   a  b  c  d  e  f  g  h\n")
	return sb.String()
}
