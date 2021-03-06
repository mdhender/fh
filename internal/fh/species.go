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
	"io/ioutil"
	"path/filepath"
)

type SpeciesData struct {
	ID         string `json:"id"`
	Number     int    `json:"number"` // one-based index of species
	Name       string `json:"name"`   // Name of species.
	Government struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"government"`
	Home struct {
		System *NamedSystem     `json:"system"`
		World  *NamedPlanetData `json:"world"`
	} `json:"home"`
	NamedSystems []*NamedSystem     `json:"systems"`
	NamedPlanets []*NamedPlanetData `json:"planets"`
	Ships        []*ShipData        `json:"ships"`
	Gases        struct {
		Required struct {
			Type GasType `json:"type"`
			Min  int     `json:"min_pct"` // Minimum needed percentage.
			Max  int     `json:"max_pct"` // Maximum allowed percentage.
		} `json:"required"` // Gas required by species.
		Neutral []GasType `json:"neutral"` // Gases neutral to species.
		Poison  []GasType `json:"poison"`  // Gases poisonous to species.
	} `json:"gases"`
	AutoOrders       bool   // AUTO command was issued.
	TechLevel        [6]int // Actual tech levels.
	InitTechLevel    [6]int // Tech levels at start of turn.
	TechKnowledge    [6]int // Unapplied tech level knowledge.
	TechEps          [6]int // Experience points for tech levels.
	HPOriginalBase   int    // If non-zero, home planet was bombed either by bombardment or germ warfare and has not yet fully recovered. Value is total economic base before bombing.
	EconUnits        int    // Number of economic units.
	FleetCost        int    // Total fleet maintenance cost.
	FleetPercentCost int    // Fleet maintenance cost as a percentage times one hundred.
	Contact          []bool // A bit is set if corresponding species has been met.
	Ally             []bool // A bit is set if corresponding species is considered an ally.
	Enemy            []bool // A bit is set if corresponding species is considered an enemy.
}

func (s *SpeciesData) AddNamedPlanet(nampla *NamedPlanetData) {
	for _, n := range s.NamedPlanets {
		if nampla.Name == n.Name {
			return
		}
	}
	s.NamedPlanets = append(s.NamedPlanets, nampla)
}

func (s *SpeciesData) ClosestUnvisitedSystem(ship *ShipData, stars []*StarData) *StarData {
	var closestStar *StarData
	var closest_distance int
	for _, star := range stars {
		/* Check if bit is already set. */
		if star.VisitedBy[s.Name] {
			continue
		}

		temp_distance := ship.Coords.DistanceSquaredTo(star.Coords)
		if closestStar == nil || temp_distance < closest_distance {
			closest_distance = temp_distance
			closestStar = star
		}
	}
	return closestStar
}

func (s *SpeciesData) GetNamedPlanet(name string) *NamedPlanetData {
	for _, n := range s.NamedPlanets {
		if name == n.Name {
			return n
		}
	}
	return nil
}

func (s *SpeciesData) GetNamedPlanetAt(c Coords) *NamedPlanetData {
	for _, n := range s.NamedPlanets {
		if n.Planet == nil {
			panic(fmt.Sprintf("getNamedPlanetAt found nil nampla %q planet", n.Name))
		}
		if n.Planet.Coords.SamePlanet(c) {
			return n
		}
	}
	return nil
}

/* Get life support tech level needed. */
func (s *SpeciesData) LifeSupportNeeded(colony *PlanetData) int {
	var ls_needed int

	// temperature class
	if colony.TemperatureClass < s.Home.World.Planet.TemperatureClass {
		ls_needed += 3 * (s.Home.World.Planet.TemperatureClass - colony.TemperatureClass)
	} else if colony.TemperatureClass > s.Home.World.Planet.TemperatureClass {
		ls_needed += 3 * (colony.TemperatureClass - s.Home.World.Planet.TemperatureClass)
	}

	// pressure class
	if colony.PressureClass < s.Home.World.Planet.PressureClass {
		ls_needed += 3 * (s.Home.World.Planet.PressureClass - colony.PressureClass)
	} else if colony.PressureClass > s.Home.World.Planet.PressureClass {
		ls_needed += 3 * (colony.PressureClass - s.Home.World.Planet.PressureClass)
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

func (s *SpeciesData) NamedPlanetsReversed() []*NamedPlanetData {
	if len(s.NamedPlanets) == 0 {
		return nil
	}
	list := make([]*NamedPlanetData, len(s.NamedPlanets), len(s.NamedPlanets))
	for i, j := len(s.NamedPlanets)-1, 0; i >= 0; i, j = i-1, j+1 {
		list[j] = s.NamedPlanets[i]
	}
	return list
}

// Report does that
func (s *SpeciesData) Report(l *Logger, galaxyPath string, turnNumber int, testMode, ignore_field_distorters, truncate_name bool, locations []*SpeciesLocationData, getPlanet func(Coords) *PlanetData, getSpecies func(id int) *SpeciesData, allSpecies []*SpeciesData) error {
	var otherSpecies []*SpeciesData
	for _, alien := range allSpecies {
		if s.Number != alien.Number {
			otherSpecies = append(otherSpecies, alien)
		}
	}

	// initialize flags
	for _, ship := range s.Ships {
		ship.alreadyListed = false
	}

	s.ReportIncludeLogFile(l, galaxyPath, turnNumber)

	s.ReportHeader(l, turnNumber)
	s.ReportTechLevels(l)
	s.ReportGases(l)
	s.ReportFleetMaintenance(l)

	s.ReportContacts(l, otherSpecies)
	s.ReportDeclaredAllies(l, otherSpecies)
	s.ReportDeclaredEnemies(l, otherSpecies)

	s.ReportEconomicUnits(l)
	s.ReportProducingPlanets(l)
	headerPrinted := false
	headerPrinted = s.ReportNonProducingPlanets(l, headerPrinted, ignore_field_distorters, truncate_name)
	headerPrinted = s.ReportShipsNotOnPlanet(l, testMode, headerPrinted, ignore_field_distorters, truncate_name)

	l.Printf("\n\n* * * * * * * * * * * * * * * * * * * * * * * * *\n")

	printingAlien := true
	s.ReportAliens(l, locations, printingAlien, getSpecies)

	return nil
}

// Report aliens at locations where current species has inhabited planets or ships
func (s *SpeciesData) ReportAliens(l *Logger, locations []*SpeciesLocationData, printingAlien bool, getSpecies func(id int) *SpeciesData) {
	for _, my_loc := range locations {
		if my_loc.S != s.Number {
			continue
		}

		header_printed := false
		for _, its_loc := range locations {
			// is there an alien at this location?
			alienHere := its_loc.S != s.Number && (my_loc.X == its_loc.X && my_loc.Y == its_loc.Y && my_loc.Z == its_loc.Z)
			if !alienHere {
				continue
			}

			// there is an alien here
			alien := getSpecies(its_loc.S)

			/* Check if we have a named planet in this system. If so, use it when you print the header. */
			we_have_planet_here := false
			var our_nampla *NamedPlanetData
			for _, nampla := range s.NamedPlanets {
				// TODO: what is so special about orbit 99?
				if nampla.Planet.Coords.Orbit == 99 || !nampla.Planet.Coords.SameSystem(Coords{X: my_loc.X, Y: my_loc.Y, Z: my_loc.Z}) {
					continue
				}
				we_have_planet_here = true
				our_nampla = nampla
				break
			}

			/* Print all inhabited alien namplas at this location. */
			for _, alien_nampla := range alien.NamedPlanets {
				if alien_nampla.Planet.Coords.Orbit == 99 || !alien_nampla.Planet.Coords.SameSystem(Coords{X: my_loc.X, Y: my_loc.Y, Z: my_loc.Z}) {
					continue
				} else if !alien_nampla.Status.Populated {
					continue
				}

				/* Check if current species has a colony on the same planet. */
				we_have_colony_here := false
				for _, nampla := range s.NamedPlanets {
					if !(nampla.Status.Populated && nampla.Planet.Coords.SamePlanet(alien_nampla.Planet.Coords)) {
						continue
					}
					we_have_colony_here = true
					break
				}
				if alien_nampla.Hidden && !we_have_colony_here {
					continue
				}

				if !header_printed {
					l.Printf("\n\nAliens at x = %d, y = %d, z = %d", my_loc.X, my_loc.Y, my_loc.Z)
					if we_have_planet_here {
						l.Printf(" (PL %s star system)", our_nampla.Name)
					}
					l.Printf(":\n")
					header_printed = true
				}

				industry := alien_nampla.MIBase + alien_nampla.MABase

				var temp1 string
				if alien_nampla.Status.MiningColony {
					temp1 = fmt.Sprintf("%s", "Mining colony")
				} else if alien_nampla.Status.ResortColony {
					temp1 = fmt.Sprintf("%s", "Resort colony")
				} else if alien_nampla.Status.HomePlanet {
					temp1 = fmt.Sprintf("%s", "Home planet")
				} else if industry > 0 {
					temp1 = fmt.Sprintf("%s", "Colony planet")
				} else {
					temp1 = fmt.Sprintf("%s", "Uncolonized planet")
				}

				temp2 := fmt.Sprintf("  %s PL %s (pl #%d)", temp1, alien_nampla.Name, alien_nampla.Planet.Coords.Orbit)
				n := 53 - len(temp2)
				for j := 0; j < n; j++ {
					temp2 += " "
				}
				l.Printf("%sSP %s\n", temp2, alien.Name)

				economicBase := industry != 0
				if industry < 100 {
					industry = (industry + 5) / 10
				} else {
					industry = ((industry + 50) / 100) * 10
				}

				if !economicBase {
					l.Printf("      (No economic base.)\n")
				} else {
					l.Printf("      (Economic base is approximately %d.)\n", industry)
				}

				/* If current species has a colony on the same planet, report any PDs and any shipyards. */
				if we_have_colony_here {
					if alien_nampla.ItemQuantity[PD] == 1 {
						l.Printf("      (There is 1 %s on the planet.)\n", itemData[PD].name)
					} else if alien_nampla.ItemQuantity[PD] > 1 {
						l.Printf("      (There are %d %ss on the planet.)\n", alien_nampla.ItemQuantity[PD], itemData[PD].name)
					}

					if alien_nampla.Shipyards == 1 {
						l.Printf("      (There is 1 shipyard on the planet.)\n")
					} else if alien_nampla.Shipyards > 1 {
						l.Printf("      (There are %d shipyards on the planet.)\n", alien_nampla.Shipyards)
					}
				}

				/* Also report if alien colony is actively hiding. */
				if alien_nampla.Hidden {
					l.Printf("      (Colony is actively hiding from alien observation.)\n")
				}
			}

			/* Print all alien ships at this location. */
			for _, alien_ship := range alien.Ships {
				// TODO: what is so special about orbit 99?
				if alien_ship.Coords.Orbit == 99 || !alien_ship.Coords.SameSystem(Coords{X: my_loc.X, Y: my_loc.Y, Z: my_loc.Z}) {
					continue
				}

				/* An alien ship cannot hide if it lands on the surface of a planet populated by the current species. */
				alien_can_hide := true
				for _, nampla := range s.NamedPlanets {
					if !nampla.Planet.Coords.SamePlanet(alien_ship.Coords) {
						continue
					}
					if nampla.Status.Populated {
						alien_can_hide = false
						break
					}
				}

				if alien_can_hide && alien_ship.Status.OnSurface {
					continue
				} else if alien_can_hide && alien_ship.Status.UnderConstruction {
					continue
				}

				if !header_printed {
					l.Printf("\n\nAliens at x = %d, y = %d, z = %d", my_loc.X, my_loc.Y, my_loc.Z)

					if we_have_planet_here {
						l.Printf(" (PL %s star system)", our_nampla.Name)
					}

					l.Printf(":\n")
					header_printed = true
				}

				alien_ship.Report(l, !header_printed, printingAlien, alien)
			}
		}
	}
}

/* List species that have been met. */
func (s *SpeciesData) ReportContacts(l *Logger, otherSpecies []*SpeciesData) {
	n := 0
	for _, alien := range otherSpecies {
		if !s.Contact[alien.Number] {
			continue
		}
		if n == 0 {
			l.String("\nSpecies met: ")
		} else {
			l.String(", ")
		}
		l.String("SP ")
		l.String(alien.Name)
		n++
	}
	if n > 0 {
		l.Char('\n')
	}
}

/* List declared allies that have been met */
func (s *SpeciesData) ReportDeclaredAllies(l *Logger, otherSpecies []*SpeciesData) {
	n := 0
	for _, alien := range otherSpecies {
		if !s.Contact[alien.Number] || !s.Ally[alien.Number] {
			continue
		}
		if n == 0 {
			l.String("\nAllies: ")
		} else {
			l.String(", ")
		}
		l.String("SP ")
		l.String(alien.Name)
		n++
	}
	if n > 0 {
		l.Char('\n')
	}
}

/* List declared enemies that have been met */
func (s *SpeciesData) ReportDeclaredEnemies(l *Logger, otherSpecies []*SpeciesData) {
	n := 0
	for _, alien := range otherSpecies {
		if !s.Contact[alien.Number] || !s.Enemy[alien.Number] {
			continue
		}
		if n == 0 {
			l.String("\nEnemies: ")
		} else {
			l.String(", ")
		}
		l.String("SP ")
		l.String(alien.Name)
		n++
	}
	if n > 0 {
		l.Char('\n')
	}
}

func (s *SpeciesData) ReportEconomicUnits(l *Logger) {
	l.Printf("\nEconomic units = %d\n", s.EconUnits)
}

/* List fleet maintenance cost and its percentage of total production. */
func (s *SpeciesData) ReportFleetMaintenance(l *Logger) {
	fleet_percent_cost := s.FleetPercentCost
	l.Printf("\nFleet maintenance cost = %d (%d.%02d%% of total production)\n", s.FleetCost, fleet_percent_cost/100, fleet_percent_cost%100)
	if fleet_percent_cost > 10000 {
		fleet_percent_cost = 10000
	}
}

func (s *SpeciesData) ReportMishapChance(l *Logger, ship *ShipData, dest Coords) {
	if dest.X == 9999 {
		l.Printf("Mishap chance = ???")
		return
	}

	mishap_chance := (100 * ship.Coords.DistanceSquaredTo(dest)) / s.TechLevel[GV]
	if ship.Age > 0 && mishap_chance < 10000 {
		success_chance := 10000 - mishap_chance
		success_chance -= (2 * ship.Age * success_chance) / 100
		mishap_chance = 10000 - success_chance
	}
	if mishap_chance > 10000 {
		mishap_chance = 10000
	}

	l.Printf("mishap chance = %d.%02d%%", mishap_chance/100, mishap_chance%100)
}

func (s *SpeciesData) ReportGases(l *Logger) {
	l.Printf("\n\n\nAtmospheric Requirement: %d%%-%d%% %s", s.Gases.Required.Min, s.Gases.Required.Max, s.Gases.Required.Type.Char())
	l.Printf("\nNeutral Gases:")
	for i, gas := range s.Gases.Neutral {
		if i != 0 {
			l.Printf(",")
		}
		l.Printf(" %s", gas.Char())
	}
	l.Printf("\nPoisonous Gases:")
	for i, gas := range s.Gases.Poison {
		if i != 0 {
			l.Printf(",")
		}
		l.Printf(" %s", gas.Char())
	}
	l.Printf("\n")
}

/* Print header for status report. */
func (s *SpeciesData) ReportHeader(l *Logger, turnNumber int) {
	l.Printf("\n\t\t\t SPECIES STATUS\n\n\t\t\tSTART OF TURN %d\n\n", turnNumber)
	l.Printf("Species name: %s\n", s.Name)
	l.Printf("Government name: %s\n", s.Government.Name)
	l.Printf("Government type: %s\n", s.Government.Type)
}

// ReportIncludeLogFile copies the log file for the prior turn into the current report
func (s *SpeciesData) ReportIncludeLogFile(l *Logger, galaxyPath string, turnNumber int) {
	msg, err := ioutil.ReadFile(filepath.Join(galaxyPath, fmt.Sprintf("sp%02d.log", s.Number)))
	if err != nil {
		return
	}
	priorTurnNumber := turnNumber - 1
	if priorTurnNumber > 0 {
		l.Printf("\n\n\t\t\tEVENT LOG FOR TURN %d\n", priorTurnNumber)
	}
	l.Write(msg)
}

// Print report for each producing planet
func (s *SpeciesData) ReportProducingPlanets(l *Logger) {
	for _, nampla := range s.NamedPlanets {
		if nampla.Planet.Coords.Orbit == 99 {
			continue
		}
		if nampla.MIBase == 0 && nampla.MABase == 0 && !nampla.Status.HomePlanet {
			continue
		}

		// g.do_planet_report(nampla, ship1_base, species)
		if nampla.Planet == nil {
			fmt.Printf("error: nampla %q planet is nil\n", nampla.Name)
		}
		nampla.Report(l, s, nampla.Planet, s.Ships)
	}
}

// Print one-line listing for non-producing planets
func (s *SpeciesData) ReportNonProducingPlanets(l *Logger, headerPrinted, ignore_field_distorters, truncate_name bool) bool {
	// printingAlien := false
	for _, nampla := range s.NamedPlanets {
		if nampla.Planet.Coords.Orbit == 99 {
			continue
		}
		if nampla.MIBase > 0 || nampla.MABase > 0 || nampla.Status.HomePlanet {
			continue
		}

		if !headerPrinted {
			l.Printf("\n\n* * * * * * * * * * * * * * * * * * * * * * * * *\n")
			l.Printf("\n\nOther planets and ships:\n\n")
			headerPrinted = true
		}
		l.Printf("%4d%3d%3d #%d\tPL %s", nampla.Planet.Coords.X, nampla.Planet.Coords.Y, nampla.Planet.Coords.Z, nampla.Planet.Coords.Orbit, nampla.Name)

		for j := 0; j < MAX_ITEMS; j++ {
			if nampla.ItemQuantity[j] > 0 {
				l.Printf(", %d %s", nampla.ItemQuantity[j], itemData[j].abbr)
			}
		}
		l.Printf("\n")

		/* Print any ships at this planet. */
		for _, ship := range s.Ships {
			if ship.alreadyListed || !ship.Coords.SamePlanet(nampla.Planet.Coords) {
				continue
			}
			l.Printf("\t\t%s", ship.GetName(ignore_field_distorters, truncate_name))
			for j := 0; j < MAX_ITEMS; j++ {
				if ship.ItemQuantity[j] > 0 {
					l.Printf(", %d %s", ship.ItemQuantity[j], itemData[j].abbr)
				}
			}
			l.Printf("\n")

			ship.alreadyListed = true
		}
	}
	return headerPrinted
}

func (s *SpeciesData) ReportShipsNotOnPlanet(l *Logger, testMode, headerPrinted, ignore_field_distorters, truncate_name bool) bool {
	for _, ship := range s.Ships {
		ship.ClearSpecial()
		if ship.alreadyListed {
			continue
		}
		ship.alreadyListed = true
		if ship.Coords.Orbit == 99 {
			continue
		}
		if !headerPrinted {
			l.Printf("\n\n* * * * * * * * * * * * * * * * * * * * * * * * *\n")
			l.Printf("\n\nOther planets and ships:\n\n")
			headerPrinted = true
		}

		if ship.Status.JumpedInCombat || ship.Status.ForcedJump {
			l.Printf("  ?? ?? ??\t%s", ship.GetName(ignore_field_distorters, truncate_name))
		} else if testMode && ship.ArrivedViaWormhole {
			l.Printf("  ?? ?? ??\t%s", ship.GetName(ignore_field_distorters, truncate_name))
		} else {
			l.Printf("%4d%3d%3d\t%s", ship.Coords.X, ship.Coords.Y, ship.Coords.Z, ship.GetName(ignore_field_distorters, truncate_name))
		}

		for i := 0; i < MAX_ITEMS; i++ {
			if ship.ItemQuantity[i] > 0 {
				l.Printf(", %d %s", ship.ItemQuantity[i], itemData[i].abbr)
			}
		}
		l.Printf("\n")

		if ship.Status.JumpedInCombat || ship.Status.ForcedJump {
			continue
		} else if testMode && ship.ArrivedViaWormhole {
			continue
		}

		/* Print other ships at the same location. */
		for _, ship2 := range s.Ships {
			// TODO: what is special about orbit 99?
			if ship2.alreadyListed || ship2.Coords.Orbit == 99 || !ship2.Coords.SameSystem(ship.Coords) {
				continue
			}
			l.Printf("\t\t%s", ship2.GetName(ignore_field_distorters, truncate_name))
			for j := 0; j < MAX_ITEMS; j++ {
				if ship2.ItemQuantity[j] > 0 {
					l.Printf(", %d %s", ship2.ItemQuantity[j], itemData[j].abbr)
				}
			}
			l.Printf("\n")

			ship2.alreadyListed = true
		}
	}
	return headerPrinted
}

func (s *SpeciesData) ReportTechLevels(l *Logger) {
	l.Printf("\nTech Levels:\n")
	l.Printf("   %s = %d", techData[MI].name, s.TechLevel[MI])
	if s.TechKnowledge[MI] > s.TechLevel[MI] {
		l.Printf("/%d", s.TechKnowledge[MI])
	}
	l.Printf("\n")
	l.Printf("   %s = %d", techData[MA].name, s.TechLevel[MA])
	if s.TechKnowledge[MA] > s.TechLevel[MA] {
		l.Printf("/%d", s.TechKnowledge[MA])
	}
	l.Printf("\n")
	l.Printf("   %s = %d", techData[ML].name, s.TechLevel[ML])
	if s.TechKnowledge[ML] > s.TechLevel[ML] {
		l.Printf("/%d", s.TechKnowledge[ML])
	}
	l.Printf("\n")
	l.Printf("   %s = %d", techData[GV].name, s.TechLevel[GV])
	if s.TechKnowledge[GV] > s.TechLevel[GV] {
		l.Printf("/%d", s.TechKnowledge[GV])
	}
	l.Printf("\n")
	l.Printf("   %s = %d", techData[LS].name, s.TechLevel[LS])
	if s.TechKnowledge[LS] > s.TechLevel[LS] {
		l.Printf("/%d", s.TechKnowledge[LS])
	}
	l.Printf("\n")
	l.Printf("   %s = %d", techData[BI].name, s.TechLevel[BI])
	if s.TechKnowledge[BI] > s.TechLevel[BI] {
		l.Printf("/%d", s.TechKnowledge[BI])
	}
	l.Printf("\n")
}

// The following routine provides the 'distorted' species number used to
// identify a species that uses field distortion units.
// TODO: this should be part of the SpeciesData struct
func (s *SpeciesData) Distorted() int {
	// We must use the LS tech level at the start of the turn because the
	// distorted species number must be the same throughout the turn, even
	// if the tech level changes during production.
	ls := s.InitTechLevel[LS]
	nibLo, nibHi := s.Number&0x000F, (s.Number>>4)&0x000F // lower four bits, upper four bits
	return (ls%5+3)*(4*nibLo+nibHi) + (ls%11 + 7)
}
