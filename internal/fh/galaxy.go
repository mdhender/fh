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
	"encoding/json"
	"fmt"
	"github.com/mdhender/fh/internal/prng"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type GalaxyData struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Secret            string
	Players           map[string]*Player
	Species           map[string]*SpeciesData
	DNumSpecies       int
	NumSpecies        int
	Radius            int
	NumberOfStars     int
	NumberOfWormHoles int
	NumberOfPlanets   int
	Stars             map[string]*StarData
	Templates         struct {
		Homes [10][]*PlanetData
	}
	Translate struct {
		EmailToID        map[string]string `json:"email_to_id"`
		SpeciesNameToID  map[string]string `json:"species_name_to_id"`
		IndexToSpeciesID []string          `json:"index_to_species_id"`
		IndexToStarID    []string          `json:"index_to_star_id"`
		XYZToID          map[string]string `json:"xyz_to_id"`
	}
	allSpecies []*SpeciesData
	allStars   []*StarData
	allPlanets []*PlanetData
	logFile    io.Writer
}

type Player struct {
	ID           string `json:"id"`
	EmailAddress string `json:"email"`
	Species      string `json:"species"`
}

func GenerateGalaxy(logFile io.Writer, setupData *SetupData, players []*PlayerData) (*GalaxyData, error) {
	g := &GalaxyData{
		ID:      setupData.Galaxy.Name,
		Name:    setupData.Galaxy.Name,
		Secret:  "your-private-key-belongs-here",
		Players: make(map[string]*Player),
		Species: make(map[string]*SpeciesData),
		Stars:   make(map[string]*StarData),
		logFile: logFile,
	}
	g.Translate.EmailToID = make(map[string]string)
	g.Translate.SpeciesNameToID = make(map[string]string)
	g.Translate.XYZToID = make(map[string]string)

	// initialize from some player data
	for _, player := range players {
		id := player.Email
		g.Players[id] = &Player{
			ID:           id,
			EmailAddress: player.Email,
			Species:      player.SpeciesName,
		}
		g.Translate.EmailToID[id] = player.Email
	}

	// init?
	d_num_species := len(players)
	if d_num_species < MIN_SPECIES || MAX_SPECIES < d_num_species {
		return nil, fmt.Errorf("number of species must be between %d and %d, inclusive", MIN_SPECIES, MAX_SPECIES)
	}
	adjusted_number_of_species := d_num_species
	if setupData.Galaxy.LowDensity {
		// add 50% more species to the mix as a way to trick the program into adding more stars
		adjusted_number_of_species = (d_num_species * 3) / 2
		if adjusted_number_of_species < MIN_SPECIES || MAX_SPECIES < adjusted_number_of_species {
			g.Log("Low density option giving, boosting species count to %d\n", adjusted_number_of_species)
			return nil, fmt.Errorf("adjusted number of species must be between %d and %d, inclusive", MIN_SPECIES, MAX_SPECIES)
		}
	}
	g.DNumSpecies = d_num_species

	// get approximate number of star systems to generate
	desired_num_stars := (adjusted_number_of_species * STANDARD_NUMBER_OF_STAR_SYSTEMS) / STANDARD_NUMBER_OF_SPECIES
	g.Log("For %d species, there should be about %d stars.\n", d_num_species, desired_num_stars)
	if setupData.Galaxy.Overrides.UseOverrides {
		if setupData.Galaxy.LowDensity {
			g.Log("For %d species, a low density game needs about %d stars.\n", d_num_species, desired_num_stars)
		} else {
			g.Log("For %d species, a game needs about %d stars.\n", d_num_species, desired_num_stars)
		}
		desired_num_stars = setupData.Galaxy.Overrides.NumberOfStars
	}
	if desired_num_stars < MIN_STARS || MAX_STARS < desired_num_stars {
		return nil, fmt.Errorf("number of stars must be between %d and %d, inclusive", MIN_STARS, MAX_STARS)
	}

	// get size of galaxy to generate.
	volume := desired_num_stars * STANDARD_GALACTIC_RADIUS * STANDARD_GALACTIC_RADIUS * STANDARD_GALACTIC_RADIUS / STANDARD_NUMBER_OF_STAR_SYSTEMS
	galactic_radius := 2
	for galactic_radius*galactic_radius*galactic_radius < volume {
		galactic_radius++
	}
	if setupData.Galaxy.Overrides.UseOverrides {
		g.Log("For %d stars, the galaxy should have a radius of about %d parsecs.", desired_num_stars, galactic_radius)
		galactic_radius = setupData.Galaxy.Overrides.Radius
	}
	if galactic_radius < MIN_RADIUS || galactic_radius > MAX_RADIUS {
		return nil, fmt.Errorf("radius must be between %d and %d parsecs, inclusive", MIN_RADIUS, MAX_RADIUS)
	}
	galactic_diameter := 2 * galactic_radius
	g.Radius = galactic_radius

	// get the number of cubic parsecs within a sphere with a radius of galactic_radius parsecs.
	volume = (4 * 314 * galactic_radius * galactic_radius * galactic_radius) / 300

	// the probability of a star system existing at any particular set of x,y,z coordinates is one in chance_of_star
	chance_of_star := volume / desired_num_stars
	if chance_of_star < 50 {
		return nil, fmt.Errorf("galactic radius is too small for %d stars", desired_num_stars)
	} else if chance_of_star > 3200 {
		return nil, fmt.Errorf("galactic radius is too large for %d stars", desired_num_stars)
	}

	// initialize star location data
	var star_here [MAX_DIAMETER][MAX_DIAMETER]int
	for x := 0; x < galactic_diameter; x++ {
		for y := 0; y < galactic_diameter; y++ {
			star_here[x][y] = -1
		}
	}

	// randomly place stars
	for num_stars := 0; num_stars < desired_num_stars; {
		// generate coordinates randomly
		x, y, z := rnd(galactic_diameter)-1, rnd(galactic_diameter)-1, rnd(galactic_diameter)-1
		// verify the coordinates are within the galactic boundary
		real_x, real_y, real_z := x-galactic_radius, y-galactic_radius, z-galactic_radius
		sq_distance_from_center := (real_x * real_x) + (real_y * real_y) + (real_z * real_z)
		if sq_distance_from_center >= galactic_radius*galactic_radius {
			continue
		}
		// verify that we don't already have a star here
		if _, exists := g.Stars[Coords{x, y, z, 0}.String()]; exists {
			continue
		}
		// add the star at these coordinates
		star, err := GenerateStar(g, x, y, z, g.DNumSpecies)
		if err != nil {
			return nil, err
		}
		star.SystemNumber = num_stars + 1
		g.Stars[star.ID] = star
		g.Translate.IndexToStarID = append(g.Translate.IndexToStarID, star.ID)
		num_stars++
	}

	g.NumberOfStars = len(g.Stars)

	// generate natural wormholes
	minWormholeLength := 20 // galactic_radius + 3 // in parsecs
	//if minWormholeLength > 20 {
	//	minWormholeLength = 20
	//}
	for _, star := range g.AllStars() {
		if star.HomeSystem || star.WormHere || rnd(100) < 92 {
			continue
		}

		// we want to put a wormhole here if we can find a star at least that minimum distance away that doesn't already have a worm hole
		var worm_star *StarData
		for k, f := 0, rnd(desired_num_stars); k < desired_num_stars && worm_star == nil; k++ {
			ps := g.Stars[g.Translate.IndexToStarID[(k+f)%len(g.Translate.IndexToStarID)]]
			if ps == star || ps.HomeSystem || ps.WormHere {
				continue
			}
			// eliminate wormholes less than the minimum
			if star.Coords.DistanceSquaredTo(ps.Coords) < minWormholeLength*minWormholeLength {
				continue
			}
			worm_star = ps
		}
		if worm_star == nil {
			// wow. none of the existing stars met the criteria
			continue
		}

		star.WormHere = true
		star.WormCoords = worm_star.Coords

		worm_star.WormHere = true
		worm_star.Coords = star.Coords

		// todo: consider making a number of the wormholes one-way
		g.NumberOfWormHoles++
	}

	for _, star := range g.Stars {
		g.NumberOfPlanets += len(star.Planets)
	}

	g.Log("This galaxy contains a total of %d stars and %d planets.\n", len(g.Stars), g.NumberOfPlanets)
	if g.NumberOfWormHoles == 1 {
		g.Log("The galaxy contains %d natural wormhole.\n\n", g.NumberOfWormHoles)
	} else {
		g.Log("The galaxy contains %d natural wormholes.\n\n", g.NumberOfWormHoles)
	}

	return g, nil
}

