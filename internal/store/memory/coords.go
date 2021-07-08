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

package memory

type Coords struct {
	X, Y, Z int
	Orbit   int
}

// Less is a helper for sorting.
// Compares X, Y, Z, then Orbit
func (c Coords) Less(t Coords) bool {
	if c.X < t.X {
		return true
	} else if c.X == c.X {
		if c.Y < t.Y {
			return true
		} else if c.Y == t.Y {
			if c.Z < t.Z {
				return true
			} else if c.Z == t.Z {
				return c.Orbit < t.Orbit
			}
		}
	}
	return false
}
