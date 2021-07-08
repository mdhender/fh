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

package lexer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Lex(name string) (*Lexer, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	l := &Lexer{line: 1, buf: b}

	for lexeme := l.read(); lexeme.Kind != EOF; lexeme = l.read() {
		fmt.Println(lexeme.Line, lexeme.Kind.String(), lexeme.Text)
		if lexeme.Kind == EOL {
			if len(l.lexemes) == 0 {
				continue // ignore leading blank lines
			} else if l.lexemes[len(l.lexemes)-1].Kind == EOL {
				continue // ignore consecutive blank lines
			}
		}
		l.lexemes = append(l.lexemes, lexeme)
	}

	// append an EOL if there was none on the last line of the input
	if len(l.lexemes) != 0 && l.lexemes[len(l.lexemes)-1].Kind != EOL {
		l.lexemes = append(l.lexemes, &Lexeme{Line: l.lexemes[len(l.lexemes)-1].Line, Kind: EOL})
	}
	l.eof = &Lexeme{Line: l.line, Kind: EOF}

	col := 1
	for _, lexeme := range l.lexemes {
		lexeme.ArgNo = col
		if lexeme.Kind == EOL {
			col = 0
		}
		col++
	}

	// free up the buffer
	l.buf = nil

	return l, nil
}

type Lexer struct {
	line, next int
	buf        []byte
	lexemes    []*Lexeme
	eof        *Lexeme
}

func (l *Lexer) Next() *Lexeme {
	if len(l.lexemes) <= l.next {
		return l.eof
	}
	lexeme := l.lexemes[l.next]
	l.next++
	return lexeme
}

func (l *Lexer) Peek() *Lexeme {
	if len(l.lexemes) <= l.next {
		return l.eof
	}
	return l.lexemes[l.next]
}

func (l *Lexer) Save() *Lexer {
	return &Lexer{line: l.line, next: l.next, buf: l.buf, lexemes: l.lexemes, eof: l.eof}
}