// GetGalaxy loads data from a JSON file.
func GetGalaxy(galaxyPath string) (*GalaxyData, error) {
	data, err := ioutil.ReadFile(filepath.Join(galaxyPath, "galaxy.json"))
	if err != nil {
		return nil, err
	}
	var galaxy GalaxyData
	if err := json.Unmarshal(data, &galaxy); err != nil {
		return nil, err
	}
	return &galaxy, nil
}

func (g *GalaxyData) AddHomePlanets(w io.Writer, galaxyPath, outputPath string, setupData *SetupData, player *PlayerData, s *SpeciesData) error {
	// randomly pick a star that not to close to other species or wormholes
	coords, err := g.GetFirstXYZ(setupData.Galaxy.MinimumDistance, setupData.Galaxy.ForbidNearbyWormholes)
	if err != nil {
		return err
	}
	s.Home.System = &NamedSystem{
		Name: player.HomePlanetName,
		Star: g.GetStarAt(coords),
	}
	if s.Home.System.Star == nil {
		return fmt.Errorf("There is no star at %s", coords)
	}

	// convert the system at those coordinates to a home system

	// fetch the home system template and update the star with values from the template
	s.Home.System.Star.ConvertToHomeSystem(g.Templates.Homes[s.Home.System.Star.NumPlanets])
	pn := s.Home.System.Star.HomePlanetNumber()
	_, _ = fmt.Fprintf(w, "Converted system %s, home planet %d\n", coords, pn)

	// get pointer to home planet
	s.Home.World = &NamedPlanetData{
		ID:   player.HomePlanetName,
		Name: player.HomePlanetName,
	}
	s.Home.World.Planet = s.Home.System.Star.Planets[s.Home.System.Star.HomePlanetIndex()]
	s.AddNamedPlanet(s.Home.World)

	// AddSpecies step in setup_game.py
	_, _ = fmt.Fprintf(w, "Scan of star system:\n\n")
	_ = s.Home.System.Star.Scan(w, nil)
	_, _ = fmt.Fprintf(w, "\n")

	/* Check tech levels. */
	totalTechLevels := 0
	totalTechLevels += player.BI
	totalTechLevels += player.GV
	totalTechLevels += player.LS
	totalTechLevels += player.ML
	if totalTechLevels != 15 {
		_, _ = fmt.Fprintf(w, "\n\tERROR! ML + GV + LS + BI is not equal to 15!\n\n")
		return fmt.Errorf("total tech levels must sum up to 15")
	}
	// set player-specified tech levels (mining and manufacturing are each 10)
	s.TechLevel[BI] = player.BI
	s.TechLevel[GV] = player.GV
	s.TechLevel[LS] = player.LS
	s.TechLevel[MA] = 10
	s.TechLevel[MI] = 10
	s.TechLevel[ML] = player.ML

	// initialize other tech stuff
	for i := MI; i <= BI; i++ {
		j := s.TechLevel[i]
		s.TechKnowledge[i] = j
		s.InitTechLevel[i] = j
		s.TechEps[i] = 0
	}

	// confirm that required gas is present
	s.Gases.Required.Type = O2 // (we're biased towards oxygen breathers?)
	for _, gas := range s.Home.World.Planet.Gases {
		if gas.Type == s.Gases.Required.Type {
			s.Gases.Required.Min = gas.Percentage / 2
			if s.Gases.Required.Min < 1 {
				s.Gases.Required.Min = 1
			}
			s.Gases.Required.Max = 2 * gas.Percentage
			if s.Gases.Required.Max < 20 {
				s.Gases.Required.Max += 20
			} else if s.Gases.Required.Max > 100 {
				// TODO: i prefer 99% for the max
				s.Gases.Required.Max = 100
			}
		}
	}
	if s.Gases.Required.Max == 0 {
		_, _ = fmt.Fprintf(w, "\n\tERROR! Planet does not have %s(%s)!\n", s.Gases.Required.Type.String(), s.Gases.Required.Type.Char())
		return fmt.Errorf("planet does not have required gas %s", s.Gases.Required.Type.Char())
	}

	// all home planet gases are either required or neutral
	num_neutral := len(s.Home.World.Planet.Gases)
	var goodGas [14]bool
	for _, gas := range s.Home.World.Planet.Gases {
		goodGas[gas.Type] = true
	}
	if !goodGas[HE] {
		// Helium must always be neutral since it is a noble gas.
		goodGas[HE] = true
		num_neutral++
	}
	if !goodGas[H2O] {
		// This game is biased towards oxygen breathers, so make H2O neutral also.
		goodGas[H2O] = true
		num_neutral++
	}
	// Start with the good_gas array and add neutral gases until there are exactly seven of them.
	// One of the seven gases will be the required gas.
	for num_neutral < 7 {
		if n := prng.Roll(13); !goodGas[n] {
			goodGas[n] = true
			num_neutral++
		}
	}

	// add the neutral and poison gases
	for n := 1; n <= 13; n++ {
		t := GasType(n)
		if !goodGas[n] {
			s.Gases.Poison = append(s.Gases.Poison, t)
		} else if t != s.Gases.Required.Type { // required gas isn't neutral!
			s.Gases.Neutral = append(s.Gases.Neutral, t)
		}
	}

	// Do mining and manufacturing bases of home planet.
	// Initial mining and production capacity will be 25 times sum of MI and MA plus a small random amount.
	// Mining and manufacturing base will be reverse-calculated from the capacity.
	levels := s.TechLevel[MI] + s.TechLevel[MA]
	n := (25 * levels) + prng.Roll(levels) + prng.Roll(levels) + prng.Roll(levels)
	s.Home.World.MIBase = (n * s.Home.World.Planet.MiningDifficulty) / (10 * s.TechLevel[MI])
	s.Home.World.MABase = (10 * n) / s.TechLevel[MA]

	// initialize contact/ally/enemy masks
	s.Contact = make([]bool, g.DNumSpecies+1, g.DNumSpecies+1)
	s.Ally = make([]bool, g.DNumSpecies+1, g.DNumSpecies+1)
	s.Enemy = make([]bool, g.DNumSpecies+1, g.DNumSpecies+1)

	s.Home.World.Status.HomePlanet = true
	s.Home.World.Status.Populated = true
	s.Home.World.PopUnits = HP_AVAILABLE_POP
	s.Home.World.Shipyards = 1

	/* Print summary. */
	_, _ = fmt.Fprintf(w, "\n  Summary for species #%d:\n", s.Number)
	_, _ = fmt.Fprintf(w, "\tName of species: %s\n", s.Name)
	_, _ = fmt.Fprintf(w, "\tName of home planet: %s\n", s.Home.World.Name)
	_, _ = fmt.Fprintf(w, "\t\tCoordinates: %s #%d\n", s.Home.World.Planet.Coords, s.Home.World.Planet.Coords.Orbit)
	_, _ = fmt.Fprintf(w, "\tName of government: %s\n", s.Government.Name)
	_, _ = fmt.Fprintf(w, "\tType of government: %s\n\n", s.Government.Type)

	_, _ = fmt.Fprintf(w, "\tTech levels: %s = %d,  %s = %d,  %s = %d\n", techData[MI].name, s.TechLevel[MI], techData[MA].name, s.TechLevel[MA], techData[ML].name, s.TechLevel[ML])
	_, _ = fmt.Fprintf(w, "\t             %s = %d,  %s = %d,  %s = %d\n", techData[GV].name, s.TechLevel[GV], techData[LS].name, s.TechLevel[LS], techData[BI].name, s.TechLevel[BI])

	_, _ = fmt.Fprintf(w, "\n\n\tFor this species, the required gas is %s (%d%%-%d%%).\n", s.Gases.Required.Type.Char(), s.Gases.Required.Min, s.Gases.Required.Max)

	_, _ = fmt.Fprintf(w, "\tGases neutral to species:")
	for _, gasType := range s.Gases.Neutral {
		_, _ = fmt.Fprintf(w, " %s ", gasType.Char())
	}

	_, _ = fmt.Fprintf(w, "\n\tGases poisonous to species:")
	for _, gasType := range s.Gases.Poison {
		_, _ = fmt.Fprintf(w, " %s ", gasType.Char())
	}

	_, _ = fmt.Fprintf(w, "\n\n\tInitial mining base = %d.%d. Initial manufacturing base = %d.%d.\n", s.Home.World.MIBase/10, s.Home.World.MIBase%10, s.Home.World.MABase/10, s.Home.World.MABase%10)
	_, _ = fmt.Fprintf(w, "\tIn the first turn, %d raw material units will be produced,\n", (10*s.TechLevel[MI]*s.Home.World.MIBase)/s.Home.World.Planet.MiningDifficulty)
	_, _ = fmt.Fprintf(w, "\tand the total production capacity will be %d.\n\n", (s.TechLevel[MA]*s.Home.World.MABase)/10)

	// set visited_by bit in star data
	s.Home.System.Star.VisitedBy[s.ID] = true

	/* Create log file for first turn. Write home star system data to it. */
	speciesLogFile := filepath.Join(outputPath, fmt.Sprintf("sp%02d.log.txt", s.Number))
	wl, err := os.Create(speciesLogFile)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(wl, "\nScan of home star system for SP %s:\n\n", s.Name)
	_ = s.Home.System.Star.Scan(wl, s)
	_, _ = fmt.Fprintf(wl, "\n")

	_, _ = fmt.Fprintf(w, "Created file %q\n", speciesLogFile)

	return nil
}

