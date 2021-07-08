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
	"fmt"
	"github.com/mdhender/fh/internal/store/jsondb"
)

func Locations(jdb *jsondb.Store, turnData *TurnData, test, verbose bool) {
	// reset the total economic base and efficiency for each planet.
	// why? why don't we trust the data store to be accurate?
	for _, p := range jdb.Planets {
		p.TotalEconomicBase = 0
		p.EconEfficiency = 100 // default to 100%
	}

	// calculate the total economic base on each planet.
	// this is based on all colonies, but excludes home colonies.
	for _, sp := range jdb.Species {
		for nid, nampla := range sp.NamedPlanets {
			if nampla.Orbit == 99 { // TODO: 99 means no planet here
				continue
			}
			if !nampla.Status.HomePlanet {
				jdb.Planets[nampla.PlanetIndex].TotalEconomicBase += nampla.MiBase + nampla.MaBase
				if nampla.MiBase+nampla.MaBase != 0 {
					fmt.Printf("planet %4d mi %4d ma %4d SP%02d %q\n", nampla.PlanetIndex, nampla.MiBase, nampla.MaBase, sp.Id, nid)
				}
			}
		}
	}

	// update economic efficiencies of all planets
	for _, p := range jdb.Planets {
		if p.Id < 1 {
			continue
		}
		excess := (p.TotalEconomicBase - 2000) / 20
		if excess > 0 {
			p.EconEfficiency = (100 * (excess + 2000)) / p.TotalEconomicBase
		}
	}

	for _, p := range jdb.Planets {
		if p.TotalEconomicBase != 0 {
			fmt.Printf("planet %4d eb %6d ee %3d\n", p.Id, p.TotalEconomicBase, p.EconEfficiency)
		}
	}

	DoLocations(jdb)
}

// DoLocations determines the current locations of colonies and ships.
// It updates Locations in the data store.
// TODO: What does Locations really mean? Is it reset each turn and shows
// only the current locations? Or is it a history of locations visited?
func DoLocations(jdb *jsondb.Store) {
	for _, sp := range jdb.Species {
		for _, nampla := range sp.NamedPlanets {
			if nampla.Orbit == 99 { // ignore empty orbits?
				continue
			}
			if nampla.Status.Populated {
				addLocation(jdb, fmt.Sprintf("SP%02d", sp.Id), nampla.Coords)
			}
		}
		for _, ship := range sp.Ships {
			if ship.Orbit == 99 { // ignore empty orbits?
				continue
			}
			// TODO: ship status should be const, not strings
			if ship.Status == "FORCED_JUMP" || ship.Status == "JUMPED_IN_COMBAT" {
				continue
			}
			addLocation(jdb, fmt.Sprintf("SP%02d", sp.Id), ship.Coords)
		}
	}
}

func addLocation(jdb *jsondb.Store, id string, coords jsondb.Coords) {
	key := coords.Key()
	a, ok := jdb.Locations[key]
	if !ok {
		jdb.Locations[key] = []string{id}
	} else {
		for _, s := range a {
			if s == id {
				return
			}
		}
		jdb.Locations[key] = append(a, id)
	}
}
