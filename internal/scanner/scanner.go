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
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type position struct {
	line   int // one based values from start of file
	col    int // one based values from start of line
	offset int // zero based offset into buffer
}

type ORDERS struct {
	sections []*SECTION
	err      error
}
type SECTION struct {
	line     int // one based values from start of file
	col      int // one based values from start of line
	Name     string
	commands []*COMMAND
	err      error
}
type COMMAND struct {
	line int // one based values from start of file
	col  int // one based values from start of line
	text []byte
	args []*ARG
	err  error
}
type ARG struct {
	line int // one based values from start of file
	col  int // one based values from start of line
	text []byte
	err  error
}

type token struct {
	line int // one based values from start of file
	col  int // one based values from start of line
	text []byte
}

type scanner struct {
	line   int // one based values from start of file
	col    int // one based values from start of line
	offset int // zero based offset into buffer
	buffer []byte
}

func (s scanner) eof() bool {
	return len(s.buffer) <= s.offset
}

// getch silently converts \r\n to \n
func (s scanner) getch() rune {
	r, w := utf8.DecodeRune(s.buffer[s.offset:])
	s.offset += w
	if r == '\r' {
		if nr, nw := utf8.DecodeRune(s.buffer[s.offset:]); nr == '\n' {
			r, w = nr, nw
		}
	}
	if r == '\n' {
		s.line, s.col = s.line+1, 0
	} else {
		s.col++
	}
	return r
}

// getComment will read the next comment.
// it will return nil if only there is no comment.
// it will read up to, but not including, eol or eof.
// cr and invalid UTF-8 characters in comments are ignored.
func (s *scanner) getComment() *token {
	r, w := utf8.DecodeRune(s.buffer[s.offset:])
	if r != ';' {
		return nil
	}
	line, col := s.line, s.col
	s.offset += w
	for !s.eof() {
		if r, w = utf8.DecodeRune(s.buffer[s.offset:]); r == '\n' {
			break
		}
		s.offset += w
	}
	return &token{line: line, col: col}
}

// getEOL will read the new new-line or carriage-return new-line.
// it will return nil if there's not a match.
func (s *scanner) getEOL() *token {
	r, w := utf8.DecodeRune(s.buffer[s.offset:])
	if r == '\r' {
		if nr, nw := utf8.DecodeRune(s.buffer[s.offset+w:]); nr == '\n' {
			r, w = nr, w+nw
		}
	}
	if r != '\n' {
		return nil
	}
	token := token{line: s.line, col: s.col}
	s.line, s.col, s.offset = s.line+1, 0, s.offset+w
	return &token
}

// getSpaces will read the next run of spaces.
// it will return nil only if there are no spaces.
// it will read up to, but not including, eol, eof, or the first non-space rune.
// invalid UTF-8 characters are considered non-space runes.
func (s *scanner) getSpaces() *token {
	line, col, offset := s.line, s.col, s.offset
	for r, w := utf8.DecodeRune(s.buffer[s.offset:]); r != '\n' && unicode.IsSpace(r); r, w = utf8.DecodeRune(s.buffer[s.offset:]) {
		s.offset, s.col = s.offset+w, s.col+1
	}
	if s.offset == offset {
		return nil
	}
	return &token{line: line, col: col}
}

// getWord will read the next word.
// it will return nil only if there is no word to read.
func (s *scanner) getWord() *token {
	line, col, offset := s.line, s.col, s.offset
	r, w := utf8.DecodeRune(s.buffer[s.offset:])
	if unicode.IsSpace(r) || r == ';' {
		return nil
	}
	s.offset += w
	for !s.eof() {
		if r, w = utf8.DecodeRune(s.buffer[s.offset:]); unicode.IsSpace(r) || r == ';' || r == ',' {
			break
		}
		if r == utf8.RuneError {
			// how to handle invalid utf-8 values?
		}
		s.offset, s.col = s.offset+w, s.col+1
	}
	word, lenWord := strings.ToUpper(string(s.buffer[offset:s.offset])), s.offset-offset
	if r == ',' {
		// must consume comma as a word terminator
		s.offset, s.col = s.offset+w, s.col+1
	}

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
		if word == keyword { // keywords are always returned uppercase
			return &token{line: line, col: col, text: []byte(keyword)}
		}
	}

	// space is not a delimiter for ship, colony, or species names.
	// any other delimiter forces end of word.
	if r != ' ' {
		return &token{line: line, col: col, text: s.buffer[offset : offset+lenWord]}
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

	if !isName {
		// what can this be? an integer?
		// doesn't matter, return it as is.
		return &token{line: line, col: col, text: s.buffer[offset:s.offset]}
	}

	// names are terminated by any delimiter except a space.
	for !s.eof() {
		if r, w = utf8.DecodeRune(s.buffer[s.offset:]); r == ';' || r == ',' || r == '\t' || r == '\n' || r == '\r' {
			break
		}
		s.offset, s.col = s.offset+w, s.col+1
	}

	// token must include the keyword for colony, ship, or species
	t := token{line: line, col: col, text: s.buffer[offset:s.offset]}
	if r == ',' { // must consume comma as a word terminator
		s.offset, s.col = s.offset+w, s.col+1
	}

	return &t
}

