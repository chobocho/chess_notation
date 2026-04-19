package chess

import (
	"fmt"
	"strings"
)

// ParseSAN parses a SAN move in the context of pos and returns the matching Move.
// It handles castling, pawn captures, promotions, disambiguators, and trailing check/mate markers.
func ParseSAN(pos *Position, san string) (Move, error) {
	orig := san
	san = strings.TrimSpace(san)
	// Strip trailing annotation markers.
	for len(san) > 0 {
		last := san[len(san)-1]
		if last == '+' || last == '#' || last == '!' || last == '?' {
			san = san[:len(san)-1]
			continue
		}
		break
	}
	if san == "" {
		return Move{}, fmt.Errorf("san: empty move %q", orig)
	}

	legal := pos.GenerateLegal()

	// Castling.
	normalized := strings.ReplaceAll(san, "0", "O")
	if normalized == "O-O" || normalized == "O-O-O" {
		want := FlagCastleKing
		if normalized == "O-O-O" {
			want = FlagCastleQueen
		}
		for _, m := range legal {
			if m.Flags&want != 0 {
				return m, nil
			}
		}
		return Move{}, fmt.Errorf("san: castling %q not legal", orig)
	}

	// Determine piece type.
	piece := Pawn
	i := 0
	if san[0] >= 'A' && san[0] <= 'Z' && san[0] != 'O' {
		t := PieceTypeFromLetter(san[0])
		if t == NoPiece {
			return Move{}, fmt.Errorf("san: bad piece letter in %q", orig)
		}
		piece = t
		i++
	}

	// Trailing promotion "=Q" / "Q" / "(Q)".
	promo := NoPiece
	if piece == Pawn {
		// Find a promotion suffix at the end.
		rest := san[i:]
		// "(Q)"
		if n := len(rest); n >= 3 && rest[n-3] == '(' && rest[n-1] == ')' {
			if t := PieceTypeFromLetter(rest[n-2]); t != NoPiece {
				promo = t
				san = san[:len(san)-3]
			}
		}
		if promo == NoPiece {
			n := len(san)
			if n >= 2 && san[n-2] == '=' {
				if t := PieceTypeFromLetter(san[n-1]); t != NoPiece {
					promo = t
					san = san[:n-2]
				}
			}
		}
		if promo == NoPiece {
			n := len(san)
			if n >= 1 {
				if t := PieceTypeFromLetter(san[n-1]); t != NoPiece {
					promo = t
					san = san[:n-1]
				}
			}
		}
	}

	body := san[i:]
	if len(body) < 2 {
		return Move{}, fmt.Errorf("san: too short %q", orig)
	}

	// Destination is always the last two characters (after optional x).
	destStr := body[len(body)-2:]
	toSq, err := ParseSquare(destStr)
	if err != nil {
		return Move{}, fmt.Errorf("san: bad dest in %q: %w", orig, err)
	}
	mid := body[:len(body)-2]

	capture := false
	if n := len(mid); n > 0 && mid[n-1] == 'x' {
		capture = true
		mid = mid[:n-1]
	}

	disambFile := -1
	disambRank := -1
	for j := 0; j < len(mid); j++ {
		c := mid[j]
		switch {
		case c >= 'a' && c <= 'h':
			disambFile = int(c - 'a')
		case c >= '1' && c <= '8':
			disambRank = int(c - '1')
		default:
			return Move{}, fmt.Errorf("san: bad disambiguator in %q", orig)
		}
	}

	// Find matching legal move.
	var matches []Move
	for _, m := range legal {
		if m.To != toSq {
			continue
		}
		moverPiece := pos.Board[m.From]
		if moverPiece.Type() != piece {
			continue
		}
		if disambFile >= 0 && m.From.File() != disambFile {
			continue
		}
		if disambRank >= 0 && m.From.Rank() != disambRank {
			continue
		}
		if promo != NoPiece && m.Promotion != promo {
			continue
		}
		if promo == NoPiece && m.IsPromotion() {
			continue
		}
		if capture && !m.IsCapture() {
			continue
		}
		matches = append(matches, m)
	}
	if len(matches) == 0 {
		return Move{}, fmt.Errorf("san: no legal move matches %q", orig)
	}
	if len(matches) > 1 {
		return Move{}, fmt.Errorf("san: ambiguous move %q (%d matches)", orig, len(matches))
	}
	return matches[0], nil
}

// EncodeSAN returns the standard algebraic notation for m in pos.
// Caller must supply a move that is legal in pos.
func EncodeSAN(pos *Position, m Move) string {
	if m.Flags&FlagCastleKing != 0 {
		return addCheckSuffix(pos, m, "O-O")
	}
	if m.Flags&FlagCastleQueen != 0 {
		return addCheckSuffix(pos, m, "O-O-O")
	}

	mover := pos.Board[m.From]
	var sb strings.Builder

	if mover.Type() == Pawn {
		if m.IsCapture() {
			sb.WriteByte(byte('a' + m.From.File()))
			sb.WriteByte('x')
		}
		sb.WriteString(m.To.String())
		if m.IsPromotion() {
			sb.WriteByte('=')
			sb.WriteByte(PieceTypeLetter(m.Promotion))
		}
		return addCheckSuffix(pos, m, sb.String())
	}

	sb.WriteByte(PieceTypeLetter(mover.Type()))

	// Disambiguation: look at other legal moves of same piece type that land on m.To.
	legal := pos.GenerateLegal()
	var ambig []Move
	for _, other := range legal {
		if other.From == m.From || other.To != m.To {
			continue
		}
		if pos.Board[other.From].Type() != mover.Type() {
			continue
		}
		ambig = append(ambig, other)
	}
	if len(ambig) > 0 {
		sameFile, sameRank := false, false
		for _, o := range ambig {
			if o.From.File() == m.From.File() {
				sameFile = true
			}
			if o.From.Rank() == m.From.Rank() {
				sameRank = true
			}
		}
		switch {
		case !sameFile:
			sb.WriteByte(byte('a' + m.From.File()))
		case !sameRank:
			sb.WriteByte(byte('1' + m.From.Rank()))
		default:
			sb.WriteByte(byte('a' + m.From.File()))
			sb.WriteByte(byte('1' + m.From.Rank()))
		}
	}

	if m.IsCapture() {
		sb.WriteByte('x')
	}
	sb.WriteString(m.To.String())
	return addCheckSuffix(pos, m, sb.String())
}

// addCheckSuffix applies m to a clone and appends + or # if appropriate.
func addCheckSuffix(pos *Position, m Move, s string) string {
	p := pos.Clone()
	p.Make(m)
	if !p.InCheck() {
		return s
	}
	if len(p.GenerateLegal()) == 0 {
		return s + "#"
	}
	return s + "+"
}
