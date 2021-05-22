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

import "fmt"

type Coords struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	Z     int `json:"z"`
	Orbit int `json:"orbit,omitempty"` // zero means the star, non-zero is planet number
}

func (c Coords) DeltaXYZ(t Coords) (int, int, int) {
	dX, dY, dZ := c.X-t.X, c.Y-t.Y, c.Z-t.Z
	if dX < 0 {
		dX *= -1
	}
	if dY < 0 {
		dY *= -1
	}
	if dZ < 0 {
		dZ *= -1
	}
	return dX, dY, dZ
}

func (c Coords) DistanceSquaredTo(t Coords) int {
	dX, dY, dZ := c.X-t.X, c.Y-t.Y, c.Z-t.Z
	return (dX)*(dX) + (dY)*(dY) + (dZ)*(dZ)
}

func (c Coords) ID() string {
	return fmt.Sprintf("%03d.%03d.%03d/%02d", c.X, c.Y, c.Z, c.Orbit)
}

func (c Coords) IsSet() bool {
	return c.X != -1 && c.Y != -1 && c.Z != -1 && c.Orbit != -1
}

func (c Coords) SamePlanet(t Coords) bool {
	return c.X == t.X && c.Y == t.Y && c.Z == t.Z && c.Orbit == t.Orbit
}

func (c Coords) SameSystem(t Coords) bool {
	return c.X == t.X && c.Y == t.Y && c.Z == t.Z
}

func (c Coords) String() string {
	return fmt.Sprintf("%d %d %d", c.X, c.Y, c.Z)
}

func (c Coords) SystemID() int {
	return (c.X*1_000+c.Y)*1_000 + c.Z
}

func (c Coords) XYZ() string {
	return fmt.Sprintf("%d %d %d", c.X, c.Y, c.Z)
}
