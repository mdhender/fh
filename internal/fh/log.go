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

package fh

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"
)

type Logger struct {
	Disabled    bool
	Indentation int
	Position    int
	StartOfLine bool
	Line        []byte
	File        io.Writer
	Stdout      io.Writer
	Summary     io.Writer
}

func (l *Logger) Char(ch byte) {
	if l.Disabled {
		return
	}

	if l.Line == nil {
		l.Line = make([]byte, 128, 128)
	}

	// check if current line is getting too long
	if (ch == ' ' || ch == '\n') && l.Position > 77 {
		// find closest preceding space
		temp_position := l.Position - 1
		for temp_position >= 0 && l.Line[l.Position] != ' ' {
			temp_position--
		}
		if temp_position == -1 {
			// no spaces!
			temp_position = l.Position
		}
		front, rest := l.Line[:temp_position], l.Line[temp_position:]

		// write front of line to files
		l.Puts(front)

		// copy overflow word to beginning of next line
		l.Line = rest
		l.Position = l.Indentation + 2 // why do we add 2 here?
		for i := 0; i < l.Position; i++ {
			l.Line[i] = ' '
		}
		copy(l.Line[l.Position:], rest)
		l.Position += len(rest)

		if ch == ' ' {
			l.Line[l.Position] = ch
			l.Position++
			return
		}
	}

	// check if line is being manually terminated
	if ch == '\n' {
		// write current line to output
		l.Puts(l.Line[:l.Position])

		// set up for next line
		l.Position = 0
		l.Indentation = 0
		l.StartOfLine = true

		return
	}

	// save this character
	l.Line[l.Position] = ch
	l.Position++

	l.StartOfLine = l.StartOfLine && ch == ' '
	if l.StartOfLine {
		// save number of indenting spaces for current line
		l.Indentation++
	}
}

func (l *Logger) Int(value int) {
	if l.Disabled {
		return
	}
	l.String(fmt.Sprintf("%d", value))
}

func (l *Logger) Long(value int) {
	l.Int(value)
}

func (l *Logger) Message(msg string) {
	if l.File == nil {
		return
	}
	if _, err := l.File.Write([]byte(msg)); err != nil {
		panic(err)
	}
}

func (l *Logger) Puts(line []byte) {
	nl := []byte{'\n'}
	if l.File != nil {
		if _, err := l.File.Write(line); err != nil {
			panic(err)
		}
		if _, err := l.File.Write(nl); err != nil {
			panic(err)
		}
	}
	if l.Stdout != nil {
		if _, err := l.Stdout.Write(line); err != nil {
			panic(err)
		}
		if _, err := l.Stdout.Write(nl); err != nil {
			panic(err)
		}
	}
	if l.Summary != nil {
		if _, err := l.Summary.Write(line); err != nil {
			panic(err)
		}
		if _, err := l.Summary.Write(nl); err != nil {
			panic(err)
		}
	}
}

func (l *Logger) String(s string) {
	if l.Disabled {
		return
	}

	for _, ch := range []byte(s) {
		l.Char(ch)
	}
}

func GetMessage(galaxyPath string, n int) (string, error) {
	name := path.Join(galaxyPath, fmt.Sprintf("m%06d.msg", n))
	// open message file
	data, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Printf("\n\tWARNING! utils.c: cannot open message file '%s'!\n\n", name)
		return "", err
	}
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		fmt.Printf("\n\tWARNING! message file %q: %+v\n\n", name, err)
		return "", err
	}
	return s, nil
}
