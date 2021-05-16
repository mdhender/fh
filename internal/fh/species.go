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
	"io"
	"io/ioutil"
	"path"
)

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

func (s *SpeciesData) AddNamedPlanet(nampla *NamedPlanetData) {
	for _, n := range s.NamedPlanets {
		if nampla.Name == n.Name {
			return
		}
	}
	s.NamedPlanets = append(s.NamedPlanets, nampla)
}

func (s *SpeciesData) GetNamedPlanet(name string) *NamedPlanetData {
	for _, n := range s.NamedPlanets {
		if name == n.Name {
			return n
		}
	}
	return nil
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

// Report does that
func (s *SpeciesData) Report(w io.Writer, galaxyPath string, turnNumber int, getPlanet func(Coords) *PlanetData, allSpecies []*SpeciesData) error {
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

	s.ReportIncludeLogFile(w, galaxyPath, turnNumber)

	s.ReportHeader(w, turnNumber)
	s.ReportTechLevels(w)
	s.ReportGases(w)
	s.ReportFleetMaintenance(w)

	s.ReportContacts(w, otherSpecies)
	s.ReportDeclaredAllies(w, otherSpecies)
	s.ReportDeclaredEnemies(w, otherSpecies)

	s.ReportEconomicUnits(w)
	s.ReportProducingPlanets(w, getPlanet)

	return nil
}

/* List declared allies that have been met */
func (s *SpeciesData) ReportDeclaredAllies(w io.Writer, otherSpecies []*SpeciesData) {
	l := &Logger{File: w} /* Use log utils for this. */
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

/* List fleet maintenance cost and its percentage of total production. */
func (s *SpeciesData) ReportFleetMaintenance(w io.Writer) {
	fleet_percent_cost := s.FleetPercentCost
	fmt.Fprintf(w, "\nFleet maintenance cost = %ld (%d.%02d%% of total production)\n", s.FleetCost, fleet_percent_cost/100, fleet_percent_cost%100)
	if fleet_percent_cost > 10000 {
		fleet_percent_cost = 10000
	}
}

/* List species that have been met. */
func (s *SpeciesData) ReportContacts(w io.Writer, otherSpecies []*SpeciesData) {
	l := &Logger{File: w} /* Use log utils for this. */
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

/* List declared enemies that have been met */
func (s *SpeciesData) ReportDeclaredEnemies(w io.Writer, otherSpecies []*SpeciesData) {
	l := &Logger{File: w} /* Use log utils for this. */
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

func (s *SpeciesData) ReportEconomicUnits(w io.Writer) {
	fmt.Fprintf(w, "\nEconomic units = %ld\n", s.EconUnits)
}

func (s *SpeciesData) ReportGases(w io.Writer) {
	fmt.Fprintf(w, "\n\n\nAtmospheric Requirement: %d%%-%d%% %s", s.Gases.Required.Min, s.Gases.Required.Max, s.Gases.Required.Type.Char())
	fmt.Fprintf(w, "\nNeutral Gases:")
	for i, gas := range s.Gases.Neutral {
		if i != 0 {
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, " %s", gas.Char())
	}
	fmt.Fprintf(w, "\nPoisonous Gases:")
	for i, gas := range s.Gases.Poison {
		if i != 0 {
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, " %s", gas.Char())
	}
	fmt.Fprintf(w, "\n")
}

/* Print header for status report. */
func (s *SpeciesData) ReportHeader(w io.Writer, turnNumber int) {
	fmt.Fprintf(w, "\n\t\t\t SPECIES STATUS\n\n\t\t\tSTART OF TURN %d\n\n", turnNumber)
	fmt.Fprintf(w, "Species name: %s\n", s.Name)
	fmt.Fprintf(w, "Government name: %s\n", s.GovtName)
	fmt.Fprintf(w, "Government type: %s\n", s.GovtType)
}

// ReportIncludeLogFile copies the log file for the prior turn into the current report
func (s *SpeciesData) ReportIncludeLogFile(w io.Writer, galaxyPath string, turnNumber int) {
	log_file, err := ioutil.ReadFile(path.Join(galaxyPath, fmt.Sprintf("sp%02d.log", s.Number)))
	if err != nil {
		return
	}
	priorTurnNumber := turnNumber - 1
	if priorTurnNumber > 0 {
		_, _ = fmt.Fprintf(w, "\n\n\t\t\tEVENT LOG FOR TURN %d\n", priorTurnNumber)
	}
	_, _ = w.Write(log_file)
}

// Print report for each producing planet
func (s *SpeciesData) ReportProducingPlanets(w io.Writer, getPlanet func(Coords) *PlanetData) {
	for _, nampla := range s.NamedPlanets {
		if nampla.Coords.Orbit == 99 {
			continue
		}
		if nampla.MIBase == 0 && nampla.MABase == 0 && !nampla.Status.HomePlanet {
			continue
		}

		// g.do_planet_report(nampla, ship1_base, species)
		nampla.Report(w, s, getPlanet(nampla.Coords), s.Ships)
	}
}

func (s *SpeciesData) ReportTechLevels(w io.Writer) {
	fmt.Fprintf(w, "\nTech Levels:\n")
	fmt.Fprintf(w, "   %s = %d", techData[MI].name, s.TechLevel[MI])
	if s.TechKnowledge[MI] > s.TechLevel[MI] {
		fmt.Fprintf(w, "/%d", s.TechKnowledge[MI])
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "   %s = %d", techData[MA].name, s.TechLevel[MA])
	if s.TechKnowledge[MA] > s.TechLevel[MA] {
		fmt.Fprintf(w, "/%d", s.TechKnowledge[MA])
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "   %s = %d", techData[ML].name, s.TechLevel[ML])
	if s.TechKnowledge[ML] > s.TechLevel[ML] {
		fmt.Fprintf(w, "/%d", s.TechKnowledge[ML])
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "   %s = %d", techData[GV].name, s.TechLevel[GV])
	if s.TechKnowledge[GV] > s.TechLevel[GV] {
		fmt.Fprintf(w, "/%d", s.TechKnowledge[GV])
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "   %s = %d", techData[LS].name, s.TechLevel[LS])
	if s.TechKnowledge[LS] > s.TechLevel[LS] {
		fmt.Fprintf(w, "/%d", s.TechKnowledge[LS])
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "   %s = %d", techData[BI].name, s.TechLevel[BI])
	if s.TechKnowledge[BI] > s.TechLevel[BI] {
		fmt.Fprintf(w, "/%d", s.TechKnowledge[BI])
	}
	fmt.Fprintf(w, "\n")
}
