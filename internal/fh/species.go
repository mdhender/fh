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
	Number   int    `json:"number"`    // one-based index of species
	Name     string `json:"name"`      // Name of species.
	GovtName string `json:"govt_name"` // Name of government.
	GovtType string `json:"govt_type"` // Type of government.
	Home     struct {
		Coords Coords      `json:"coords"`
		System *StarData   `json:"-"`
		Planet *PlanetData `json:"-"`
	} `json:"home"`
	HomeNampla string `json:"home_planet_id"`
	Gases      struct {
		Required struct {
			Type GasType `json:"type"`
			Min  int     `json:"min_pct"` // Minimum needed percentage.
			Max  int     `json:"max_pct"` // Maximum allowed percentage.
		} `json:"required"` // Gas required by species.
		Neutral []GasType `json:"neutral"` // Gases neutral to species.
		Poison  []GasType `json:"poison"`  // Gases poisonous to species.
	} `json:"gases"`
	AutoOrders       bool               // AUTO command was issued.
	TechLevel        [6]int             // Actual tech levels.
	InitTechLevel    [6]int             // Tech levels at start of turn.
	TechKnowledge    [6]int             // Unapplied tech level knowledge.
	NumNamplas       int                // Number of named planets, including home planet and colonies.
	NamedPlanets     []*NamedPlanetData `json:"named_planets"`
	Ships            []*ShipData        `json:"ships"`
	NumShips         int                // Number of ships.
	TechEps          [6]int             // Experience points for tech levels.
	HPOriginalBase   int                // If non-zero, home planet was bombed either by bombardment or germ warfare and has not yet fully recovered. Value is total economic base before bombing.
	EconUnits        int                // Number of economic units.
	FleetCost        int                // Total fleet maintenance cost.
	FleetPercentCost int                // Fleet maintenance cost as a percentage times one hundred.
	Contact          []bool             // A bit is set if corresponding species has been met.
	Ally             []bool             // A bit is set if corresponding species is considered an ally.
	Enemy            []bool             // A bit is set if corresponding species is considered an enemy.
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
		if gas.Type == s.Gases.Required.Type {
			requiredGasFound = s.Gases.Required.Min <= gas.Percentage && gas.Percentage <= s.Gases.Required.Max
		} else {
			// compare with poisonous gases
			for _, poison := range s.Gases.Poison {
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
