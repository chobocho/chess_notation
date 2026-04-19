package chess

import "testing"

func TestParseSANBasic(t *testing.T) {
	pos := NewPosition()
	cases := []struct {
		san  string
		uci  string
	}{
		{"e4", "e2e4"},
		{"Nf3", "g1f3"},
	}
	for _, c := range cases {
		p := pos.Clone()
		m, err := ParseSAN(p, c.san)
		if err != nil {
			t.Fatalf("%s: %v", c.san, err)
		}
		if m.UCI() != c.uci {
			t.Fatalf("%s: got uci %s, want %s", c.san, m.UCI(), c.uci)
		}
	}
}

func TestCastlingSAN(t *testing.T) {
	pos, _ := ParseFEN("r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1")
	m, err := ParseSAN(pos, "O-O")
	if err != nil {
		t.Fatal(err)
	}
	if m.UCI() != "e1g1" {
		t.Fatalf("O-O uci = %s", m.UCI())
	}
	// Re-encode.
	if s := EncodeSAN(pos, m); s != "O-O" {
		t.Fatalf("encode O-O = %s", s)
	}
	m, err = ParseSAN(pos, "O-O-O")
	if err != nil {
		t.Fatal(err)
	}
	if m.UCI() != "e1c1" {
		t.Fatalf("O-O-O uci = %s", m.UCI())
	}
}

func TestDisambiguation(t *testing.T) {
	// Two knights both able to go to d2: Nbd2 vs Nfd2 (b1 and f3 -> d2).
	pos, _ := ParseFEN("4k3/8/8/8/8/5N2/8/1N2K3 w - - 0 1")
	m, err := ParseSAN(pos, "Nbd2")
	if err != nil {
		t.Fatal(err)
	}
	if m.From.String() != "b1" {
		t.Fatalf("Nbd2 from = %s", m.From.String())
	}
	if s := EncodeSAN(pos, m); s != "Nbd2" {
		t.Fatalf("encode = %s", s)
	}
	m, err = ParseSAN(pos, "Nfd2")
	if err != nil {
		t.Fatal(err)
	}
	if m.From.String() != "f3" {
		t.Fatalf("Nfd2 from = %s", m.From.String())
	}
}

func TestEnPassant(t *testing.T) {
	pos, _ := ParseFEN("rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3")
	m, err := ParseSAN(pos, "exf6")
	if err != nil {
		t.Fatal(err)
	}
	if !m.IsEnPassant() {
		t.Fatalf("expected en passant flag")
	}
	if m.UCI() != "e5f6" {
		t.Fatalf("uci = %s", m.UCI())
	}
}

func TestPromotion(t *testing.T) {
	pos, _ := ParseFEN("8/P7/8/8/8/8/8/4k2K w - - 0 1")
	m, err := ParseSAN(pos, "a8=Q")
	if err != nil {
		t.Fatal(err)
	}
	if m.Promotion != Queen {
		t.Fatalf("promotion = %v", m.Promotion)
	}
	if m.UCI() != "a7a8q" {
		t.Fatalf("uci = %s", m.UCI())
	}
	// Also accept without "=".
	m, err = ParseSAN(pos, "a8Q")
	if err != nil {
		t.Fatal(err)
	}
	if m.Promotion != Queen {
		t.Fatalf("a8Q promo = %v", m.Promotion)
	}
}

func TestCheckmateSuffix(t *testing.T) {
	// Scholar's mate: Qxf7#.
	fen := "r1bqkb1r/pppp1Qpp/2n2n2/4p3/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 0 4"
	pos, err := ParseFEN(fen)
	if err != nil {
		t.Fatal(err)
	}
	if !pos.InCheck() {
		t.Fatalf("expected check")
	}
	if len(pos.GenerateLegal()) != 0 {
		t.Fatalf("expected mate")
	}
	// Build the mating move and check encoding in the previous position.
	prev, _ := ParseFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/8/PPPP1PPP/RNBQK1NR w KQkq - 0 1")
	// Set white queen's move Qxf7# from d1 -> f7 won't work without pieces set; use a known prior.
	prev, _ = ParseFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/8/PPPP1PPP/RNBQK1NR w KQkq - 0 1")
	// Play a sequence that leads to the mate: 3. Qh5 Nf6?? 4. Qxf7#.
	// To keep this test tight, test encoding on a mate-in-one position directly.
	matePos, _ := ParseFEN("r1bqkb1r/pppp1ppp/2n2n2/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 0 4")
	m, err := ParseSAN(matePos, "Qxf7#")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if s := EncodeSAN(matePos, m); s != "Qxf7#" {
		t.Fatalf("encode = %s, want Qxf7#", s)
	}
	_ = prev
	_ = fen
}

func TestSANRoundtripKasparovOpening(t *testing.T) {
	// A short known-good opening sequence.
	sans := []string{
		"e4", "c5", "Nf3", "d6", "d4", "cxd4", "Nxd4", "Nf6", "Nc3", "a6",
		"Be2", "e5", "Nb3", "Be7", "O-O", "O-O",
	}
	pos := NewPosition()
	for _, s := range sans {
		m, err := ParseSAN(pos, s)
		if err != nil {
			t.Fatalf("parse %q: %v", s, err)
		}
		got := EncodeSAN(pos, m)
		if got != s {
			t.Fatalf("roundtrip: parsed %q -> encoded %q", s, got)
		}
		pos.Make(m)
	}
}
