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

import "bytes"

type Scanner struct {
	line, col int
	buffer []byte
}

type Kind int
const (
	EOF Kind = iota
	NL
	Spaces
)

type Token struct {
	Line, Col int
	K Kind
	V []byte
}

func (s Scanner) IsEOF() bool {
	return len(s.buffer) == 0
}

func (s Scanner) NL() (*Token, Scanner, error) {
	if bytes.HasPrefix(s.buffer, []byte{'\n'}) {
		return &Token{Line:s.line, Col: s.col, K: NL, V: []byte{'\n'}}, Scanner{line: s.line+1, col: 0, buffer:s.buffer[1:]}, nil
	} else if bytes.HasPrefix(s.buffer, []byte{'\r','\n'}) {
		return &Token{Line:s.line, Col: s.col, K: NL, V: []byte{'\n'}}, Scanner{line: s.line+1, col: 0, buffer:s.buffer[2:]}, nil
	}
	return nil, s, nil
}