func (g *GalaxyData) AddSpecies(s *SpeciesData) {
	g.Translate.SpeciesNameToID[s.Name] = s.ID
	g.Translate.IndexToSpeciesID = append(g.Translate.IndexToSpeciesID, s.ID)
	g.Species[s.ID] = s
	g.allSpecies = append(g.allSpecies, s)
}

// AlienIsVisible returns true only if:
//     the alien has a ship or starbase here that is in orbit or in deep space
//  or both species have a colony on the same planet
//  or the alien has a colony in the system that is not hidden
func (g *GalaxyData) AlienIsVisible(species, alien *SpeciesData, coords Coords) bool {
	for _, ship := range alien.Ships {
		if !coords.SameSystem(ship.Coords) {
			continue
		} else if ship.ItemQuantity[FD] == ship.Tonnage {
			// TODO: whatever FD is, it helps hide the ship?
			continue
		}
		if ship.Status.InOrbit || ship.Status.InDeepSpace {
			return true
		}
	}

	/* Check if alien has a planet that is not hidden. */
	for _, alien_nampla := range alien.NamedPlanets {
		if !coords.SameSystem(coords) {
			continue
		} else if !alien_nampla.Status.Populated {
			continue
		}
		if !alien_nampla.Hidden {
			return true
		}

		/* The colony is hidden. See if we have population on the same planet. */
		for _, nampla := range species.NamedPlanets {
			if !nampla.Planet.Coords.SamePlanet(alien_nampla.Planet.Coords) {
				continue
			} else if !nampla.Status.Populated {
				continue
			}
			/* We have population on the same planet, so the alien cannot hide. */
			return true
		}
	}

	return false
}

