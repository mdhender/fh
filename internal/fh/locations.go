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

func DoLocations(g *GalaxyData) *SpeciesLocation {
	loc := &SpeciesLocation{data: make(map[string][]int)}

	for _, species := range g.AllSpecies() {
		for _, nampla := range species.NamedPlanets {
			// TODO: what is special about 99?
			if nampla.Coords.Orbit == 99 {
				continue
			}

			if nampla.Status.Populated {
				loc.AddLocation(species.Number, nampla.Coords)
			}
		}

		for _, ship := range species.Ships {
			// TODO: what is special about 99?
			if ship.Coords.Orbit == 99 {
				continue
			}
			if ship.Status.ForcedJump || ship.Status.JumpedInCombat {
				continue
			}

			loc.AddLocation(species.Number, ship.Coords)
		}
	}

	return loc
}

func (s *SpeciesLocation) AddLocation(sn int, c Coords) {
	for _, n := range s.data[c.String()] {
		if sn == n {
			// species already in this location
			return
		}
	}
	s.data[c.String()] = append(s.data[c.String()], sn)
}