func (s *scanner) skipWhitespace() {
	for s.getSpaces() != nil || s.getComment() != nil {
		//
	}
}

//func Scan(b []byte) (ORDERS, error) {
//	var orders ORDERS
//	var section *SECTION
//	var command *COMMAND
//	var arg *ARG
//	s := &scanner{line: 1, col: 1, buffer: b}
//	// orders: section* EOF.
//	// section: START SECTION-NAME command* END.
//	// command: COMMAND ARG* EOL.
//	for !s.eof() {
//	}
//	if i != len(b) {
//		return orders, fmt.Errorf("scanner: %d:%d: halted early", pos.line, pos.col)
//	}
//	return orders, nil
//}

func acceptSection(s *scanner) (*SECTION, error) {
	for s.getSpaces() != nil || s.getComment() != nil || s.getEOL() != nil {
		continue
	}
	if s.eof() {
		return nil, nil
	}
	word := s.getWord()
	if word == nil || !bytes.Equal(word.text, []byte{'S', 'T', 'A', 'R', 'T'}) {
		// ignore words outside of section
		return nil, nil
	}

	section := &SECTION{line: word.line, col: word.col}

	for s.getSpaces() != nil && s.getComment() == nil {
		//
	}
	name := s.getWord()
	if name == nil {
		section.err = fmt.Errorf("%d: expected section name: got nothing", section.line)
		// recover by scanning to end of line
		for !s.eof() && s.getEOL() == nil {
			s.getSpaces()
			s.getComment()
			s.getWord()
		}
		return section, nil
	}
	section.Name = string(name.text)
	switch section.Name {
	case "COMBAT":
	case "JUMPS":
	case "PRE-DEPARTURE":
	case "PRODUCTION":
	case "POST-ARRIVAL":
	case "STRIKES":
		section.err = fmt.Errorf("%d: expected section name: got %q", section.line, string(name.text))
		// recover by scanning to end of line
		for !s.eof() && s.getEOL() == nil {
			s.getSpaces()
			s.getComment()
			s.getWord()
		}
		return section, nil
	}
	// ignore everything following the section command up to the end of the line
	for !s.eof() && s.getEOL() == nil {
		s.getSpaces()
		s.getComment()
		s.getWord()
	}

	var command *token
	for !s.eof() {
		if s.getSpaces() != nil || s.getComment() != nil || s.getEOL() != nil {
			continue
		}
		command = s.getWord()
		if bytes.Equal(command.text, []byte{'E', 'N', 'D'}) {
			// ignore everything following the END command up to the end of the line
			for !s.eof() && s.getEOL() == nil {
				s.getSpaces()
				s.getComment()
				s.getWord()
			}
			return section, nil
		}
	}

	section.err = fmt.Errorf("%d: found end-of-input before END", s.line)
	return section, nil
}

func New(b []byte) Scanner {
	s := Scanner{buffer: b}
	for _, line := range bytes.Split(b, []byte{'\n'}) {
		if comment := bytes.IndexByte(line, ';'); comment != -1 {
			line = line[:comment]
		}
		if cr := bytes.IndexByte(line, '\r'); cr != -1 {
			line = line[:cr]
		}
		line = bytes.Trim(line, " \t")
		for {
			nl := line
			nl = bytes.ReplaceAll(nl, []byte{'\t', ' '}, []byte{'\t'})
			nl = bytes.ReplaceAll(nl, []byte{' ', '\t'}, []byte{'\t'})
			nl = bytes.ReplaceAll(nl, []byte{'\t', '\t'}, []byte{'\t'})
			if len(nl) == len(line) {
				break
			}
			line = nl
		}
		s.lines = append(s.lines, string(line))
	}
	return s
}

