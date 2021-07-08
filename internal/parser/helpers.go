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

func bdup(src []byte) (dst []byte) {
	if len(src) != 0 {
		dst = make([]byte, len(src), len(src))
		copy(dst, src)
	}
	return dst
}

func rdup(src []rune) (dst []rune) {
	if len(src) != 0 {
		dst = make([]rune, len(src), len(src))
		copy(dst, src)
	}
	return dst
}
