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

package logger

import (
	"fmt"
	"io"
)

type Logger struct {
	Disabled bool
	File     io.WriteCloser
}

func (w *Logger) Close() {
	if w.File == nil {
		return
	}
	if err := w.File.Close(); err != nil {
		panic(err)
	}
	w.Disabled, w.File = true, nil
}

func (w *Logger) Printf(f string, a ...interface{}) {
	if w == nil || w.Disabled || w.File == nil {
		return
	}
	_, _ = fmt.Fprintf(w.File, f, a...)
}

func (w *Logger) Write(b []byte) {
	if w.Disabled || w.File == nil {
		return
	}
	_, _ = w.File.Write(b)
}