type Lexeme struct {
	Kind    Kind
	Integer int
	Text    string
}

type Scanner struct {
	buffer []byte
	lines  []string
}

type Kind int

const (
	Unknown Kind = iota
	Colony
	Command
	EOF
	EOL
	Integer
	Item
	Ship
	Spaces
	Species
	Start
	Text
)

func (s Scanner) Accept(x string) (Lexeme, Scanner) {
	switch x {
	case "blank line":
		return s.acceptBlankLine()
	case "eof":
		return s.acceptEOF()
	case "eol":
		return s.acceptEOL()
	case "spaces":
		return s.acceptSpaces()
	}
	panic("!")
}

func (s Scanner) Expect(x string) (Lexeme, Scanner, error) {
	lexeme, rest := s.Accept(x)
	if lexeme.Kind == Unknown {

		return lexeme, rest, nil
	}
	panic("!")
}

func (s Scanner) acceptBlankLine() (Lexeme, Scanner) {
	if s.eof() {
		return Lexeme{}, s
	}
	bb := &nab{b: s.buffer}
	bb.runOf([]byte{' ', '\t'})
	if bb.peek() == ';' {
		bb.runTo([]byte{'\n'})
	}
	if bb.peek() == '\r' {
		bb.next()
	}
	if bb.peek() != '\n' {
		if bb.o == len(bb.b) {
			// last line had no end-of-line character
		} else {
			return Lexeme{}, s
		}
	}
	return Lexeme{Kind: Text, Text: string(s.buffer[:bb.o])}, Scanner{buffer: s.buffer[bb.o:]}
}

