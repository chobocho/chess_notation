package chess

import (
	"errors"
	"strings"
)

// Node is a node in the game history tree.
// The root node has Move.From == NoSquare and represents the starting position.
type Node struct {
	Move     Move
	SAN      string // SAN in its parent position; empty at root
	UCI      string
	FEN      string // FEN of the position reached by applying Move
	Ply      int    // 0 at root
	Parent   *Node
	Children []*Node // first child is mainline continuation
	Comment  string
	NAGs     []int
}

// Game represents a single game with a move tree and metadata.
type Game struct {
	Tags  map[string]string
	Root  *Node
	Start *Position // starting position (clone); nil means StartFEN
	Cur   *Node     // current cursor; equal to Root at start
}

// NewGame creates an empty game starting from the standard position.
func NewGame() *Game {
	start := NewPosition()
	g := &Game{
		Tags:  map[string]string{},
		Start: start.Clone(),
	}
	g.Root = &Node{Ply: 0, FEN: start.FEN(), Move: Move{From: NoSquare, To: NoSquare}}
	g.Cur = g.Root
	return g
}

// NewGameFromFEN creates a game whose root is the given FEN.
func NewGameFromFEN(fen string) (*Game, error) {
	pos, err := ParseFEN(fen)
	if err != nil {
		return nil, err
	}
	g := &Game{
		Tags:  map[string]string{},
		Start: pos.Clone(),
	}
	g.Root = &Node{Ply: 0, FEN: pos.FEN(), Move: Move{From: NoSquare, To: NoSquare}}
	g.Cur = g.Root
	return g, nil
}

// PositionAt reconstructs the position at node n by replaying moves from the start.
func (g *Game) PositionAt(n *Node) *Position {
	pos := g.Start.Clone()
	var path []*Node
	for x := n; x != nil && x != g.Root; x = x.Parent {
		path = append(path, x)
	}
	for i := len(path) - 1; i >= 0; i-- {
		pos.Make(path[i].Move)
	}
	return pos
}

// AddMove adds a move as a child of the current cursor. The move must be legal.
// Returns the new node and advances the cursor to it.
// If a mainline child already exists with the same move, the cursor simply advances.
func (g *Game) AddMove(m Move) (*Node, error) {
	pos := g.PositionAt(g.Cur)
	legal := pos.GenerateLegal()
	found := false
	for _, lm := range legal {
		if lm == m {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("chess: move not legal in current position")
	}
	// If identical move already exists as a child, reuse it.
	for _, ch := range g.Cur.Children {
		if ch.Move == m {
			g.Cur = ch
			return ch, nil
		}
	}
	san := EncodeSAN(pos, m)
	pos.Make(m)
	node := &Node{
		Move:   m,
		SAN:    san,
		UCI:    m.UCI(),
		FEN:    pos.FEN(),
		Ply:    g.Cur.Ply + 1,
		Parent: g.Cur,
	}
	g.Cur.Children = append(g.Cur.Children, node)
	g.Cur = node
	return node, nil
}

// AddSAN parses san in the current position and adds it as a move.
func (g *Game) AddSAN(san string) (*Node, error) {
	pos := g.PositionAt(g.Cur)
	m, err := ParseSAN(pos, san)
	if err != nil {
		return nil, err
	}
	return g.AddMove(m)
}

// Next advances the cursor to the first child (mainline). Returns nil if none.
func (g *Game) Next() *Node {
	if g.Cur == nil || len(g.Cur.Children) == 0 {
		return nil
	}
	g.Cur = g.Cur.Children[0]
	return g.Cur
}

// Back moves the cursor to the parent. Returns nil at root.
func (g *Game) Back() *Node {
	if g.Cur == nil || g.Cur.Parent == nil {
		return nil
	}
	g.Cur = g.Cur.Parent
	return g.Cur
}

// Goto walks the mainline to the given ply (from the root). Clamps at ends.
func (g *Game) Goto(ply int) *Node {
	n := g.Root
	for n.Ply < ply && len(n.Children) > 0 {
		n = n.Children[0]
	}
	for n.Ply > ply && n.Parent != nil {
		n = n.Parent
	}
	g.Cur = n
	return n
}

// Mainline returns the mainline nodes from the first move.
func (g *Game) Mainline() []*Node {
	var out []*Node
	n := g.Root
	for len(n.Children) > 0 {
		n = n.Children[0]
		out = append(out, n)
	}
	return out
}

// MainlineSAN returns SANs along the mainline.
func (g *Game) MainlineSAN() []string {
	nodes := g.Mainline()
	out := make([]string, len(nodes))
	for i, n := range nodes {
		out[i] = n.SAN
	}
	return out
}

// MovetextMainline renders a space-separated move list with move numbers, mainline only.
func (g *Game) MovetextMainline() string {
	var sb strings.Builder
	nodes := g.Mainline()
	// Determine side to move at the root (from Start).
	startSide := White
	if g.Start != nil {
		startSide = g.Start.SideToMove
	}
	startFullmove := 1
	if g.Start != nil {
		startFullmove = g.Start.FullmoveNumber
	}
	for i, n := range nodes {
		side := startSide
		if i%2 == 1 {
			side = startSide.Opp()
		}
		if side == White {
			if sb.Len() > 0 {
				sb.WriteByte(' ')
			}
			// Move number.
			full := startFullmove + i/2
			sb.WriteString(itoa(full))
			sb.WriteByte('.')
			sb.WriteByte(' ')
		} else if i == 0 {
			// Black-to-move start: write "N..."
			sb.WriteString(itoa(startFullmove))
			sb.WriteString("... ")
		} else {
			sb.WriteByte(' ')
		}
		sb.WriteString(n.SAN)
	}
	return sb.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
