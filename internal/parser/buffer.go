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

package parser

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type buffer struct {
	s       string
	line, o int
	b       []byte
	pb      []rune
}

func (b *buffer) snark() string {
	var s string
	for o := len(b.pb) - 1; len(s) < 10 && o >= 0; o-- {
		s += string(b.pb[o])
	}
	if b.o < len(b.b) {
		t := b.b[b.o:]
		for len(t) > 0 && len(s) < 10 {
			r, w := utf8.DecodeRune(t)
			if r == utf8.RuneError {
				break
			}
			s += string(t[:w])
			t = t[w:]
		}
	}
	return s
}
func (b *buffer) clone() *buffer {
	return &buffer{b: bdup(b.b), line: b.line, o: b.o, pb: rdup(b.pb), s: b.s}
}

func (b *buffer) eof() bool {
	return len(b.pb) == 0 && b.o == len(b.b)
}

func (b *buffer) get() rune {
	if len(b.pb) != 0 {
		r := b.pb[len(b.pb)-1]
		b.pb = b.pb[:len(b.pb)-1]
		b.s = b.snark()
		return r
	} else if !(b.o < len(b.b)) {
		return utf8.RuneError
	} else if r, w := utf8.DecodeRune(b.b[b.o:]); r == utf8.RuneError {
		b.s = b.snark()
		return r
	} else {
		b.o += w
		b.s = b.snark()
		return r
	}
}

func (b *buffer) hasPrefix(pfx string) bool {
	n := 0
	for _, p := range pfx {
		if r := b.peek(); unicode.ToUpper(r) == unicode.ToUpper(p) {
			b.get()
			n++
		}
	}
	if n == len(pfx) {
		switch b.peek() {
		case ' ', '\t', ';', '\r', '\n':
			return true
		}
	}
	return false
}

func (b *buffer) peek() rune {
	if len(b.pb) != 0 {
		return b.pb[len(b.pb)-1]
	} else if !(b.o < len(b.b)) {
		return utf8.RuneError
	}
	r, _ := utf8.DecodeRune(b.b[b.o:])
	return r
}

func (b *buffer) runOf(set string) string {
	sb := &strings.Builder{}
	r := b.peek()
	for strings.IndexRune(set, r) != -1 {
		sb.WriteRune(r)
		b.get()
		r = b.peek()
	}
	return sb.String()
}

func (b *buffer) runTo(delim string) string {
	sb := &strings.Builder{}
	r := b.peek()
	for strings.IndexRune(delim, r) == -1 {
		sb.WriteRune(r)
		b.get()
		r = b.peek()
	}
	return sb.String()
}

func (b *buffer) unget(r rune) {
	b.pb = append(b.pb, r)
}