func (g *GalaxyData) AllPlanets() []*PlanetData {
	if g.allPlanets == nil {
		for _, star := range g.AllStars() {
			for _, planet := range star.Planets {
				g.allPlanets = append(g.allPlanets, planet)
			}
		}
	}
	return g.allPlanets
}

func (g *GalaxyData) AllSpecies() []*SpeciesData {
	if g.allSpecies == nil {
		for _, id := range g.Translate.IndexToSpeciesID {
			g.allSpecies = append(g.allSpecies, g.Species[id])
		}
	}
	return g.allSpecies
}

func (g *GalaxyData) AllStars() []*StarData {
	stars := g.allStars
	if len(stars) != len(g.Translate.IndexToStarID) {
		stars = make([]*StarData, len(g.Translate.IndexToStarID), len(g.Translate.IndexToStarID))
		for i, id := range g.Translate.IndexToStarID {
			stars[i] = g.GetStarByID(id)
		}
	}
	return stars
}

// GetFirstXYZ returns the first system that is not a home system
// or has a worm hole or is within a given distance of any other home
// system.
func (g *GalaxyData) GetFirstXYZ(d int, forbidWormHoles bool) (Coords, error) {
	minDSquared := d * d
	var forbiddenSystems []*StarData
	for _, star := range g.AllStars() {
		if star.HomeSystem || (star.WormHere && forbidWormHoles) {
			forbiddenSystems = append(forbiddenSystems, star)
		}
	}
	for _, origin := range g.AllStars() {
		if origin == nil || origin.HomeSystem || origin.WormHere || origin.NumPlanets < 3 {
			continue
		}
		nearForbiddenSystem := false
		for _, star := range forbiddenSystems {
			if origin.Coords.DistanceSquaredTo(star.Coords) < minDSquared {
				nearForbiddenSystem = true
				break
			}
		}
		if !nearForbiddenSystem {
			return origin.Coords, nil
		}
	}
	return Coords{}, fmt.Errorf("all suitable systems are within %d parsecs of each other", d)
}

