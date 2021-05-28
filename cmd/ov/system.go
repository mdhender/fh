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

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
)

type System struct {
	X, Y, Z int
}

func getSystems(name string) ([]*System, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var s []*System
	for l, line := range bytes.Split(b, []byte{'\n'}) {
		f := words(line)
		if f == nil {
			fmt.Printf("Read in %d systems\n", l)
			break
		}
		x, err := strconv.Atoi(string(f[2]))
		if err != nil {
			return nil, fmt.Errorf("line %d: x: %w", l+1, err)
		}
		y, err := strconv.Atoi(string(f[5]))
		if err != nil {
			return nil, fmt.Errorf("line %d: y: %w", l+1, err)
		}
		z, err := strconv.Atoi(string(f[8]))
		if err != nil {
			return nil, fmt.Errorf("line %d: z: %w", l+1, err)
		}
		s = append(s, &System{X:x, Y:y, Z:z})
	}
	return s, nil
}

