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
	"fmt"
	"io"
	"strings"
	"unicode"
)

/* This routine is intended to take a long argument and return a pointer to a string that has embedded commas to make the string more readable. */
func Commas(value int) string {
	isNegative := value < 0
	var absValue int
	if isNegative {
		absValue = -1 * value
	} else {
		absValue = value
	}
	src := []byte(fmt.Sprintf("%d", absValue))
	digitsToCopy := len(src) + (len(src)-1)/3
	if isNegative {
		digitsToCopy++
	}
	dst := make([]byte, digitsToCopy, digitsToCopy)
	for i, j, digitsCopied := len(src)-1, len(dst)-1, 0; i >= 0; i, j = i-1, j-1 {
		if digitsCopied == 3 {
			dst[j] = ','
			j, digitsCopied = j-1, 0
		}
		dst[j] = src[i]
		digitsCopied++
	}
	if isNegative {
		dst[0] = '-'
	}
	return string(dst)
}

// EstimateNumberOfSystems returns a good guess at how many systems to generate.
// The number is based on the number of players. If the game-master specifies the
// low-density flag, then we bump that number by 50%. That, in turn, causes the
// calculated galactic radius to increase to keep the density about the same.
func EstimateNumberOfSystems(numberOfSpecies int, highDensity bool) int {
	if highDensity {
		return (3 * numberOfSpecies * STANDARD_NUMBER_OF_STAR_SYSTEMS) / (2 * STANDARD_NUMBER_OF_SPECIES)
	}
	return (numberOfSpecies * STANDARD_NUMBER_OF_STAR_SYSTEMS) / STANDARD_NUMBER_OF_SPECIES
}

func IsValidName(name string) error {
	if name != strings.TrimSpace(name) {
		return fmt.Errorf("name can't have leading or trailing spaces")
	} else if name == "" {
		return fmt.Errorf("name can't be blank")
	}
	for _, ch := range name {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || unicode.IsSpace(ch) || ch == '.' || ch == '\'' || ch == '-') {
			return fmt.Errorf("invalid character %q in name", string(ch))
		}
	}
	return nil
}

type Loggy interface {
	Log(format string, a ...interface{})
}

func ValidateNumberOfPlayers(n int) error {
	if !(MIN_SPECIES <= n && n <= MAX_SPECIES) {
		return fmt.Errorf("number of players must be between %d and %d, inclusive", MIN_SPECIES, MAX_SPECIES)
	}
	return nil
}

type Writer struct {
	Disabled bool
	File     io.WriteCloser
}

func (w *Writer) Close() error {
	if w.File == nil {
		return nil
	}
	err := w.File.Close()
	w.Disabled, w.File = true, nil
	return err
}

func (w *Writer) Printf(f string, a ...interface{}) {
	if w == nil || w.Disabled || w.File == nil {
		return
	}
	_, _ = fmt.Fprintf(w.File, f, a...)
}

func (w *Writer) Write(b []byte) {
	if w.Disabled || w.File == nil {
		return
	}
	_, _ = w.File.Write(b)
}
