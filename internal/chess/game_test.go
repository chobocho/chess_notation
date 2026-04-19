package chess

import "testing"

func TestGameForwardBack(t *testing.T) {
	g := NewGame()
	for _, s := range []string{"e4", "e5", "Nf3", "Nc6"} {
		if _, err := g.AddSAN(s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
	if got := g.Cur.Ply; got != 4 {
		t.Fatalf("ply = %d", got)
	}
	g.Back()
	g.Back()
	if g.Cur.SAN != "e5" {
		t.Fatalf("cur.SAN = %q", g.Cur.SAN)
	}
	g.Next()
	if g.Cur.SAN != "Nf3" {
		t.Fatalf("next.SAN = %q", g.Cur.SAN)
	}
	g.Goto(0)
	if g.Cur != g.Root {
		t.Fatalf("goto(0) did not reach root")
	}
	g.Goto(100)
	if g.Cur.SAN != "Nc6" {
		t.Fatalf("goto past end = %q", g.Cur.SAN)
	}
}

func TestMovetextMainline(t *testing.T) {
	g := NewGame()
	for _, s := range []string{"e4", "e5", "Nf3", "Nc6"} {
		if _, err := g.AddSAN(s); err != nil {
			t.Fatal(err)
		}
	}
	got := g.MovetextMainline()
	want := "1. e4 e5 2. Nf3 Nc6"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
