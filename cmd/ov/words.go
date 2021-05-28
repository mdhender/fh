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


func words(b []byte) [][]byte {
	if len(b) == 0 {
		return nil
	}
	var v [][]byte
	var w []byte
	for len(b) != 0 {
		switch b[0] {
		case ' ', '\t', '\r':
			if w != nil {
				v = append(v, w)
			}
			for len(b) != 0 && (b[0] == ' ' || b[0] == '\t' || b[0] == '\r') {
				b = b[1:]
			}
			w = nil
		default:
			w, b = append(w, b[0]), b[1:]
		}
	}
	if w != nil {
		v = append(v, w)
	}
	return v
}
