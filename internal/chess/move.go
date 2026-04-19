package chess

import "fmt"

// MoveFlag bits describing a move's nature.
type MoveFlag uint8

const (
	FlagCapture MoveFlag = 1 << iota
	FlagEnPassant
	FlagCastleKing
	FlagCastleQueen
	FlagPawnDouble
	FlagPromotion
)

// Move is a single chess move.
type Move struct {
	From, To  Square
	Promotion PieceType // NoPiece unless FlagPromotion set
	Flags     MoveFlag
}

func (m Move) IsCastle() bool     { return m.Flags&(FlagCastleKing|FlagCastleQueen) != 0 }
func (m Move) IsCapture() bool    { return m.Flags&FlagCapture != 0 }
func (m Move) IsEnPassant() bool  { return m.Flags&FlagEnPassant != 0 }
func (m Move) IsPromotion() bool  { return m.Flags&FlagPromotion != 0 }
func (m Move) IsPawnDouble() bool { return m.Flags&FlagPawnDouble != 0 }

// UCI returns the long-algebraic form, e.g. "e2e4", "e7e8q".
func (m Move) UCI() string {
	s := m.From.String() + m.To.String()
	if m.IsPromotion() {
		s += string([]byte{byte(PieceTypeLetter(m.Promotion) + ('a' - 'A'))})
	}
	return s
}

// ParseUCI parses "e2e4" / "e7e8q".
func ParseUCI(s string) (Move, error) {
	if len(s) < 4 || len(s) > 5 {
		return Move{}, fmt.Errorf("uci: bad length %q", s)
	}
	from, err := ParseSquare(s[0:2])
	if err != nil {
		return Move{}, err
	}
	to, err := ParseSquare(s[2:4])
	if err != nil {
		return Move{}, err
	}
	m := Move{From: from, To: to}
	if len(s) == 5 {
		t := PieceTypeFromLetter(s[4] - ('a' - 'A'))
		if t == NoPiece {
			return Move{}, fmt.Errorf("uci: bad promo %q", s)
		}
		m.Promotion = t
		m.Flags |= FlagPromotion
	}
	return m, nil
}

// Undo captures state needed to unmake a move.
type Undo struct {
	Move           Move
	Captured       Piece
	CapturedSquare Square // for en passant this is the captured pawn's square
	PrevCastling   Castling
	PrevEP         Square
	PrevHalfmove   int
	PrevFullmove   int
}

// Make applies m to pos and returns an Undo that can reverse it.
// Caller is responsible for ensuring m is legal in pos.
func (pos *Position) Make(m Move) Undo {
	u := Undo{
		Move:         m,
		PrevCastling: pos.Castling,
		PrevEP:       pos.EPSquare,
		PrevHalfmove: pos.HalfmoveClock,
		PrevFullmove: pos.FullmoveNumber,
	}
	mover := pos.Board[m.From]
	isPawn := mover.Type() == Pawn

	// Handle capture (including en passant).
	if m.IsEnPassant() {
		var capSq Square
		if pos.SideToMove == White {
			capSq = m.To - 8
		} else {
			capSq = m.To + 8
		}
		u.Captured = pos.Board[capSq]
		u.CapturedSquare = capSq
		pos.Board[capSq] = Empty
	} else if pos.Board[m.To] != Empty {
		u.Captured = pos.Board[m.To]
		u.CapturedSquare = m.To
	}

	// Move the piece.
	pos.Board[m.To] = mover
	pos.Board[m.From] = Empty

	// Promotion.
	if m.IsPromotion() {
		pos.Board[m.To] = MakePiece(pos.SideToMove, m.Promotion)
	}

	// Castling: move the rook.
	if m.Flags&FlagCastleKing != 0 {
		if pos.SideToMove == White {
			pos.Board[Sq(5, 0)] = pos.Board[Sq(7, 0)]
			pos.Board[Sq(7, 0)] = Empty
		} else {
			pos.Board[Sq(5, 7)] = pos.Board[Sq(7, 7)]
			pos.Board[Sq(7, 7)] = Empty
		}
	} else if m.Flags&FlagCastleQueen != 0 {
		if pos.SideToMove == White {
			pos.Board[Sq(3, 0)] = pos.Board[Sq(0, 0)]
			pos.Board[Sq(0, 0)] = Empty
		} else {
			pos.Board[Sq(3, 7)] = pos.Board[Sq(0, 7)]
			pos.Board[Sq(0, 7)] = Empty
		}
	}

	// Update castling rights.
	// Any move from/to a corner or the king squares clears the relevant rights.
	switch m.From {
	case Sq(4, 0):
		pos.Castling &^= WhiteKingside | WhiteQueenside
	case Sq(4, 7):
		pos.Castling &^= BlackKingside | BlackQueenside
	case Sq(0, 0):
		pos.Castling &^= WhiteQueenside
	case Sq(7, 0):
		pos.Castling &^= WhiteKingside
	case Sq(0, 7):
		pos.Castling &^= BlackQueenside
	case Sq(7, 7):
		pos.Castling &^= BlackKingside
	}
	switch m.To {
	case Sq(0, 0):
		pos.Castling &^= WhiteQueenside
	case Sq(7, 0):
		pos.Castling &^= WhiteKingside
	case Sq(0, 7):
		pos.Castling &^= BlackQueenside
	case Sq(7, 7):
		pos.Castling &^= BlackKingside
	}

	// En passant target square.
	if m.IsPawnDouble() {
		if pos.SideToMove == White {
			pos.EPSquare = m.From + 8
		} else {
			pos.EPSquare = m.From - 8
		}
	} else {
		pos.EPSquare = NoSquare
	}

	// Halfmove clock.
	if isPawn || u.Captured != Empty {
		pos.HalfmoveClock = 0
	} else {
		pos.HalfmoveClock++
	}

	// Fullmove number increments after black moves.
	if pos.SideToMove == Black {
		pos.FullmoveNumber++
	}

	pos.SideToMove = pos.SideToMove.Opp()
	return u
}

// Unmake reverses the given Undo.
func (pos *Position) Unmake(u Undo) {
	pos.SideToMove = pos.SideToMove.Opp()
	pos.Castling = u.PrevCastling
	pos.EPSquare = u.PrevEP
	pos.HalfmoveClock = u.PrevHalfmove
	pos.FullmoveNumber = u.PrevFullmove

	m := u.Move
	moved := pos.Board[m.To]
	// Undo promotion: restore pawn.
	if m.IsPromotion() {
		moved = MakePiece(pos.SideToMove, Pawn)
	}
	pos.Board[m.From] = moved
	pos.Board[m.To] = Empty

	// Restore captured piece.
	if u.Captured != Empty {
		pos.Board[u.CapturedSquare] = u.Captured
	}

	// Undo castling rook.
	if m.Flags&FlagCastleKing != 0 {
		if pos.SideToMove == White {
			pos.Board[Sq(7, 0)] = pos.Board[Sq(5, 0)]
			pos.Board[Sq(5, 0)] = Empty
		} else {
			pos.Board[Sq(7, 7)] = pos.Board[Sq(5, 7)]
			pos.Board[Sq(5, 7)] = Empty
		}
	} else if m.Flags&FlagCastleQueen != 0 {
		if pos.SideToMove == White {
			pos.Board[Sq(0, 0)] = pos.Board[Sq(3, 0)]
			pos.Board[Sq(3, 0)] = Empty
		} else {
			pos.Board[Sq(0, 7)] = pos.Board[Sq(3, 7)]
			pos.Board[Sq(3, 7)] = Empty
		}
	}
}
