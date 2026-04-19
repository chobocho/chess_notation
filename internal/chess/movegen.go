package chess

var knightOffsets = [8][2]int{
	{1, 2}, {2, 1}, {2, -1}, {1, -2},
	{-1, -2}, {-2, -1}, {-2, 1}, {-1, 2},
}

var kingOffsets = [8][2]int{
	{1, 0}, {1, 1}, {0, 1}, {-1, 1},
	{-1, 0}, {-1, -1}, {0, -1}, {1, -1},
}

var bishopDirs = [4][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
var rookDirs = [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

// IsSquareAttacked reports whether `by` attacks `sq` in `pos`.
// Considers raw piece attacks; ignores own-king legality.
func (pos *Position) IsSquareAttacked(sq Square, by Color) bool {
	f, r := sq.File(), sq.Rank()

	// Pawn attacks: a pawn of color `by` on square x attacks squares diagonally forward.
	// So sq is attacked by a pawn of color `by` if there is such a pawn diagonally "backward" from sq's perspective.
	var pawnDir int
	if by == White {
		pawnDir = -1 // the attacking pawn would be one rank below sq
	} else {
		pawnDir = 1
	}
	for _, df := range [2]int{-1, 1} {
		nf, nr := f+df, r+pawnDir
		if nf >= 0 && nf < 8 && nr >= 0 && nr < 8 {
			p := pos.Board[Sq(nf, nr)]
			if p.Color() == by && p.Type() == Pawn && p != Empty {
				return true
			}
		}
	}

	// Knight.
	for _, o := range knightOffsets {
		nf, nr := f+o[0], r+o[1]
		if nf < 0 || nf >= 8 || nr < 0 || nr >= 8 {
			continue
		}
		p := pos.Board[Sq(nf, nr)]
		if p != Empty && p.Color() == by && p.Type() == Knight {
			return true
		}
	}

	// King.
	for _, o := range kingOffsets {
		nf, nr := f+o[0], r+o[1]
		if nf < 0 || nf >= 8 || nr < 0 || nr >= 8 {
			continue
		}
		p := pos.Board[Sq(nf, nr)]
		if p != Empty && p.Color() == by && p.Type() == King {
			return true
		}
	}

	// Bishop / Queen diagonals.
	for _, d := range bishopDirs {
		nf, nr := f+d[0], r+d[1]
		for nf >= 0 && nf < 8 && nr >= 0 && nr < 8 {
			p := pos.Board[Sq(nf, nr)]
			if p != Empty {
				if p.Color() == by && (p.Type() == Bishop || p.Type() == Queen) {
					return true
				}
				break
			}
			nf += d[0]
			nr += d[1]
		}
	}

	// Rook / Queen orthogonals.
	for _, d := range rookDirs {
		nf, nr := f+d[0], r+d[1]
		for nf >= 0 && nf < 8 && nr >= 0 && nr < 8 {
			p := pos.Board[Sq(nf, nr)]
			if p != Empty {
				if p.Color() == by && (p.Type() == Rook || p.Type() == Queen) {
					return true
				}
				break
			}
			nf += d[0]
			nr += d[1]
		}
	}

	return false
}

// InCheck reports whether the side to move is in check.
func (pos *Position) InCheck() bool {
	ks := pos.KingSquare(pos.SideToMove)
	if ks == NoSquare {
		return false
	}
	return pos.IsSquareAttacked(ks, pos.SideToMove.Opp())
}

// GeneratePseudoLegal returns all moves ignoring king-in-check legality.
func (pos *Position) GeneratePseudoLegal() []Move {
	moves := make([]Move, 0, 64)
	us := pos.SideToMove
	for s := Square(0); s < 64; s++ {
		p := pos.Board[s]
		if p == Empty || p.Color() != us {
			continue
		}
		switch p.Type() {
		case Pawn:
			moves = pos.genPawn(moves, s)
		case Knight:
			moves = pos.genLeaper(moves, s, knightOffsets[:])
		case Bishop:
			moves = pos.genSlider(moves, s, bishopDirs[:])
		case Rook:
			moves = pos.genSlider(moves, s, rookDirs[:])
		case Queen:
			moves = pos.genSlider(moves, s, bishopDirs[:])
			moves = pos.genSlider(moves, s, rookDirs[:])
		case King:
			moves = pos.genLeaper(moves, s, kingOffsets[:])
			moves = pos.genCastle(moves, s)
		}
	}
	return moves
}

func (pos *Position) genLeaper(moves []Move, from Square, offsets [][2]int) []Move {
	us := pos.SideToMove
	f, r := from.File(), from.Rank()
	for _, o := range offsets {
		nf, nr := f+o[0], r+o[1]
		if nf < 0 || nf >= 8 || nr < 0 || nr >= 8 {
			continue
		}
		to := Sq(nf, nr)
		p := pos.Board[to]
		if p != Empty && p.Color() == us {
			continue
		}
		var flags MoveFlag
		if p != Empty {
			flags |= FlagCapture
		}
		moves = append(moves, Move{From: from, To: to, Flags: flags})
	}
	return moves
}

func (pos *Position) genSlider(moves []Move, from Square, dirs [][2]int) []Move {
	us := pos.SideToMove
	f, r := from.File(), from.Rank()
	for _, d := range dirs {
		nf, nr := f+d[0], r+d[1]
		for nf >= 0 && nf < 8 && nr >= 0 && nr < 8 {
			to := Sq(nf, nr)
			p := pos.Board[to]
			if p == Empty {
				moves = append(moves, Move{From: from, To: to})
			} else {
				if p.Color() != us {
					moves = append(moves, Move{From: from, To: to, Flags: FlagCapture})
				}
				break
			}
			nf += d[0]
			nr += d[1]
		}
	}
	return moves
}

func (pos *Position) genPawn(moves []Move, from Square) []Move {
	us := pos.SideToMove
	f, r := from.File(), from.Rank()
	var dir, startRank, promoRank int
	if us == White {
		dir = 1
		startRank = 1
		promoRank = 7
	} else {
		dir = -1
		startRank = 6
		promoRank = 0
	}

	// Single push.
	nr := r + dir
	if nr >= 0 && nr < 8 {
		to := Sq(f, nr)
		if pos.Board[to] == Empty {
			if nr == promoRank {
				moves = appendPromotions(moves, from, to, 0)
			} else {
				moves = append(moves, Move{From: from, To: to})
			}
			// Double push.
			if r == startRank {
				to2 := Sq(f, r+2*dir)
				if pos.Board[to2] == Empty {
					moves = append(moves, Move{From: from, To: to2, Flags: FlagPawnDouble})
				}
			}
		}
	}

	// Captures.
	for _, df := range [2]int{-1, 1} {
		nf, nr := f+df, r+dir
		if nf < 0 || nf >= 8 || nr < 0 || nr >= 8 {
			continue
		}
		to := Sq(nf, nr)
		p := pos.Board[to]
		if p != Empty && p.Color() != us {
			if nr == promoRank {
				moves = appendPromotions(moves, from, to, FlagCapture)
			} else {
				moves = append(moves, Move{From: from, To: to, Flags: FlagCapture})
			}
		}
		// En passant.
		if pos.EPSquare != NoSquare && to == pos.EPSquare {
			moves = append(moves, Move{From: from, To: to, Flags: FlagCapture | FlagEnPassant})
		}
	}
	return moves
}

func appendPromotions(moves []Move, from, to Square, extra MoveFlag) []Move {
	for _, pt := range [4]PieceType{Queen, Rook, Bishop, Knight} {
		moves = append(moves, Move{
			From:      from,
			To:        to,
			Promotion: pt,
			Flags:     FlagPromotion | extra,
		})
	}
	return moves
}

func (pos *Position) genCastle(moves []Move, from Square) []Move {
	us := pos.SideToMove
	them := us.Opp()
	var rank int
	var ks, qs Castling
	if us == White {
		rank = 0
		ks, qs = WhiteKingside, WhiteQueenside
	} else {
		rank = 7
		ks, qs = BlackKingside, BlackQueenside
	}
	if from != Sq(4, rank) {
		return moves
	}
	if pos.IsSquareAttacked(Sq(4, rank), them) {
		return moves
	}

	if pos.Castling&ks != 0 {
		if pos.Board[Sq(5, rank)] == Empty && pos.Board[Sq(6, rank)] == Empty &&
			pos.Board[Sq(7, rank)] == MakePiece(us, Rook) &&
			!pos.IsSquareAttacked(Sq(5, rank), them) &&
			!pos.IsSquareAttacked(Sq(6, rank), them) {
			moves = append(moves, Move{From: from, To: Sq(6, rank), Flags: FlagCastleKing})
		}
	}
	if pos.Castling&qs != 0 {
		if pos.Board[Sq(1, rank)] == Empty && pos.Board[Sq(2, rank)] == Empty &&
			pos.Board[Sq(3, rank)] == Empty &&
			pos.Board[Sq(0, rank)] == MakePiece(us, Rook) &&
			!pos.IsSquareAttacked(Sq(3, rank), them) &&
			!pos.IsSquareAttacked(Sq(2, rank), them) {
			moves = append(moves, Move{From: from, To: Sq(2, rank), Flags: FlagCastleQueen})
		}
	}
	return moves
}

// GenerateLegal returns only legal moves.
func (pos *Position) GenerateLegal() []Move {
	pseudo := pos.GeneratePseudoLegal()
	legal := make([]Move, 0, len(pseudo))
	for _, m := range pseudo {
		u := pos.Make(m)
		// After Make, SideToMove has flipped. The mover's color is pos.SideToMove.Opp().
		mover := pos.SideToMove.Opp()
		ks := pos.KingSquare(mover)
		if ks != NoSquare && !pos.IsSquareAttacked(ks, mover.Opp()) {
			legal = append(legal, m)
		}
		pos.Unmake(u)
	}
	return legal
}
