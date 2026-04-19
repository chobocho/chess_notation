package pgn

import "strings"

// SplitChunks splits a multi-game PGN text into per-game substrings so each
// chunk can be stored as a self-contained pgn_raw blob. Splits on blank-line
// boundaries preceding "[Event " tag blocks.
//
// expected is the number of games parsed from the same text. If the split
// count does not match expected, the function returns a slice of empty
// strings of length expected and callers should fall back to storing the
// full text for every game.
func SplitChunks(text string, expected int) []string {
	const sep = "\n[Event "
	var chunks []string
	idx := 0
	for {
		if idx == 0 {
			p := strings.Index(text, "[Event ")
			if p < 0 {
				break
			}
			idx = p
		}
		after := idx + 1
		var next int
		if p := strings.Index(text[after:], sep); p < 0 {
			next = -1
		} else {
			next = after + p + 1 // skip the leading '\n'
		}
		if next < 0 {
			chunks = append(chunks, text[idx:])
			break
		}
		chunks = append(chunks, text[idx:next])
		idx = next
	}
	if len(chunks) != expected {
		return make([]string, expected)
	}
	return chunks
}
