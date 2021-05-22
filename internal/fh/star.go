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
)

type StarData struct {
	ID           string                                                           `json:"id"`
	SystemNumber int                                                              `json:"system_number"` // one base index
	Coords       Coords                                                           `json:"xyz"`           /* Coordinates. */
	Type         StarType        /* Dwarf, degenerate, main sequence or giant. */ // was `type`
	Color        StarColor       /* Star color. Blue, blue-white, etc. */
	Size         int             /* Star size, from 0 thru 9 inclusive. */
	NumPlanets   int             /* Number of usable planets in star system. */
	HomeSystem   bool            /* TRUE if this is a good potential home system. */
	Wormhole     *StarData       // set only if the system contains a natural wormhole
	Message      int             /* Message associated with this star system, if any. */
	VisitedBy    map[string]bool `json:"visited_by"` // map of species id, true if corresponding species has been here.
	PlanetIndex  int             /* Index (starting at zero) into the file "planets.dat" of the first planet in the star system. */
	Planets      []*PlanetData
}

func NewStar(l *Logger) (*StarData, error) {
	star := &StarData{
		Size:      rnd(10) - 1, // size is totally random
		VisitedBy: make(map[string]bool),
	}

	// determine type of star, favoring MAIN_SEQUENCE stars.
	// the type of the star will influence the number of dice we roll
	// when generating the planets.
	var numberOfDice int
	switch rnd(GIANT + 6) {
	case 1:
		star.Type, numberOfDice = DWARF, 1
	case 2:
		star.Type, numberOfDice = DEGENERATE, 2
	case 3:
		star.Type, numberOfDice = GIANT, 3
	default:
		star.Type, numberOfDice = MAIN_SEQUENCE, 2
	}

	// color of star is totally random, but the color impacts the number
	// of dice that we roll later when generating the planets.
	var sizeOfDie int
	switch c := rnd(RED); c {
	case BLUE:
		star.Color, sizeOfDie = StarColor(c), 8
	case BLUE_WHITE:
		star.Color, sizeOfDie = StarColor(c), 7
	case WHITE:
		star.Color, sizeOfDie = StarColor(c), 6
	case YELLOW_WHITE:
		star.Color, sizeOfDie = StarColor(c), 5
	case YELLOW:
		star.Color, sizeOfDie = StarColor(c), 4
	case ORANGE:
		star.Color, sizeOfDie = StarColor(c), 3
	case RED:
		star.Color, sizeOfDie = StarColor(c), 2
	default:
		return nil, fmt.Errorf("assert(StarColor != %d)", c)
	}

	// initialize the planet generator to bias it towards fewer planets
	numPlanets := -2
	for i := 0; i < numberOfDice; i++ {
		numPlanets += rnd(sizeOfDie)
	}
	// then adjust if too few or too many planets
	for numPlanets < 1 {
		numPlanets += rnd(2)
	}
	for numPlanets > 9 {
		numPlanets -= rnd(3)
	}
	star.Planets = make([]*PlanetData, numPlanets, numPlanets)

	l.Printf("Generated star (type %-13s) (planets %d)\n", star.Type, len(star.Planets))

	return star, nil
}

func GenerateStar(lg Loggy, at Coords, nSpecies int) (*StarData, error) {
	lg.Log("Generating star (%3d, %3d, %3d)\n", at.X, at.Y, at.Z)

	/* Set coordinates. */
	star := &StarData{
		ID:          at.String(),
		Coords:      at,
		NumPlanets:  -2, // default value to initialize the planet generator
		PlanetIndex: -1,
		VisitedBy:   make(map[string]bool),
	}

	/* Determine type of star. Make MAIN_SEQUENCE the most common star type. */
	switch rnd(GIANT + 6) {
	case 1:
		star.Type = DWARF
	case 2:
		star.Type = DEGENERATE
	case 3:
		star.Type = GIANT
	default:
		star.Type = MAIN_SEQUENCE
	}

	/* Determine the number of planets in orbit around the star. The algorithm is something I tweaked until I liked it. It's weird, but it works. */
	/* Color and size of star are totally random. */
	star.Size = rnd(10) - 1
	switch c := rnd(RED); c {
	case BLUE:
		star.Color = BLUE
	case BLUE_WHITE:
		star.Color = BLUE_WHITE
	case WHITE:
		star.Color = WHITE
	case YELLOW_WHITE:
		star.Color = YELLOW_WHITE
	case YELLOW:
		star.Color = YELLOW
	case ORANGE:
		star.Color = ORANGE
	case RED:
		star.Color = RED
	default:
		return nil, fmt.Errorf("assert(StarColor != %d)", c)
	}

	/* Size of die. Big stars (blue, blue-white) roll bigger dice. Smaller stars (orange, red) roll smaller dice. */
	var sizeOfDie int
	switch star.Color {
	case BLUE:
		sizeOfDie = 8
	case BLUE_WHITE:
		sizeOfDie = 7
	case WHITE:
		sizeOfDie = 6
	case YELLOW_WHITE:
		sizeOfDie = 5
	case YELLOW:
		sizeOfDie = 4
	case ORANGE:
		sizeOfDie = 3
	case RED:
		sizeOfDie = 2
	}

	/* Number of rolls: dwarves have 1 roll, degenerates and main sequence stars have 2 rolls, and giants have 3 rolls. */
	var numberOfDice int
	switch star.Type {
	case DWARF:
		numberOfDice = 1
	case DEGENERATE:
		numberOfDice = 2
	case MAIN_SEQUENCE:
		numberOfDice = 2
	case GIANT:
		numberOfDice = 3
	default:
		panic(fmt.Sprintf("assert(star.Type != %d)", star.Type))
	}

	for i := 1; i <= numberOfDice; i++ {
		star.NumPlanets += rnd(sizeOfDie)
	}
	// adjust if too few or too many planets
	for star.NumPlanets < 1 {
		star.NumPlanets += rnd(2)
	}
	for star.NumPlanets > 9 {
		star.NumPlanets -= rnd(3)
	}

	lg.Log("Generating star (%3d, %3d, %3d) (type %-13s) (planets %d)\n", at.X, at.Y, at.Z, star.Type, star.NumPlanets)

	// generate planets
	var err error
	star.Planets, err = GeneratePlanet(star.ID, star.Coords, star.NumPlanets)
	if err != nil {
		return nil, err
	}

	return star, nil
}

