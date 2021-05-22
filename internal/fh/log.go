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
	"os"
	"path/filepath"
)

type Logger struct {
	Disabled bool
	File     io.WriteCloser
	Stdout   io.WriteCloser
	Summary  io.WriteCloser
}

func (l *Logger) Char(ch byte) {
	l.Write([]byte{ch})
}

func (l *Logger) Close() {
	if l.File != nil {
		l.File.Close()
	}
	if l.Stdout != nil && l.Stdout != os.Stdout {
		l.Stdout.Close()
	}
	if l.Summary != nil {
		l.Summary.Close()
	}
}

func (l *Logger) Int(value int) {
	l.Printf("%d", value)
}

func (l *Logger) Long(value int) {
	l.Printf("%d", value)
}

func (l *Logger) Message(msg string) {
	if l.File == nil {
		return
	}
	if _, err := l.File.Write([]byte(msg)); err != nil {
		panic(err)
	}
}

func (l *Logger) Printf(format string, a ...interface{}) {
	l.Write([]byte(fmt.Sprintf(format, a...)))
}

func (l *Logger) Write(b []byte) {
	if l.Disabled || len(b) == 0 {
		return
	}
	if l.File != nil {
		if _, err := l.File.Write(b); err != nil {
			panic(err)
		}
	}
	if l.Stdout != nil {
		if _, err := l.Stdout.Write(b); err != nil {
			panic(err)
		}
	}
	if l.Summary != nil {
		if _, err := l.Summary.Write(b); err != nil {
			panic(err)
		}
	}
}

func (l *Logger) String(s string) {
	l.Write([]byte(s))
}

func GetMessage(galaxyPath string, n int) (string, error) {
	name := filepath.Join(galaxyPath, fmt.Sprintf("m%06d.msg", n))
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
