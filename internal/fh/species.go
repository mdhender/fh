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

type SpeciesData struct {
	ID       string `json:"id"`
	Number   int    // one-based index of species
	Name     string // Name of species.
	GovtName string // Name of government.
	GovtType string // Type of government.
	Home     struct {
		Coords Coords      `json:"coords"`
		PN     int         `json:"pn"` // planet number?
		System *StarData   `json:"-"`
		Planet *PlanetData `json:"-"`
	} `json:"home"`
	HomeNampla       string    `json:"home_planet_id"`
	RequiredGas      GasType   // Gas required by species.
	RequiredGasMin   int       // Minimum needed percentage.
	RequiredGasMax   int       // Maximum allowed percentage.
	NeutralGas       []GasType // Gases neutral to species.
	PoisonGas        []GasType // Gases poisonous to species.
	AutoOrders       bool      // AUTO command was issued.
	TechLevel        [6]int    // Actual tech levels.
	InitTechLevel    [6]int    // Tech levels at start of turn.
	TechKnowledge    [6]int    // Unapplied tech level knowledge.
	NumNamplas       int       // Number of named planets, including home planet and colonies.
	NumShips         int       // Number of ships.
	TechEps          [6]int    // Experience points for tech levels.
	HPOriginalBase   int       // If non-zero, home planet was bombed either by bombardment or germ warfare and has not yet fully recovered. Value is total economic base before bombing.
	EconUnits        int       // Number of economic units.
	FleetCost        int       // Total fleet maintenance cost.
	FleetPercentCost int       // Fleet maintenance cost as a percentage times one hundred.
	Contact          []bool    // A bit is set if corresponding species has been met.
	Ally             []bool    // A bit is set if corresponding species is considered an ally.
	Enemy            []bool    // A bit is set if corresponding species is considered an enemy.
	Translate        struct {
		PlanetNameToID []string `json:"planet_name_to_id"`
	} `json:"translate"`
}

/* Get life support tech level needed. */

func (s *SpeciesData) LifeSupportNeeded(colony *PlanetData) int {
	var ls_needed int

	// temperature class
	if colony.TemperatureClass < s.Home.Planet.TemperatureClass {
		ls_needed += 3 * (s.Home.Planet.TemperatureClass - colony.TemperatureClass)
	} else if colony.TemperatureClass > s.Home.Planet.TemperatureClass {
		ls_needed += 3 * (colony.TemperatureClass - s.Home.Planet.TemperatureClass)
	}

	// pressure class
	if colony.PressureClass < s.Home.Planet.PressureClass {
		ls_needed += 3 * (s.Home.Planet.PressureClass - colony.PressureClass)
	} else if colony.PressureClass > s.Home.Planet.PressureClass {
		ls_needed += 3 * (colony.PressureClass - s.Home.Planet.PressureClass)
	}

	// check for required and poisonous gases on planet
	requiredGasFound := false
	for _, gas := range colony.Gases {
		if gas.Percentage == 0 {
			continue
		}
		// check for required gas at required levels
		if gas.Type == s.RequiredGas {
			requiredGasFound = s.RequiredGasMin <= gas.Percentage && gas.Percentage <= s.RequiredGasMax
		} else {
			// compare with poisonous gases
			for _, poison := range s.PoisonGas {
				if gas.Type == poison {
					ls_needed += 3
					break
				}
			}
		}
	}
	if !requiredGasFound {
		ls_needed += 3
	}

	return ls_needed
}