func (g *GalaxyData) GetNamedPlanet(s *SpeciesData, name string) *NamedPlanetData {
	return s.GetNamedPlanet(name)
}

func (g *GalaxyData) GetPlanet(coords Coords) *PlanetData {
	for _, p := range g.AllPlanets() {
		if coords.SamePlanet(p.Coords) {
			return p
		}
	}
	return nil
}

func (g *GalaxyData) GetSpeciesByDistortedID(id int) *SpeciesData {
	for _, s := range g.AllSpecies() {
		if id == s.Distorted() {
			return s
		}
	}
	return nil // not a legitimate species
}

func (g *GalaxyData) GetSpeciesByID(id string) *SpeciesData {
	return g.Species[id]
}

func (g *GalaxyData) GetSpeciesByName(name string) *SpeciesData {
	id, ok := g.Translate.SpeciesNameToID[name]
	if !ok {
		return nil
	}
	return g.GetSpeciesByID(id)
}

func (g *GalaxyData) GetSpeciesByNumber(n int) *SpeciesData {
	for _, s := range g.AllSpecies() {
		if n == s.Number {
			return s
		}
	}
	return nil
}

func (g *GalaxyData) GetStarAt(c Coords) *StarData {
	return g.Stars[c.String()]
}

func (g *GalaxyData) GetStarByID(id string) *StarData {
	return g.Stars[id]
}