func (s Scanner) acceptNamedEntity() (Lexeme, Scanner) {
	if len(s.buffer) < 3 {
		return Lexeme{}, s
	}
	bb, k := &nab{b: s.buffer}, Unknown
	switch bb.nextToUpper() {
	case 'B':
		switch bb.nextToUpper() {
		case 'C' /* Battlecruiser */, 'M' /* Battlemoon */, 'R' /* Battlestar */, 'S' /* Battleship */, 'W' /* Battleworld */ :
			switch bb.nextToUpper() {
			case 'S': // sublight
				switch bb.next() {
				case ' ':
					k = Ship
				}
			case ' ':
				k = Ship
			}
		}
	case 'C':
		switch bb.nextToUpper() {
		case 'A' /* Heavy Cruiser */, 'C' /* Command Cruiser */, 'L' /* Light Cruiser */, 'S' /* Strike Cruiser */, 'T' /* Corvette */ :
			switch bb.nextToUpper() {
			case 'S': // sublight
				switch bb.next() {
				case ' ':
					k = Ship
				}
			case ' ':
				k = Ship
			}
		}
	case 'D':
		switch bb.nextToUpper() {
		case 'D' /* Destroyer */, 'N' /* Dreadnought */ :
			switch bb.nextToUpper() {
			case 'S': // sublight
				switch bb.next() {
				case ' ':
					k = Ship
				}
			case ' ':
				k = Ship
			}
		}
	case 'E':
		switch bb.nextToUpper() {
		case 'S': // Escort
			switch bb.nextToUpper() {
			case 'S': // sublight
				switch bb.next() {
				case ' ':
					k = Ship
				}
			case ' ':
				k = Ship
			}
		}
	case 'F':
		switch bb.nextToUpper() {
		case 'F': // Frigate
			switch bb.nextToUpper() {
			case 'S': // sublight
				switch bb.next() {
				case ' ':
					k = Ship
				}
			case ' ':
				k = Ship
			}
		}
	case 'P':
		switch bb.nextToUpper() {
		case 'B': // Picketboat
			switch bb.nextToUpper() {
			case 'S': // sublight
				switch bb.next() {
				case ' ':
					k = Ship
				}
			case ' ':
				k = Ship
			}
		case 'L': // Planet (Colony)
			if bb.next() == ' ' {
				k = Colony
			}
		}
	case 'S':
		switch bb.nextToUpper() {
		case 'D': // Super Dreadnought
			switch bb.nextToUpper() {
			case 'S': // sublight
				switch bb.next() {
				case ' ':
					k = Ship
				}
			case ' ':
				k = Ship
			}
		case 'P': // Species
			if bb.next() == ' ' {
				k = Species
			}
		}
	case 'T':
		switch bb.nextToUpper() {
		case 'R': // Transport
			digits := bb.runOf([]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'})
			if len(digits) != 0 {
				switch bb.nextToUpper() {
				case 'S': // sublight
					switch bb.next() {
					case ' ':
						k = Ship
					}
				case ' ':
					k = Ship
				}
			}
		}
	}

	if k == Unknown {
		return Lexeme{}, s
	}
	if spaces := bb.runOf([]byte{' '}); len(spaces) != 1 {
		// must be exactly one space following the type
		return Lexeme{}, s
	}

	// the entity name is all the characters following up to a terminating tab, comma, semi, or EOL
	name := bb.runTo([]byte{'\t', ',', ';', '\r', '\n'})
	if len(name) == 0 {
		// no entity name following the type?
		return Lexeme{}, s
	}
	return Lexeme{Kind: k, Text: string(name)}, Scanner{buffer: s.buffer[bb.o:]}
}

func (s Scanner) acceptEOF() (Lexeme, Scanner) {
	if !s.eof() {
		return Lexeme{}, s
	}
	return Lexeme{Kind: EOF}, s
}

func (s Scanner) acceptEOL() (Lexeme, Scanner) {
	if s.eof() {
		return Lexeme{}, s
	} else if s.buffer[0] == '\n' {
		return Lexeme{Kind: EOL}, Scanner{buffer: s.buffer[1:]}
	} else if len(s.buffer) > 1 && s.buffer[0] == '\r' && s.buffer[1] == '\n' {
		return Lexeme{Kind: EOL}, Scanner{buffer: s.buffer[2:]}
	}
	return Lexeme{}, s
}

func (s Scanner) acceptInt() (Lexeme, Scanner) {
	n := s.runOf([]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'})
	if len(n) == 0 {
		return Lexeme{}, s
	} else if len(s.buffer) > len(n) {
		if bytes.IndexByte([]byte{' ', '\t', '\r', '\n', ';'}, s.buffer[len(n)]) == -1 {
			return Lexeme{}, s
		}
	}
	i, err := strconv.Atoi(string(s.buffer[:len(n)]))
	if err != nil {
		return Lexeme{}, s
	}
	return Lexeme{Kind: Integer, Integer: i}, Scanner{buffer: s.buffer[len(n):]}
}

func (s Scanner) acceptSpaces() (Lexeme, Scanner) {
	if n := s.runOf([]byte{' ', '\t'}); len(n) != 0 {
		return Lexeme{Kind: Spaces, Text: string(s.buffer[:len(n)])}, Scanner{buffer: s.buffer[len(n):]}
	}
	return Lexeme{}, s
}

func (s Scanner) acceptStart() (Lexeme, Scanner) {
	if len(s.buffer) < 6 || !s.hasPrefix("start") {
		return Lexeme{}, s
	} else if bytes.IndexByte([]byte{' ', '\t'}, s.buffer[5]) == -1 {
		return Lexeme{}, s
	}
	return Lexeme{Kind: Start}, Scanner{buffer: s.buffer[5:]}
}

func (s Scanner) eof() bool {
	return len(s.buffer) == 0
}

func (s Scanner) eol() bool {
	switch len(s.buffer) {
	case 0:
		return false
	case 1:
		return s.buffer[0] == '\n'
	}
	return s.buffer[0] == '\r' && s.buffer[1] == '\n'
}

func (s Scanner) hasPrefix(t string) bool {
	return len(s.buffer) > len(t) && bytes.EqualFold(s.buffer[:len(t)], []byte(t))
}

func (s Scanner) runOf(set []byte) []byte {
	n := 0
	for n < len(s.buffer) && bytes.IndexByte(set, s.buffer[n]) != -1 {
		n++
	}
	if n == 0 {
		return nil
	}
	return s.buffer[:n]
}
