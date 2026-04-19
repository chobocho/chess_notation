// Package pgn parses Portable Game Notation, including tag pairs, movetext,
// comments (`{}` and `;`), NAGs (`$`), recursive variations (`()`), and game
// termination markers.
package pgn

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/chobocho/chess_notation/internal/chess"
)

// Parse reads PGN text and returns all games, validating each move against the chess engine.
func Parse(r io.Reader) ([]*chess.Game, error) {
	br := bufio.NewReader(r)
	var all []*chess.Game
	// Buffer the whole input; PGN files are typically small.
	var sb strings.Builder
	for {
		b, err := br.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		sb.WriteByte(b)
	}
	text := sb.String()

	p := &parser{src: text}
	for !p.eof() {
		p.skipWhitespace()
		if p.eof() {
			break
		}
		g, err := p.parseGame()
		if err != nil {
			return nil, err
		}
		if g != nil {
			all = append(all, g)
		}
	}
	return all, nil
}

// ParseString is a convenience wrapper.
func ParseString(s string) ([]*chess.Game, error) {
	return Parse(strings.NewReader(s))
}

type parser struct {
	src string
	pos int
}

func (p *parser) eof() bool { return p.pos >= len(p.src) }

func (p *parser) peek() byte {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) next() byte {
	if p.eof() {
		return 0
	}
	b := p.src[p.pos]
	p.pos++
	return b
}

func (p *parser) skipWhitespace() {
	for !p.eof() {
		b := p.peek()
		switch {
		case b == ' ' || b == '\t' || b == '\r' || b == '\n':
			p.pos++
		case b == ';': // line comment to EOL
			for !p.eof() && p.peek() != '\n' {
				p.pos++
			}
		default:
			return
		}
	}
}

// parseGame parses one game; returns nil if only whitespace/comments were consumed.
func (p *parser) parseGame() (*chess.Game, error) {
	tags := map[string]string{}
	for !p.eof() {
		p.skipWhitespace()
		if p.eof() || p.peek() != '[' {
			break
		}
		k, v, err := p.parseTagPair()
		if err != nil {
			return nil, err
		}
		tags[k] = v
	}
	p.skipWhitespace()
	if p.eof() && len(tags) == 0 {
		return nil, nil
	}

	// Build game with optional FEN tag.
	var g *chess.Game
	if fen, ok := tags["FEN"]; ok {
		gg, err := chess.NewGameFromFEN(fen)
		if err != nil {
			return nil, fmt.Errorf("pgn: bad FEN tag: %w", err)
		}
		g = gg
	} else {
		g = chess.NewGame()
	}
	for k, v := range tags {
		g.Tags[k] = v
	}

	if err := p.parseMovetext(g, g.Root); err != nil {
		return nil, err
	}
	return g, nil
}

func (p *parser) parseTagPair() (string, string, error) {
	if p.next() != '[' {
		return "", "", fmt.Errorf("pgn: expected '['")
	}
	p.skipInlineWhitespace()
	// Key.
	start := p.pos
	for !p.eof() {
		c := p.peek()
		if c == ' ' || c == '\t' || c == '"' {
			break
		}
		p.pos++
	}
	key := p.src[start:p.pos]
	p.skipInlineWhitespace()
	if p.eof() || p.next() != '"' {
		return "", "", fmt.Errorf("pgn: expected '\"' in tag %q", key)
	}
	// Value (supports escaped \" and \\).
	var val strings.Builder
	for !p.eof() {
		c := p.next()
		if c == '\\' {
			if p.eof() {
				break
			}
			val.WriteByte(p.next())
			continue
		}
		if c == '"' {
			break
		}
		val.WriteByte(c)
	}
	p.skipInlineWhitespace()
	if !p.eof() && p.peek() == ']' {
		p.pos++
	}
	return key, val.String(), nil
}

func (p *parser) skipInlineWhitespace() {
	for !p.eof() {
		c := p.peek()
		if c == ' ' || c == '\t' {
			p.pos++
			continue
		}
		return
	}
}

// parseMovetext consumes tokens into the game tree starting at cursor.
// It recursively descends into variations enclosed in '(' ')'.
func (p *parser) parseMovetext(g *chess.Game, startNode *chess.Node) error {
	savedCur := g.Cur
	g.Cur = startNode
	defer func() { g.Cur = savedCur }()

	for {
		p.skipWhitespace()
		if p.eof() {
			return nil
		}
		c := p.peek()
		switch {
		case c == '{':
			p.pos++
			start := p.pos
			for !p.eof() && p.peek() != '}' {
				p.pos++
			}
			comment := strings.TrimSpace(p.src[start:p.pos])
			if !p.eof() {
				p.pos++ // consume '}'
			}
			if g.Cur.Comment == "" {
				g.Cur.Comment = comment
			} else {
				g.Cur.Comment += " " + comment
			}
		case c == '$':
			p.pos++
			start := p.pos
			for !p.eof() && p.peek() >= '0' && p.peek() <= '9' {
				p.pos++
			}
			if p.pos > start {
				n := 0
				for i := start; i < p.pos; i++ {
					n = n*10 + int(p.src[i]-'0')
				}
				g.Cur.NAGs = append(g.Cur.NAGs, n)
			}
		case c == '(':
			p.pos++
			// Variation is a sibling of g.Cur, played from g.Cur.Parent.
			if g.Cur.Parent == nil {
				// Variation at the root has no anchor; skip balanced parens.
				if err := p.skipVariation(); err != nil {
					return err
				}
				continue
			}
			anchor := g.Cur.Parent
			if err := p.parseMovetext(g, anchor); err != nil {
				return err
			}
			p.skipWhitespace()
			if !p.eof() && p.peek() == ')' {
				p.pos++
			}
		case c == ')':
			// End of current variation: caller handles the ')'.
			return nil
		case c == '[':
			// Start of next game's tags.
			return nil
		default:
			// Token: move number, move, or result.
			tok := p.readToken()
			if tok == "" {
				return nil
			}
			if isResult(tok) {
				return nil
			}
			if san := stripMoveNumber(tok); san != "" {
				if _, err := g.AddSAN(san); err != nil {
					return fmt.Errorf("pgn: move %q: %w", san, err)
				}
			} else if tok == "." || strings.HasSuffix(tok, ".") {
				// Pure move number token like "1." or "1...": skip.
				continue
			}
		}
	}
}

func (p *parser) readToken() string {
	start := p.pos
	for !p.eof() {
		c := p.peek()
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' ||
			c == '{' || c == '}' || c == '(' || c == ')' ||
			c == '[' || c == ']' || c == ';' || c == '$' {
			break
		}
		p.pos++
	}
	return p.src[start:p.pos]
}

func (p *parser) skipVariation() error {
	depth := 1
	for !p.eof() && depth > 0 {
		c := p.next()
		switch c {
		case '(':
			depth++
		case ')':
			depth--
		case '{':
			for !p.eof() && p.next() != '}' {
			}
		}
	}
	return nil
}

func isResult(tok string) bool {
	switch tok {
	case "1-0", "0-1", "1/2-1/2", "*":
		return true
	}
	return false
}

// stripMoveNumber removes a leading "N." or "N..." prefix. Returns the remaining SAN,
// or "" if the token was purely a move number.
func stripMoveNumber(tok string) string {
	i := 0
	for i < len(tok) && tok[i] >= '0' && tok[i] <= '9' {
		i++
	}
	if i == 0 {
		return tok
	}
	// After digits, expect dots or done.
	for i < len(tok) && tok[i] == '.' {
		i++
	}
	if i == len(tok) {
		return ""
	}
	return tok[i:]
}
