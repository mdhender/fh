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

const MAX_LOCATIONS = 10000

type sp_loc_data struct {
	s, x, y, z char /* Species number, x, y, and z. */
}

type SpeciesLocation struct {
	data map[string][]int
}

type SpeciesLocationData struct {
	S int `json:"species_number"`
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

//func DoLocations(g *GalaxyData) *SpeciesLocation {
//	loc := &SpeciesLocation{data:make(map[string][]int}
//
//	for _, species := range g.AllSpecies() {
//		nampla_base := namp_data[species.Number-1]
//		ship_base := ship_data[species.Number-1]
//
//		nampla := nampla_base - 1
//		for i := 0; i < species.NumNamplas; i++ {
//			nampla++
//
//			if nampla.pn == 99 {
//				continue
//			}
//
//			if nampla.status&POPULATED != 0 {
//				loc = add_location(loc, nampla.x, nampla.y, nampla.z, species.Number)
//			}
//		}
//
//		ship := ship_base - 1
//		for i := 0; i < species.NumShips; i++ {
//			ship++
//
//			if ship.pn == 99 {
//				continue
//			}
//			if ship.status == FORCED_JUMP || ship.status == JUMPED_IN_COMBAT {
//				continue
//			}
//
//			loc = add_location(loc, ship.x, ship.y, ship.z, species.Number)
//		}
//	}
//
//	return loc
//}
