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

package orders

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Scanner struct {
	line, col int // one-based values
	b []byte
	pb []*Token
}

type Token struct {
	Line, Col int // one-based values
	Text string
}

func NewScanner(b []byte) *Scanner {
	return &Scanner{
		line: 1,
		col:  1,
		b:    b,
	}
}

func (s *Scanner) Next() *Token {
	if len(s.pb) != 0 {
		tok := s.pb[len(s.pb)-1]
		s.pb = s.pb[:len(s.pb)]
		return tok
	}

	// skip comments and spaces
	for len(s.b) != 0 {
		r, w := utf8.DecodeRune(s.b)

		// comma is a word terminator left over from name processing
		if r == ',' {
			s.b, s.col = s.b[w:], s.col+1
			continue
		}

		// comments consume all characters up to the end of the line
		if r == ';' {
			s.b, s.col = s.b[w:], s.col+1
			for len(s.b) != 0 {
				// comments never include the new-line
				if r, w = utf8.DecodeRune(s.b); r == '\n' {
					break
				}
				s.b, s.col = s.b[w:], s.col+1
			}
			continue
		}

		// spaces are ignored between words
		if unicode.IsSpace(r) && r != '\n' {
			s.b, s.col = s.b[w:], s.col+1
			continue
		}

		break
	}
	if len(s.b) == 0 {
		return nil
	}

	pos := Token{Line: s.line, Col: s.col}

	r, w := utf8.DecodeRune(s.b)
	if r == '\n' {
		s.b, s.line, s.col = s.b[w:], s.line + 1, 1
		return &Token{Line: pos.Line, Col: pos.Col, Text: "\n"}
	}

	// if we're here, we must have a command or an argument
	sb := &strings.Builder{}
	if r == utf8.RuneError {
		sb.WriteByte('?')
	} else {
		sb.WriteRune(r)
	}
	s.b, s.col = s.b[w:], s.col+1
	for len(s.b) != 0 {
		if r, w = utf8.DecodeRune(s.b); unicode.IsSpace(r) || r == ';' || r == ',' {
			break
		} else if r == utf8.RuneError {
			sb.WriteByte('?')
		} else {
			sb.WriteRune(r)
		}
		s.b, s.col = s.b[w:], s.col+1
	}
	if r == ',' {
		// must consume comma as a word terminator
		s.b, s.col = s.b[w:], s.col+1
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
		for len(s.b) != 0 {
			if r, w = utf8.DecodeRune(s.b); r == ';' || r == ',' || r == '\t' || r == '\n' || r == '\r' {
				break
			} else if r == utf8.RuneError {
				sb.WriteByte('?')
			} else {
				sb.WriteRune(r)
			}
			s.b, s.col = s.b[w:], s.col+1
		}
		if r != ',' {
			return &Token{Line: pos.Line, Col: pos.Col, Text: strings.TrimSpace(sb.String())}
		}
	}

	return &Token{Line: pos.Line, Col: pos.Col, Text: sb.String()}
}

func (s *Scanner) Peek() *Token {
	if len(s.pb) == 0 {
		s.Push(s.Next())
	}
	return s.pb[len(s.pb)-1]
}

func (s *Scanner) Push(tok *Token) {
	s.pb = append(s.pb, tok)
}