func (g *GalaxyData) List(listPlanets, listWormholes bool) error {
	// initialize counts
	total_planets := 0
	total_wormstars := 0
	var type_count [10]int
	for i := DWARF; i <= GIANT; i++ {
		type_count[i] = 0
	}

	// for each star, list info
	for _, star := range g.AllStars() {
		if !listWormholes {
			if listPlanets {
				fmt.Printf("System #%d:\t", star.SystemNumber)
			}
			fmt.Printf("x = %d\ty = %d\tz = %d", star.Coords.X, star.Coords.Y, star.Coords.Z)
			fmt.Printf("\tstellar type = %s%s%s", star.Type.Char(), star.Color.Char(), StarSizeChar[star.Size])
			if listPlanets {
				fmt.Printf("\t%d planets.", star.NumPlanets)
			}
			fmt.Printf("\n")

			if star.NumPlanets == 0 {
				fmt.Printf("\tStar #%d went nova!", star.SystemNumber)
				fmt.Printf(" All planets were blown away!\n")
			} else if star.NumPlanets != len(star.Planets) {
				return fmt.Errorf("assert(numPlanets == lenPlanets)")
			}
		}

		total_planets += star.NumPlanets
		type_count[star.Type] += 1

		if star.WormHere {
			total_wormstars++
			if listPlanets {
				fmt.Printf("!!! Natural wormhole from here to %s\n", star.WormCoords)
			} else if listWormholes {
				fmt.Printf("Wormhole #%d: from %s to %s\n", total_wormstars, star.Coords, star.WormCoords)
				// turn off the target's worm flag to avoid double-reporting
				wormSystem := g.GetStarAt(star.WormCoords)
				if wormSystem != nil {
					wormSystem.WormHere = false
				}
			}
		}

		var home_planet *PlanetData
		if listPlanets {
			/* Check if system has a home planet. */
			for _, planet := range star.Planets {
				if planet.Special == IDEAL_HOME_PLANET || planet.Special == IDEAL_COLONY_PLANET {
					home_planet = planet
					break
				}
			}
		}

		if listPlanets {
			for i, planet := range star.Planets {
				switch planet.Special {
				case NOT_SPECIAL:
					fmt.Printf("     ")
				case IDEAL_HOME_PLANET:
					fmt.Printf(" HOM ")
				case IDEAL_COLONY_PLANET:
					fmt.Printf(" COL ")
				case RADIOACTIVE_HELLHOLE:
					fmt.Printf("     ")
				}
				fmt.Printf("#%d dia=%3d g=%d.%02d tc=%2d pc=%2d md=%d.%02d", i,
					planet.Diameter,
					planet.Gravity/100,
					planet.Gravity%100,
					planet.TemperatureClass,
					planet.PressureClass,
					planet.MiningDifficulty/100,
					planet.MiningDifficulty%100)

				if home_planet != nil {
					fmt.Printf("%4d ", LSN(planet, home_planet))
				} else {
					fmt.Printf("  ")
				}

				num_gases := len(planet.Gases)
				for i, gas := range planet.Gases {
					if gas.Percentage > 0 {
						if i > 0 {
							fmt.Printf(",")
						}
						fmt.Printf("%s(%d%%)", gas.Type.Char(), gas.Percentage)
					}
				}
				if num_gases == 0 {
					fmt.Printf("No atmosphere")
				}

				fmt.Printf("\n")
			}
		}

		if listPlanets {
			fmt.Printf("\n")
		}
	}

	if !listWormholes {
		fmt.Printf("The galaxy has a radius of %d parsecs.\n", g.Radius)
		fmt.Printf("It contains %d dwarf stars, %d degenerate stars, %d main sequence stars,\n", type_count[DWARF], type_count[DEGENERATE], type_count[MAIN_SEQUENCE])
		fmt.Printf("    and %d giant stars, for a total of %d stars.\n", type_count[GIANT], g.NumberOfStars)
		if listPlanets {
			fmt.Printf("The total number of planets in the galaxy is %d.\n", total_planets)
			fmt.Printf("The total number of natural wormholes in the galaxy is %d.\n", total_wormstars/2)
			fmt.Printf("The galaxy was designed for %d species.\n", g.DNumSpecies)
			fmt.Printf("A total of %d species have been designated so far.\n\n", g.NumSpecies)
		}
	}

	/* Internal test. */
	if g.NumberOfPlanets != total_planets {
		return fmt.Errorf("WARNING!  Program error!  Internal inconsistency!")
	}

	return nil
}