func (l *Lexer) read() *Lexeme {
	for l.skipSpaces() || l.skipComments() {
		// ignore spaces and comments
	}
	if l.iseof() {
		return &Lexeme{Line: l.line, Kind: EOF}
	}

	r, w := utf8.DecodeRune(l.buf)
	if r == '\n' {
		lexeme := &Lexeme{Kind: EOL, Line: l.line}
		l.line, l.buf = l.line+1, l.buf[w:]
		return lexeme
	}

	if t := "ALLY"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Ally}
	} else if t = "AMBUSH"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Ambush}
	} else if t = "ATTACK"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Attack}
	} else if t = "AUTO"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Auto}
	} else if t = "BASE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Base}
	} else if t = "BATTLE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Battle}
	} else if t = "BC"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "BCS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "BM"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "BMS"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: t}
	} else if t = "BR"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "BRS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "BS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "BSS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "BW"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "BWS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "BUILD"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Build}
	} else if t = "CA"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "CAS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "CC"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "CCS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "CL"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "CLS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "CS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "CSS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "CT"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "CTS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "COMBAT"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Combat}
	} else if t = "CONTINUE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Continue}
	} else if t = "DD"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "DDS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "DN"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "DNS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "DESTROY"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Destroy}
	} else if t = "DEVELOP"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Develop}
	} else if t = "DISBAND"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Disband}
	} else if t = "END"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: End}
	} else if t = "ENEMY"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Enemy}
	} else if t = "ENGAGE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Engage}
	} else if t = "ES"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "ESS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "ESTIMATE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Estimate}
	} else if t = "FF"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "FFS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "HAVEN"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Haven}
	} else if t = "HIDE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Hide}
	} else if t = "HIJACK"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Hijack}
	} else if t = "IBUILD"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: IBuild}
	} else if t = "ICONTINUE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: IContinue}
	} else if t = "INSTALL"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Install}
	} else if t = "INTERCEPT"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Intercept}
	} else if t = "JUMP"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Jump}
	} else if t = "JUMPS"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Jumps}
	} else if t = "LAND"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Land}
	} else if t = "MESSAGE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Message}
	} else if t = "MOVE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Move}
	} else if t = "NAME"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Name}
	} else if t = "NEUTRAL"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Neutral}
	} else if t = "ORBIT"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Orbit}
	} else if t = "PB"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "PBS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "PL"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Colony, Text: namu}
	} else if t = "PJUMP"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: PJump}
	} else if t = "POST-ARRIVAL"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: PostArrival}
	} else if t = "PRE-DEPARTURE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: PreDeparture}
	} else if t = "PRODUCTION"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Production}
	} else if t = "RECYCLE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Recycle}
	} else if t = "REPAIR"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Repair}
	} else if t = "RESEARCH"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Research}
	} else if t = "SCAN"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Scan}
	} else if t = "SD"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "SDS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "SEND"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Send}
	} else if t = "SHIPYARD"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Shipyard}
	} else if t = "SP"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Species, Text: namu}
	} else if t = "START"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Start}
	} else if t = "STRIKES"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Strikes}
	} else if t = "SUMMARY"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Summary}
	} else if t = "TARGET"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Target}
	} else if t = "TEACH"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Teach}
	} else if t = "TELESCOPE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Telescope}
	} else if t = "TERRAFORM"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Terraform}
	} else if t = "TR"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: Ship, Text: namu}
	} else if t = "TRS"; hasPrefix(l.buf, t) {
		n, namu := getName(l.buf[len(t):])
		l.buf = l.buf[len(t)+n:]
		return &Lexeme{Line: l.line, Kind: SublightShip, Text: namu}
	} else if t = "TRANSFER"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Transfer}
	} else if t = "UNLOAD"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Unload}
	} else if t = "UPGRADE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Upgrade}
	} else if t = "VISITED"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Visited}
	} else if t = "WITHDRAW"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Withdraw}
	} else if t = "WORMHOLE"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: Wormhole}
	} else if t = "ZZZ"; hasPrefix(l.buf, t) {
		l.buf = l.buf[len(t):]
		return &Lexeme{Line: l.line, Kind: ZZZ}
	}

	if len(l.buf) > 3 && (bytes.HasPrefix(l.buf, []byte{'T', 'R'}) || bytes.HasPrefix(l.buf, []byte{'T', 'r'}) || bytes.HasPrefix(l.buf, []byte{'t', 'R'}) || bytes.HasPrefix(l.buf, []byte{'t', 'r'})) {
		rest := l.buf[2:]
		if digits := runOf(rest, []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}); digits != 0 {
			if len(rest) > digits+3 { // 3 to include space plus name
				k := Ship
				r, w := utf8.DecodeRune(rest[digits:])
				if r == 'S' || r == 's' {
					k = SublightShip
					digits += w
				}
				n, namu := getName(rest[digits:])
				if n != 0 {
					l.buf = rest[digits+n:]
					return &Lexeme{Line: l.line, Kind: k, Text: namu}
				}
			}
		}
	}

	if digits := runOf(l.buf, []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}); digits != 0 {
		rest := l.buf[digits:]
		if len(rest) != 0 {
			r, _ := utf8.DecodeRune(rest)
			if r == ';' || unicode.IsSpace(r) {
				i, err := strconv.Atoi(string(l.buf[:digits]))
				if err == nil {
					l.buf = rest
					return &Lexeme{Line: l.line, Kind: Integer, Integer: i}
				}
			}
		}
	}

	sb := &strings.Builder{}
	lex := &Lexeme{Line: l.line, Kind: Text}
	for !l.iseof() {
		r, w := utf8.DecodeRune(l.buf)
		if r == ',' {
			l.buf = l.buf[w:]
			break
		} else if r == ';' || r == '\t' || r == '\r' || r == '\n' {
			break
		} else if r == utf8.RuneError {
			sb.WriteByte('?')
		} else {
			sb.WriteRune(r)
		}
		l.buf = l.buf[w:]
	}
	lex.Text = sb.String()

	return lex
}

