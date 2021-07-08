/*
 * FH - Far Horizons server
 * Copyright (c) 2021  Michael D Henderson
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package scanner

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Tokenizer struct {
	line, col int // one-based values
	b []byte
	pb []*Token
}
type Token struct {
	Line, Col int // one-based values
	Text string
}

func NewTokenizer(b []byte) *Tokenizer {
	return &Tokenizer{
		line: 1,
		col:  1,
		b:    b,
	}
}

func (t *Tokenizer) Next() *Token {
	if len(t.pb) != 0 {
		tok := t.pb[len(t.pb)-1]
		t.pb = t.pb[:len(t.pb)]
		return tok
	}

	// skip comments and spaces
	for len(t.b) != 0 {
		r, w := utf8.DecodeRune(t.b)

		// comma is a word terminator left over from name processing
		if r == ',' {
			t.b, t.col = t.b[w:], t.col+1
			continue
		}

		// comments consume all characters up to the end of the line
		if r == ';' {
			t.b, t.col = t.b[w:], t.col+1
			for len(t.b) != 0 {
				// comments never include the new-line
				if r, w = utf8.DecodeRune(t.b); r == '\n' {
					break
				}
				t.b, t.col = t.b[w:], t.col+1
			}
			continue
		}

		// spaces are ignored between words
		if unicode.IsSpace(r) && r != '\n' {
			t.b, t.col = t.b[w:], t.col+1
			continue
		}

		break
	}
	if len(t.b) == 0 {
		return nil
	}

	pos := Token{Line: t.line, Col: t.col}

	r, w := utf8.DecodeRune(t.b)
	if r == '\n' {
		t.b, t.line, t.col = t.b[w:], t.line + 1, 1
		return &Token{Line: pos.Line, Col: pos.Col, Text: "\n"}
	}

	// if we're here, we must have a command or an argument
	sb := &strings.Builder{}
	if r == utf8.RuneError {
		sb.WriteByte('?')
	} else {
		sb.WriteRune(r)
	}
	t.b, t.col = t.b[w:], t.col+1
	for len(t.b) != 0 {
		if r, w = utf8.DecodeRune(t.b); unicode.IsSpace(r) || r == ';' || r == ',' {
			break
		} else if r == utf8.RuneError {
			sb.WriteByte('?')
		} else {
			sb.WriteRune(r)
		}
		t.b, t.col = t.b[w:], t.col+1
	}
	if r == ',' {
		// must consume comma as a word terminator
		t.b, t.col = t.b[w:], t.col+1
	}

	word := strings.ToUpper(sb.String())

	// check for keywords. these are the "commands" for orders.
	for _, keyword := range []string{
		"START",
		"COMBAT", "PRE-DEPARTURE", "JUMPS", "PRODUCTION", "POST-ARRIVAL", "STRIKES",
		"END",
		"ALLY", "AMBUSH", "ATTACK", "AUTO",
		"BASE", "BATTLE", "BUILD",
		"CONTINUE",
		"DESTROY", "DEVELOP", "DISBAND",
		"ENEMY", "ENGAGE", "ESTIMATE",
		"HAVEN", "HIDE", "HIJACK",
		"IBUILD", "ICONTINUE", "INSTALL", "INTERCEPT",
		"JUMP",
		"LAND",
		"MESSAGE", "MOVE",
		"NAME", "NEUTRAL",
		"ORBIT",
		"PJUMP",
		"RECYCLE", "REPAIR", "RESEARCH",
		"SCAN", "SEND", "SHIPYARD", "START", "SUMMARY",
		"TARGET", "TEACH", "TELESCOPE", "TERRAFORM", "TRANSFER",
		"UNLOAD", "UPGRADE",
		"VISITED",
		"WITHDRAW", "WORMHOLE",
		"ZZZ",
	} {
		if word == keyword { // keywords are always forced to uppercase
			return &Token{Line: pos.Line, Col: pos.Col, Text: keyword}
		}
	}

	// space is not a delimiter for ship, colony, or species names.
	// any other delimiter forces end of word.
	if r != ' ' {
		return &Token{Line: pos.Line, Col: pos.Col, Text: sb.String()}
	}

	// check for colony or species names
	isName := word == "PL" || word == "SP"

	// check for ship names
	if !isName {
		for _, ship := range []string{"BC", "BCS", "BM", "BMS", "BR", "BRS", "BS", "BSS", "BW", "BWS", "CA", "CAS", "CC", "CCS", "CL", "CLS", "CS", "CSS", "CT", "CTS", "DD", "DDS", "DN", "DNS", "ES", "ESS", "FF", "FFS", "PB", "PBS", "PJUMP", "SD", "SDS", "TR", "TRS"} {
			if isName = word == ship; isName {
				break
			}
		}
		if !isName && strings.HasPrefix(word, "TR") {
			// transports must check for TRn and TRnS
			digits := word[2:]
			if len(digits) > 1 && digits[len(digits)-1] == 'S' {
				digits = digits[:len(digits)-1]
			}
			_, err := strconv.Atoi(digits)
			isName = err == nil
		}
	}

	if isName {
		// names are terminated by any delimiter except a space.
		for len(t.b) != 0 {
			if r, w = utf8.DecodeRune(t.b); r == ';' || r == ',' || r == '\t' || r == '\n' || r == '\r' {
				break
			} else if r == utf8.RuneError {
				sb.WriteByte('?')
			} else {
				sb.WriteRune(r)
			}
			t.b, t.col = t.b[w:], t.col+1
		}
		if r != ',' {
			return &Token{Line: pos.Line, Col: pos.Col, Text: strings.TrimSpace(sb.String())}
		}
	}

	return &Token{Line: pos.Line, Col: pos.Col, Text: sb.String()}
}

func (t *Tokenizer) Peek() *Token {
	if len(t.pb) == 0 {
		t.Push(t.Next())
	}
	return t.pb[len(t.pb)-1]
}

func (t *Tokenizer) Push(tok *Token) {
	t.pb = append(t.pb, tok)
}