func (s *StarData) At(x, y, z int) bool {
	return s != nil && s.Coords.X == x && s.Coords.Y == y && s.Coords.Z == z
}

// ConvertToHomeSystem converts the system to a system with a home planet
func (s *StarData) ConvertToHomeSystem(src []*PlanetData) {
	fmt.Printf("[convert] %s planets %d %d\n", s.Coords.XYZ(), len(s.Planets), len(src))
	s.HomeSystem = true

	// update the star with values from the source template
	for i, planet := range src {
		s.Planets[i] = planet.Clone()
	}

	// make minor random changes to the planets
	for _, planet := range s.Planets {
		if planet.TemperatureClass == 0 {
			// no changes
		} else if planet.TemperatureClass > 12 {
			planet.TemperatureClass -= rnd(3) - 1
		} else {
			planet.TemperatureClass += rnd(3) - 1
		}
		if planet.PressureClass == 0 {
			// no changes
		} else if planet.PressureClass > 12 {
			planet.PressureClass -= rnd(3) - 1
		} else {
			planet.PressureClass += rnd(3) - 1
		}
		if len(planet.Gases) > 2 {
			j := rnd(25) + 10
			a, b := 1, 2
			if planet.Gases[b].Percentage > 50 {
				planet.Gases[a].Percentage += j
				planet.Gases[b].Percentage -= j
			} else if planet.Gases[a].Percentage > 50 {
				planet.Gases[a].Percentage -= j
				planet.Gases[b].Percentage += j
			}
		}
		if planet.Diameter > 12 {
			planet.Diameter -= rnd(3) - 1
		} else {
			planet.Diameter += rnd(3) - 1
		}
		if planet.Gravity > 100 {
			planet.Gravity -= rnd(10)
		} else {
			planet.Gravity += rnd(10)
		}
		if planet.MiningDifficulty > 100 {
			planet.MiningDifficulty -= rnd(10)
		} else {
			planet.MiningDifficulty += rnd(10)
		}
	}
}

// returns index, not number
func (s *StarData) HomePlanetIndex() int {
	for i, planet := range s.Planets {
		if planet.Special == IDEAL_HOME_PLANET {
			return i
		}
	}
	return -1
}

// returns number, not index
func (s *StarData) HomePlanetNumber() int {
	for i, planet := range s.Planets {
		if planet.Special == IDEAL_HOME_PLANET {
			return i + 1
		}
	}
	return 0
}

func (s *StarData) Scan(l *Logger, species *SpeciesData) error {
	/* Print data for star, */
	l.Printf("Coordinates:\tx = %d\ty = %d\tz = %d", s.Coords.X, s.Coords.Y, s.Coords.Z)
	l.Printf("\tstellar type = %s%s%d", s.Type.Char(), s.Color.Char(), s.Size)

	l.Printf("   %d planets.\n\n", s.NumPlanets)

	if s.Wormhole != nil {
		l.Printf("This star system is the terminus of a natural wormhole.\n\n")
	}

	/* Print header. */
	l.Printf("               Temp  Press Mining\n")
	l.Printf("  #  Dia  Grav Class Class  Diff  LSN  Atmosphere\n")
	l.Printf(" ---------------------------------------------------------------------\n")

	/* Check for nova. */
	isNova := s.NumPlanets == 0
	for i := 0; !isNova && i < len(s.Planets); i++ {
		isNova = s.Planets[i] == nil
	}
	if isNova {
		l.Printf("\n\tThis star is a nova remnant. Any planets it may have once\n")
		l.Printf("\thad have been blown away.\n\n")
		return nil
	}

	/* Print data for each planet. */
	for i, planet := range s.Planets {
		/* Get life support tech level needed. */
		ls_needed := 99
		if species != nil {
			ls_needed = species.LifeSupportNeeded(planet)
		}

		l.Printf("  %d  %3d  %d.%02d  %2d    %2d    %d.%02d %4d  ",
			i+1, planet.Diameter,
			planet.Gravity/100, planet.Gravity%100,
			planet.TemperatureClass, planet.PressureClass,
			planet.MiningDifficulty/100, planet.MiningDifficulty%100,
			ls_needed)

		if len(planet.Gases) == 0 {
			l.Printf("No atmosphere")
		} else {
			for n, gas := range planet.Gases {
				if n > 0 {
					l.Printf(",")
				}
				l.Printf("%s(%d%%)", gas.Type.Char(), gas.Percentage)
			}
		}

		l.Printf("\n")
	}
	if s.Message != 0 {
		// TODO: there is a message that must be logged whenever this star system is scanned.
		l.Printf("TODO: there is a message that must be logged whenever this star system is scanned.\n")
		//sprintf(filename, "message%d.txt\0", star->message);
		//log_message(filename);
	}

	return nil
}