func (g *GalaxyData) Log(format string, a ...interface{}) {
	if g != nil && g.logFile != nil {
		if _, err := fmt.Fprintf(g.logFile, format, a...); err != nil {
			panic(err)
		}
	}
}

func (g *GalaxyData) MakeHomeTemplates(w io.Writer) error {
	for num_planets := 3; num_planets < 10; num_planets++ {
		_, _ = fmt.Fprintf(w, "Creating home system with %d planets...\n", num_planets)
		var planets []*PlanetData
		for planets == nil {
			planets = GenerateEarthLikePlanet(fmt.Sprintf("homes/%02d", num_planets), num_planets)
		}
		g.Templates.Homes[num_planets] = planets
	}

	return nil
}

func (g *GalaxyData) Scan(w io.Writer, c Coords) error {
	star := g.GetStarAt(c)
	if star == nil {
		_, _ = fmt.Fprintf(w, "Scan Report: There is no star system at x = %d, y = %d, z = %d.\n", c.X, c.Y, c.Z)
		return nil
	}
	return star.Scan(w, nil)
}

func (g *GalaxyData) Write(outputPath string, isVerbose bool) error {
	galaxyFile := filepath.Join(outputPath, "galaxy.json")
	if isVerbose {
		fmt.Printf("[galaxy] %-30s == %q\n", "GALAXY_FILE", galaxyFile)
	}
	if b, err := json.MarshalIndent(g, "  ", "  "); err != nil {
		return err
	} else if err := ioutil.WriteFile(galaxyFile, b, 0644); err != nil {
		return err
	}

	return nil
}
