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
	"bytes"
	"unicode"
	"unicode/utf8"
)

type nab struct {
	b []byte
	o int
}

func (n *nab) next() rune {
	if !(n.o < len(n.b)) {
		return utf8.RuneError
	} else if r, w := utf8.DecodeRune(n.b[n.o:]); r == utf8.RuneError {
		return r
	} else {
		n.o += w
		return r
	}
}

func (n *nab) nextToUpper() rune {
	if !(n.o < len(n.b)) {
		return utf8.RuneError
	} else if r, w := utf8.DecodeRune(n.b[n.o:]); r == utf8.RuneError {
		return r
	} else {
		n.o += w
		return unicode.ToUpper(r)
	}
}

func (n *nab) peek() rune {
	if !(n.o < len(n.b)) {
		return utf8.RuneError
	}
	r, _ := utf8.DecodeRune(n.b[n.o:])
	return r
}

func (n *nab) peekToUpper() rune {
	if !(n.o < len(n.b)) {
		return utf8.RuneError
	} else if r, _ := utf8.DecodeRune(n.b[n.o:]); r == utf8.RuneError {
		return r
	} else {
		return unicode.ToUpper(r)
	}
}

func (n *nab) runOf(set []byte) []byte {
	start := n.o
	for n.o < len(n.b) && bytes.IndexByte(set, n.b[n.o]) != -1 {
		n.o++
	}
	if start == n.o {
		return nil
	}
	return n.b[start:n.o]
}

func (n *nab) runTo(delim []byte) []byte {
	start := n.o
	for n.o < len(n.b) && bytes.IndexByte(delim, n.b[n.o]) == -1 {
		n.o++
	}
	if start == n.o {
		return nil
	}
	return n.b[start:n.o]
}

func (n *nab) unget() {
	if n.o > 0 {
		n.o--
	}
}
