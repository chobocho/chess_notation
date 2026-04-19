package pgn

import (
	"strings"
	"testing"
)

const samplePGN = `[Event "Test"]
[Site "?"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 4. Ba4 Nf6 5. O-O Be7 6. Re1 b5 7. Bb3 d6
8. c3 O-O 9. h3 Nb8 10. d4 Nbd7 1-0
`

func TestParseOne(t *testing.T) {
	games, err := ParseString(samplePGN)
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 1 {
		t.Fatalf("got %d games, want 1", len(games))
	}
	g := games[0]
	if g.Tags["White"] != "Alice" {
		t.Fatalf("White tag = %q", g.Tags["White"])
	}
	sans := g.MainlineSAN()
	if len(sans) != 20 {
		t.Fatalf("got %d plies, want 20", len(sans))
	}
	if sans[0] != "e4" {
		t.Fatalf("first move = %q", sans[0])
	}
	if sans[8] != "O-O" {
		t.Fatalf("ply 9 = %q (want O-O)", sans[8])
	}
}

func TestParseMultipleGames(t *testing.T) {
	src := samplePGN + "\n\n" + samplePGN
	games, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 2 {
		t.Fatalf("got %d games, want 2", len(games))
	}
}

func TestParseVariationsAndComments(t *testing.T) {
	src := `[Event "Var"]
[Result "*"]

1. e4 {King's pawn} e5 (1... c5 {Sicilian}) 2. Nf3 $1 Nc6 *
`
	games, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	g := games[0]
	sans := g.MainlineSAN()
	if strings.Join(sans, " ") != "e4 e5 Nf3 Nc6" {
		t.Fatalf("mainline = %v", sans)
	}
	// The root's first child is e4; its first child should be e5, and e5's
	// parent (which is e4) should also have c5 as a sibling variation.
	e4 := g.Root.Children[0]
	if e4.SAN != "e4" {
		t.Fatalf("e4 node = %q", e4.SAN)
	}
	if len(e4.Children) < 2 {
		t.Fatalf("expected variation sibling on e4, got %d children", len(e4.Children))
	}
	if e4.Children[1].SAN != "c5" {
		t.Fatalf("variation = %q", e4.Children[1].SAN)
	}
	// Comment attached.
	if !strings.Contains(e4.Comment, "King's pawn") {
		t.Fatalf("comment = %q", e4.Comment)
	}
	// NAG on Nf3.
	nf3 := g.Root.Children[0].Children[0].Children[0]
	if nf3.SAN != "Nf3" {
		t.Fatalf("nf3 = %q", nf3.SAN)
	}
	if len(nf3.NAGs) == 0 || nf3.NAGs[0] != 1 {
		t.Fatalf("nag = %v", nf3.NAGs)
	}
}
