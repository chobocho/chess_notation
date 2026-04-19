package chess

import "testing"

func perft(pos *Position, depth int) int64 {
	if depth == 0 {
		return 1
	}
	moves := pos.GenerateLegal()
	if depth == 1 {
		return int64(len(moves))
	}
	var n int64
	for _, m := range moves {
		u := pos.Make(m)
		n += perft(pos, depth-1)
		pos.Unmake(u)
	}
	return n
}

// Standard perft reference values (chessprogramming.org).
type perftCase struct {
	name  string
	fen   string
	depth int
	nodes int64
}

var perftCases = []perftCase{
	{"startpos d1", StartFEN, 1, 20},
	{"startpos d2", StartFEN, 2, 400},
	{"startpos d3", StartFEN, 3, 8902},
	{"startpos d4", StartFEN, 4, 197281},
	{"kiwipete d1", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 1, 48},
	{"kiwipete d2", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 2, 2039},
	{"kiwipete d3", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 3, 97862},
	{"position3 d1", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 1, 14},
	{"position3 d4", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 4, 43238},
	{"position4 d1", "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 1, 6},
	{"position4 d3", "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 3, 9467},
	{"position5 d3", "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 3, 62379},
	{"position6 d3", "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 3, 89890},
}

func TestPerft(t *testing.T) {
	for _, c := range perftCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			pos, err := ParseFEN(c.fen)
			if err != nil {
				t.Fatalf("parse fen: %v", err)
			}
			got := perft(pos, c.depth)
			if got != c.nodes {
				t.Fatalf("perft(%q, %d) = %d, want %d", c.fen, c.depth, got, c.nodes)
			}
		})
	}
}
