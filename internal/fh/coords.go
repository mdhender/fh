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
	"math"
	"math/rand"
	"time"
)

type Coords struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	Z     int `json:"z"`
	Orbit int `json:"orbit,omitempty"` // zero means the star, non-zero is planet number
}

func (c Coords) CloserThan(t Coords, d int) bool {
	dx, dy, dz := c.X-t.X, c.Y-t.Y, c.Z-t.Z
	return dx*dx+dy*dy+dz*dz < d*d
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

func (c Coords) DistanceTo(t Coords) int {
	dX, dY, dZ := c.X-t.X, c.Y-t.Y, c.Z-t.Z
	return int(math.Round(math.Sqrt(float64(dX*dX + dY*dY + dZ*dZ))))
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
	return fmt.Sprintf("%3d %3d %3d", c.X, c.Y, c.Z)
}

var __coordsPRNG *rand.Rand

// RandomCoords returns a random point within the sphere with a decent distribution.
// For alternative methods, see
//   https://karthikkaranth.me/blog/generating-random-points-in-a-sphere/
//   https://mathworld.wolfram.com/SpherePointPicking.html
func RandomCoords(radius int) Coords {
	if __coordsPRNG == nil {
		__coordsPRNG = rand.New(rand.NewSource(time.Now().UnixNano()))
		//__coordsPRNG = rand.New(rand.NewSource(0xBADC0FFEE))
	}

	r := float64(radius)
	rSquared := float64(radius * radius)
	d := float64(2 * radius)
	x := d*__coordsPRNG.Float64() - r
	y := d*__coordsPRNG.Float64() - r
	z := d*__coordsPRNG.Float64() - r
	for x*x+y*y+z*z > rSquared {
		x = d*__coordsPRNG.Float64() - r
		y = d*__coordsPRNG.Float64() - r
		z = d*__coordsPRNG.Float64() - r
	}
	return Coords{
		X: int(math.Round(x)),
		Y: int(math.Round(y)),
		Z: int(math.Round(z)),
	}
}

// BestCoord returns a random point within the sphere with a decent distribution.
func BestCoord(radius int) Coords {
	if __coordsPRNG == nil {
		__coordsPRNG = rand.New(rand.NewSource(time.Now().UnixNano()))
		//__coordsPRNG = rand.New(rand.NewSource(0xBADC0FFEE))
	}

	r := float64(radius)
	rSquared := float64(radius * radius)
	d := float64(2 * radius)
	x := d*__coordsPRNG.Float64() - r
	y := d*__coordsPRNG.Float64() - r
	z := d*__coordsPRNG.Float64() - r
	for x*x+y*y+z*z > rSquared {
		x = d*__coordsPRNG.Float64() - r
		y = d*__coordsPRNG.Float64() - r
		z = d*__coordsPRNG.Float64() - r
	}
	return Coords{
		X: int(math.Round(x)),
		Y: int(math.Round(y)),
		Z: int(math.Round(z)),
	}
}

// RandomXYZ selects a random point within the sphere with a decent distribution.
// Range of x,y,z is -r..0..r.
// Point will be at least dPoint units away from the given point.
// Point will be at least dHome units away from a home system.
func RandomXYZ(r int, p Coords, dP int, holes []*StarData, dHoles int, homes []*StarData, dHomes int, systems []*StarData, dSystems int) (int,int,int) {
	a, b, rSquared := -1 * (r + 1), 2 * r + 1, r * r
	for {
		x,y,z := a+rnd(b),a+rnd(b),a+rnd(b)
		// point must be within the sphere
		if rSquared < x*x + y*y + z*z {
			continue
		}
		// point must not be close to first point
		if dP > 0 {
			dX, dY, dZ := p.X-x, p.Y-y, p.Z-z
			if dP*dP < dX * dX + dY*dY + dZ * dZ {
				continue
			}
		}
		// point must not be close to wormhole
		if holes != nil && dHoles > 0 {
			d, tooClose := dHoles * dHoles, false
			for _, s := range holes {
				dX, dY, dZ := s.Coords.X-x, s.Coords.Y-y, s.Coords.Z-z
				if d < dX * dX + dY*dY + dZ * dZ {
					tooClose = true
					break
				}
			}
			if tooClose {
				continue
			}
		}
		// point must not be close to home system
		if homes != nil && dHomes > 0 {
			d, tooClose := dHomes * dHomes, false
			for _, s := range homes {
				dX, dY, dZ := s.Coords.X-x, s.Coords.Y-y, s.Coords.Z-z
				if d < dX * dX + dY*dY + dZ * dZ {
					tooClose = true
					break
				}
			}
			if tooClose {
				continue
			}
		}
		// point must not be close to system
		if systems != nil && dSystems > 0 {
			d, tooClose := dSystems * dSystems, false
			for _, s := range systems {
				dX, dY, dZ := s.Coords.X-x, s.Coords.Y-y, s.Coords.Z-z
				if d < dX * dX + dY*dY + dZ * dZ {
					tooClose = true
					break
				}
			}
			if tooClose {
				continue
			}
		}
		return x,y,z
	}
}