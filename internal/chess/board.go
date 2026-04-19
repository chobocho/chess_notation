package chess

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Square is 0..63, A1=0, B1=1, ..., H1=7, A2=8, ..., H8=63.
type Square int8

const NoSquare Square = -1

func Sq(file, rank int) Square { return Square(rank*8 + file) }
func (s Square) File() int     { return int(s) & 7 }
func (s Square) Rank() int     { return int(s) >> 3 }
func (s Square) Valid() bool   { return s >= 0 && s < 64 }

// String returns algebraic like "e4".
func (s Square) String() string {
	if !s.Valid() {
		return "-"
	}
	return string([]byte{byte('a' + s.File()), byte('1' + s.Rank())})
}

func ParseSquare(s string) (Square, error) {
	if len(s) != 2 {
		return NoSquare, fmt.Errorf("bad square %q", s)
	}
	f := int(s[0] - 'a')
	r := int(s[1] - '1')
	if f < 0 || f > 7 || r < 0 || r > 7 {
		return NoSquare, fmt.Errorf("bad square %q", s)
	}
	return Sq(f, r), nil
}

// Castling rights bitmask.
type Castling uint8

const (
	WhiteKingside Castling = 1 << iota
	WhiteQueenside
	BlackKingside
	BlackQueenside
)

// Position is a full chess position.
type Position struct {
	Board          [64]Piece
	SideToMove     Color
	Castling       Castling
	EPSquare       Square // NoSquare if none
	HalfmoveClock  int
	FullmoveNumber int
}

const StartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func NewPosition() *Position {
	p, err := ParseFEN(StartFEN)
	if err != nil {
		panic(err)
	}
	return p
}

// ParseFEN parses a full FEN string.
func ParseFEN(fen string) (*Position, error) {
	fields := strings.Fields(fen)
	if len(fields) < 4 {
		return nil, fmt.Errorf("fen: need at least 4 fields, got %d", len(fields))
	}
	pos := &Position{EPSquare: NoSquare, FullmoveNumber: 1}

	// Piece placement.
	ranks := strings.Split(fields[0], "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("fen: need 8 ranks, got %d", len(ranks))
	}
	for i, rankStr := range ranks {
		rank := 7 - i
		file := 0
		for j := 0; j < len(rankStr); j++ {
			c := rankStr[j]
			if c >= '1' && c <= '8' {
				file += int(c - '0')
				continue
			}
			p := PieceFromFEN(c)
			if p == Empty {
				return nil, fmt.Errorf("fen: bad piece %q", c)
			}
			if file > 7 {
				return nil, fmt.Errorf("fen: rank overflow")
			}
			pos.Board[Sq(file, rank)] = p
			file++
		}
		if file != 8 {
			return nil, fmt.Errorf("fen: rank %d has %d files", rank, file)
		}
	}

	// Side to move.
	switch fields[1] {
	case "w":
		pos.SideToMove = White
	case "b":
		pos.SideToMove = Black
	default:
		return nil, fmt.Errorf("fen: bad side %q", fields[1])
	}

	// Castling.
	if fields[2] != "-" {
		for i := 0; i < len(fields[2]); i++ {
			switch fields[2][i] {
			case 'K':
				pos.Castling |= WhiteKingside
			case 'Q':
				pos.Castling |= WhiteQueenside
			case 'k':
				pos.Castling |= BlackKingside
			case 'q':
				pos.Castling |= BlackQueenside
			default:
				return nil, fmt.Errorf("fen: bad castling %q", fields[2])
			}
		}
	}

	// En passant.
	if fields[3] != "-" {
		ep, err := ParseSquare(fields[3])
		if err != nil {
			return nil, fmt.Errorf("fen: %w", err)
		}
		pos.EPSquare = ep
	}

	if len(fields) >= 5 {
		n, err := strconv.Atoi(fields[4])
		if err != nil {
			return nil, fmt.Errorf("fen: halfmove: %w", err)
		}
		pos.HalfmoveClock = n
	}
	if len(fields) >= 6 {
		n, err := strconv.Atoi(fields[5])
		if err != nil {
			return nil, fmt.Errorf("fen: fullmove: %w", err)
		}
		pos.FullmoveNumber = n
	}

	return pos, nil
}

// FEN renders the position.
func (p *Position) FEN() string {
	var sb strings.Builder
	for rank := 7; rank >= 0; rank-- {
		empty := 0
		for file := 0; file < 8; file++ {
			pc := p.Board[Sq(file, rank)]
			if pc == Empty {
				empty++
				continue
			}
			if empty > 0 {
				sb.WriteByte(byte('0' + empty))
				empty = 0
			}
			sb.WriteByte(pc.FEN())
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
		}
		if rank > 0 {
			sb.WriteByte('/')
		}
	}
	sb.WriteByte(' ')
	if p.SideToMove == White {
		sb.WriteByte('w')
	} else {
		sb.WriteByte('b')
	}
	sb.WriteByte(' ')
	if p.Castling == 0 {
		sb.WriteByte('-')
	} else {
		if p.Castling&WhiteKingside != 0 {
			sb.WriteByte('K')
		}
		if p.Castling&WhiteQueenside != 0 {
			sb.WriteByte('Q')
		}
		if p.Castling&BlackKingside != 0 {
			sb.WriteByte('k')
		}
		if p.Castling&BlackQueenside != 0 {
			sb.WriteByte('q')
		}
	}
	sb.WriteByte(' ')
	if p.EPSquare == NoSquare {
		sb.WriteByte('-')
	} else {
		sb.WriteString(p.EPSquare.String())
	}
	fmt.Fprintf(&sb, " %d %d", p.HalfmoveClock, p.FullmoveNumber)
	return sb.String()
}

// Clone returns a deep copy.
func (p *Position) Clone() *Position {
	q := *p
	return &q
}

// KingSquare returns the square of the given color's king, or NoSquare.
func (p *Position) KingSquare(c Color) Square {
	target := MakePiece(c, King)
	for s := Square(0); s < 64; s++ {
		if p.Board[s] == target {
			return s
		}
	}
	return NoSquare
}

// ErrNotFound is returned when a search yields nothing.
var ErrNotFound = errors.New("not found")