func (l *Lexer) iseof() bool {
	return len(l.buf) == 0
}

func (l *Lexer) get() rune {
	if l.iseof() {
		return utf8.RuneError
	}
	r, w := utf8.DecodeRune(l.buf)
	if w > 0 {
		l.buf = l.buf[w:]
	}
	return r
}

func (l *Lexer) peek() rune {
	r, _ := utf8.DecodeRune(l.buf)
	return r
}

// skipComments will read past comments.
// it will read up to, but including, eol or eof.
// invalid UTF-8 characters in comments are ignored.
func (l *Lexer) skipComments() bool {
	if len(l.buf) > 0 && l.buf[0] == ';' {
		for len(l.buf) != 0 && l.buf[0] != '\n' {
			l.buf = l.buf[1:]
		}
		return true
	}
	return false
}

// skipSpaces will read past unicode spaces and comments.
// it will return at eof, eol, or the first non-space rune.
// invalid UTF-8 characters are considered non-space runes
func (l *Lexer) skipSpaces() bool {
	r, w := utf8.DecodeRune(l.buf)
	if r != '\n' && unicode.IsSpace(r) {
		l.buf = l.buf[w:]
		for len(l.buf) != 0 {
			r, w = utf8.DecodeRune(l.buf)
			if r == '\n' || !unicode.IsSpace(r) {
				break
			}
			l.buf = l.buf[w:]
		}
		return true
	}
	return false
}

func getName(s []byte) (n int, name string) {
	if len(s) == 0 || s[0] != ' ' {
		return 0, ""
	}
	// Something like "TR1 Freddy      ; comment" should return "Freddy"
	// but ""TR1 Freddy      ," should return "Freddy      ". The trimSpaces
	// variable controls that.
	trimSpaces := false

	s, n = s[1:], n+1
	sb := &strings.Builder{}
	for len(s) != 0 {
		r, w := utf8.DecodeRune(s)
		if r == ',' {
			s, n = s[1:], n+1
			break
		} else if r == ';' || r == '\t' || r == '\r' || r == '\n' {
			trimSpaces = true
			break
		} else if r == utf8.RuneError {
			sb.WriteByte('?')
		} else {
			sb.WriteRune(r)
		}
		s, n = s[w:], n+w
	}
	if trimSpaces {
		return n, strings.TrimRight(sb.String(), " ")
	}
	return n, sb.String()
}

func hasPrefix(s []byte, pfx string) bool {
	for _, p := range pfx {
		r, w := utf8.DecodeRune(s)
		if unicode.ToUpper(r) != unicode.ToUpper(p) {
			return false
		}
		s = s[w:]
	}
	r, _ := utf8.DecodeRune(s)
	return unicode.IsSpace(r) || r == ';'
}

func runOf(s, set []byte) int {
	n := 0
	for len(s) != 0 && bytes.IndexByte(set, s[0]) != -1 {
		s, n = s[1:], n+1
	}
	return n
}

func runTo(s, delim []byte) int {
	n := 0
	for len(s) != 0 && bytes.IndexByte(delim, s[0]) == -1 {
		s, n = s[1:], n+1
	}
	return n
}

type Line struct {
	line    int
	pos     int
	lexemes []*Lexeme
}

func (l *Line) Clone() *Line {
	return &Line{line: l.line, pos: l.pos, lexemes: l.lexemes}
}

func (l *Line) Peek() *Lexeme {
	if l.pos < len(l.lexemes) {
		return l.lexemes[l.pos]
	}
	return &Lexeme{Line: l.line, Kind: EOL}
}

func (l *Line) Pop() *Lexeme {
	if l.pos < len(l.lexemes) {
		l.pos++
		return l.lexemes[l.pos-1]
	}
	return &Lexeme{Line: l.line, Kind: EOL}
}
