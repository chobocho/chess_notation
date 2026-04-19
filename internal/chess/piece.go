package chess

type Color uint8

const (
	White Color = 0
	Black Color = 1
)

func (c Color) Opp() Color { return c ^ 1 }

type PieceType uint8

const (
	NoPiece PieceType = 0
	Pawn   PieceType = 1
	Knight PieceType = 2
	Bishop PieceType = 3
	Rook   PieceType = 4
	Queen  PieceType = 5
	King   PieceType = 6
)

type Piece uint8

const Empty Piece = 0

func MakePiece(c Color, t PieceType) Piece {
	if t == NoPiece {
		return Empty
	}
	return Piece(uint8(t) | (uint8(c) << 3))
}

func (p Piece) Type() PieceType { return PieceType(p & 7) }
func (p Piece) Color() Color    { return Color((p >> 3) & 1) }
func (p Piece) IsEmpty() bool   { return p == Empty }

// FEN letter for the piece: uppercase white, lowercase black.
func (p Piece) FEN() byte {
	if p == Empty {
		return '.'
	}
	var ch byte
	switch p.Type() {
	case Pawn:
		ch = 'P'
	case Knight:
		ch = 'N'
	case Bishop:
		ch = 'B'
	case Rook:
		ch = 'R'
	case Queen:
		ch = 'Q'
	case King:
		ch = 'K'
	default:
		return '.'
	}
	if p.Color() == Black {
		ch += 'a' - 'A'
	}
	return ch
}

func PieceFromFEN(b byte) Piece {
	var c Color
	if b >= 'a' && b <= 'z' {
		c = Black
		b -= 'a' - 'A'
	} else {
		c = White
	}
	var t PieceType
	switch b {
	case 'P':
		t = Pawn
	case 'N':
		t = Knight
	case 'B':
		t = Bishop
	case 'R':
		t = Rook
	case 'Q':
		t = Queen
	case 'K':
		t = King
	default:
		return Empty
	}
	return MakePiece(c, t)
}

// PieceTypeLetter returns the SAN letter for a non-pawn piece type.
func PieceTypeLetter(t PieceType) byte {
	switch t {
	case Knight:
		return 'N'
	case Bishop:
		return 'B'
	case Rook:
		return 'R'
	case Queen:
		return 'Q'
	case King:
		return 'K'
	}
	return 0
}

// PieceTypeFromLetter parses a SAN piece letter (N/B/R/Q/K). Returns NoPiece on unknown.
func PieceTypeFromLetter(b byte) PieceType {
	switch b {
	case 'N':
		return Knight
	case 'B':
		return Bishop
	case 'R':
		return Rook
	case 'Q':
		return Queen
	case 'K':
		return King
	}
	return NoPiece
}

// Unicode glyphs for board rendering.
func (p Piece) Glyph() rune {
	if p == Empty {
		return ' '
	}
	switch p {
	case MakePiece(White, King):
		return '\u2654'
	case MakePiece(White, Queen):
		return '\u2655'
	case MakePiece(White, Rook):
		return '\u2656'
	case MakePiece(White, Bishop):
		return '\u2657'
	case MakePiece(White, Knight):
		return '\u2658'
	case MakePiece(White, Pawn):
		return '\u2659'
	case MakePiece(Black, King):
		return '\u265A'
	case MakePiece(Black, Queen):
		return '\u265B'
	case MakePiece(Black, Rook):
		return '\u265C'
	case MakePiece(Black, Bishop):
		return '\u265D'
	case MakePiece(Black, Knight):
		return '\u265E'
	case MakePiece(Black, Pawn):
		return '\u265F'
	}
	return '?'
}
