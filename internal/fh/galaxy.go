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
	"path"
	"strconv"
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
	TurnNumber        int
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
}

type Player struct {
	ID           string `json:"id"`
	EmailAddress string `json:"email"`
	Species      string `json:"species"`
}

func GenerateGalaxy(logFile io.Writer, setupData *SetupData) (*GalaxyData, error) {
	galaxy := &GalaxyData{
		ID:      setupData.Galaxy.Name,
		Name:    setupData.Galaxy.Name,
		Secret:  "your-private-key-belongs-here",
		Players: make(map[string]*Player),
		Species: make(map[string]*SpeciesData),
		Stars:   make(map[string]*StarData),
	}
	galaxy.Translate.EmailToID = make(map[string]string)
	galaxy.Translate.SpeciesNameToID = make(map[string]string)
	galaxy.Translate.XYZToID = make(map[string]string)

	// initialize from some player data
	for _, player := range setupData.Players {
		id := player.Email
		galaxy.Players[id] = &Player{
			ID:           id,
			EmailAddress: player.Email,
			Species:      player.SpeciesName,
		}
		galaxy.Translate.EmailToID[id] = player.Email
	}

	// init?
	d_num_species := len(setupData.Players)
	if d_num_species < MIN_SPECIES || MAX_SPECIES < d_num_species {
		return nil, fmt.Errorf("number of species must be between %d and %d, inclusive", MIN_SPECIES, MAX_SPECIES)
	}
	adjusted_number_of_species := d_num_species
	if setupData.Galaxy.LowDensity {
		// add 50% more species to the mix as a way to trick the program into adding more stars
		adjusted_number_of_species = (d_num_species * 3) / 2
		if adjusted_number_of_species < MIN_SPECIES || MAX_SPECIES < adjusted_number_of_species {
			_, _ = fmt.Fprintf(logFile, "Low density option giving, boosting species count to %d\n", adjusted_number_of_species)
			return nil, fmt.Errorf("adjusted number of species must be between %d and %d, inclusive", MIN_SPECIES, MAX_SPECIES)
		}
	}
	galaxy.DNumSpecies = d_num_species

	// get approximate number of star systems to generate
	desired_num_stars := (adjusted_number_of_species * STANDARD_NUMBER_OF_STAR_SYSTEMS) / STANDARD_NUMBER_OF_SPECIES
	_, _ = fmt.Fprintf(logFile, "For %d species, there should be about %d stars.\n", d_num_species, desired_num_stars)
	if setupData.Galaxy.Overrides.UseOverrides {
		if setupData.Galaxy.LowDensity {
			_, _ = fmt.Fprintf(logFile, "For %d species, a low density game needs about %d stars.\n", d_num_species, desired_num_stars)
		} else {
			_, _ = fmt.Fprintf(logFile, "For %d species, a game needs about %d stars.\n", d_num_species, desired_num_stars)
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
		_, _ = fmt.Fprintf(logFile, "For %d stars, the galaxy should have a radius of about %d parsecs.", desired_num_stars, galactic_radius)
		galactic_radius = setupData.Galaxy.Overrides.Radius
	}
	if galactic_radius < MIN_RADIUS || galactic_radius > MAX_RADIUS {
		return nil, fmt.Errorf("radius must be between %d and %d parsecs, inclusive", MIN_RADIUS, MAX_RADIUS)
	}
	galactic_diameter := 2 * galactic_radius
	galaxy.Radius = galactic_radius

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
		if _, exists := galaxy.Stars[Coords{x, y, z, 0}.String()]; exists {
			continue
		}
		// add the star at these coordinates
		star, err := GenerateStar(x, y, z, galaxy.DNumSpecies)
		if err != nil {
			return nil, err
		}
		star.SystemNumber = num_stars + 1
		galaxy.Stars[star.ID] = star
		galaxy.Translate.IndexToStarID = append(galaxy.Translate.IndexToStarID, star.ID)
		num_stars++
	}

	galaxy.NumberOfStars = len(galaxy.Stars)

	// generate natural wormholes
	minWormholeLength := 20 // galactic_radius + 3 // in parsecs
	//if minWormholeLength > 20 {
	//	minWormholeLength = 20
	//}
	for _, star := range galaxy.AllStars() {
		if star.HomeSystem || star.WormHere || rnd(100) < 92 {
			continue
		}

		// we want to put a wormhole here if we can find a star at least that minimum distance away that doesn't already have a worm hole
		var worm_star *StarData
		for k, f := 0, rnd(desired_num_stars); k < desired_num_stars && worm_star == nil; k++ {
			ps := galaxy.Stars[galaxy.Translate.IndexToStarID[(k+f)%len(galaxy.Translate.IndexToStarID)]]
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
		galaxy.NumberOfWormHoles++
	}

	for _, star := range galaxy.Stars {
		galaxy.NumberOfPlanets += len(star.Planets)
	}

	_, _ = fmt.Fprintf(logFile, "This galaxy contains a total of %d stars and %d planets.\n", len(galaxy.Stars), galaxy.NumberOfPlanets)
	if galaxy.NumberOfWormHoles == 1 {
		_, _ = fmt.Fprintf(logFile, "The galaxy contains %d natural wormhole.\n\n", galaxy.NumberOfWormHoles)
	} else {
		_, _ = fmt.Fprintf(logFile, "The galaxy contains %d natural wormholes.\n\n", galaxy.NumberOfWormHoles)
	}

	return galaxy, nil
}

// GetGalaxy loads data from a JSON file.
func GetGalaxy(name string) (*GalaxyData, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var galaxy GalaxyData
	if err := json.Unmarshal(data, &galaxy); err != nil {
		return nil, err
	}
	return &galaxy, nil
}

func (g *GalaxyData) AddHomePlanets(w io.Writer, galaxyPath string, setupData *SetupData, player *PlayerData, s *SpeciesData) error {
	home_nampla := &NamedPlanetData{ID: player.HomePlanetName}
	home_nampla.Name = player.HomePlanetName
	s.AddNamedPlanet(home_nampla)

	s.HomeNampla = home_nampla.ID
	s.GovtName = player.GovName
	s.GovtType = player.GovType

	// HomeSystemAuto step in setup_game.py
	forbidNearbyWormholes := setupData.Galaxy.ForbidNearbyWormholes
	minDistance := setupData.Galaxy.MinimumDistance
	coords, err := g.GetFirstXYZ(minDistance, forbidNearbyWormholes)
	if err != nil {
		return err
	}
	// convert the system at those coordinates to a home system
	star := g.GetStarAt(coords)
	if star == nil {
		return fmt.Errorf("There is no star at %s", coords)
	}
	// fetch the home system template and update the star with values from the template
	star.ConvertToHomeSystem(g.Templates.Homes[star.NumPlanets])
	pn := star.HomePlanetNumber()
	_, _ = fmt.Fprintf(w, "Converted system %s, home planet %d\n", coords, pn)

	// get pointer to home planet
	s.Home.Planet = star.Planets[star.HomePlanetIndex()]

	// AddSpecies step in setup_game.py
	s.Home.Coords = Coords{coords.X, coords.Y, coords.Z, coords.Orbit}
	home_nampla.Coords = Coords{coords.X, coords.Y, coords.Z, coords.Orbit}

	_, _ = fmt.Fprintf(w, "Scan of star system:\n\n")
	star.Scan(os.Stdout, nil)
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
	for _, gas := range s.Home.Planet.Gases {
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
	num_neutral := len(s.Home.Planet.Gases)
	var goodGas [14]bool
	for _, gas := range s.Home.Planet.Gases {
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
	home_nampla.MIBase = (n * s.Home.Planet.MiningDifficulty) / (10 * s.TechLevel[MI])
	home_nampla.MABase = (10 * n) / s.TechLevel[MA]

	// initialize contact/ally/enemy masks
	s.Contact = make([]bool, g.DNumSpecies+1, g.DNumSpecies+1)
	s.Ally = make([]bool, g.DNumSpecies+1, g.DNumSpecies+1)
	s.Enemy = make([]bool, g.DNumSpecies+1, g.DNumSpecies+1)

	s.NumNamplas = 1 // just the home planet for now ("nampla" means "named planet")
	home_nampla.Status.HomePlanet = true
	home_nampla.Status.Populated = true
	home_nampla.PopUnits = HP_AVAILABLE_POP
	home_nampla.Shipyards = 1

	/* Print summary. */
	_, _ = fmt.Fprintf(w, "\n  Summary for species #%d:\n", s.Number)
	_, _ = fmt.Fprintf(w, "\tName of species: %s\n", s.Name)
	_, _ = fmt.Fprintf(w, "\tName of home planet: %s\n", home_nampla.Name)
	_, _ = fmt.Fprintf(w, "\t\tCoordinates: %s #%d\n", s.Home.Coords, s.Home.Coords.Orbit)
	_, _ = fmt.Fprintf(w, "\tName of government: %s\n", s.GovtName)
	_, _ = fmt.Fprintf(w, "\tType of government: %s\n\n", s.GovtType)

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

	_, _ = fmt.Fprintf(w, "\n\n\tInitial mining base = %d.%d. Initial manufacturing base = %d.%d.\n", home_nampla.MIBase/10, home_nampla.MIBase%10, home_nampla.MABase/10, home_nampla.MABase%10)
	_, _ = fmt.Fprintf(w, "\tIn the first turn, %d raw material units will be produced,\n", (10*s.TechLevel[MI]*home_nampla.MIBase)/s.Home.Planet.MiningDifficulty)
	_, _ = fmt.Fprintf(w, "\tand the total production capacity will be %d.\n\n", (s.TechLevel[MA]*home_nampla.MABase)/10)

	// set visited_by bit in star data
	star.VisitedBy[s.ID] = true

	/* Create log file for first turn. Write home star system data to it. */
	logFile := path.Join(galaxyPath, fmt.Sprintf("sp%02d.log", s.Number))
	wl, err := os.Create(logFile)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "\nScan of home star system for SP %s:\n\n", s.Name)
	_ = star.Scan(wl, s)
	_, _ = fmt.Fprintf(w, "\n")

	_, _ = fmt.Fprintf(w, "Created file %q\n", logFile)

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
			if !nampla.Coords.SamePlanet(alien_nampla.Coords) {
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

// Finish completes a turn
func (g *GalaxyData) Finish(w io.Writer, galaxyPath string, test_mode, verbose_mode bool) error {
	if verbose_mode {
		_, _ = fmt.Fprintf(w, "\nFinishing up for all species...\n")
	}

	l := &Logger{Stdout: os.Stdout}

	var header_printed bool
	print_header := func() {
		l.String("\nOther events:\n")
		header_printed = true
	}

	transaction, err := GetTransactionData(galaxyPath)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Loaded %d transactions\n", len(transaction))
		return err
	}

	// bump the turn number
	g.TurnNumber++
	turnPath := path.Join(galaxyPath, fmt.Sprintf("t%06d", g.TurnNumber))

	// Total economic base includes all of the colonies on the planet, not just the one species.
	total_econ_base := make([]int, len(g.AllPlanets()), len(g.AllPlanets()))

	// add mining difficulty increases for each planet, use the increase calcuated on the prior turn
	for _, planet := range g.AllPlanets() {
		planet.MiningDifficulty += planet.MDIncrease
		planet.MDIncrease = 0
	}

	/* Main loop. For each species, take appropriate action. */
	for _, species := range g.AllSpecies() {
		// display name of species
		if verbose_mode {
			_, _ = fmt.Fprintf(w, "  Now doing SP %s...", species.Name)
		}

		// check if player submitted orders for this turn.
		orderFile := path.Join(turnPath, fmt.Sprintf("sp%02d.ord", species.Number))
		_, err = ioutil.ReadFile(orderFile)
		orders_received := err == nil || g.TurnNumber == 1
		if verbose_mode {
			if !orders_received {
				_, _ = fmt.Fprintf(w, " WARNING: player did not submit orders this turn!")
			}
			_, _ = fmt.Fprintf(w, "\n")
		}

		// open log file
		var err error
		l.File, err = os.Create(path.Join(turnPath, fmt.Sprintf("sp%02d.log", species.Number)))
		if err != nil {
			return err
		}
		l.Stdout = nil

		// TODO: try to get straight on Turn 0 being setup and Turn 1 being first turn orders are processed
		check := struct {
			mishaps            bool /* Check if any ships of this species experienced mishaps. */
			disbanded          bool /* Take care of any disbanded colonies. */
			transferInEU       bool /* Check if this species is the recipient of a transfer of economic units from another species. */
			jumpPortalsUsed    bool /* Check if any jump portals of this species were used by aliens. */
			detectedTelescopes bool /* Check if any starbases of this species detected the use of gravitic telescopes by aliens. */
			transferInTL       bool /* Check if this species is the recipient of a tech transfer from another species. */
			increaseTL         bool /* Calculate tech level increases. */
			transferInKN       bool /* Check if this species is the recipient of a knowledge transfer from another species. */
			loopNamedPlanets   bool /* Loop through each nampla for this species. */
			loopShips          bool /* Loop through all ships for this species. */
			alienIncursion     bool /* Check if this species has a populated planet that another species tried to land on. */
			alienConstruction  bool /* Check if this species is the recipient of interspecies construction. */
			besiegingOthers    bool /* Check if this species is besieging another species and detects forbidden construction, landings, etc. */
			messages           bool // check if this species is the recipient of a message from another species
		}{
			mishaps:            g.TurnNumber > 1,
			disbanded:          g.TurnNumber > 1,
			transferInEU:       g.TurnNumber > 1,
			jumpPortalsUsed:    g.TurnNumber > 1,
			detectedTelescopes: g.TurnNumber > 1,
			transferInTL:       g.TurnNumber > 1,
			increaseTL:         g.TurnNumber > 1,
			transferInKN:       g.TurnNumber > 1,
			loopNamedPlanets:   g.TurnNumber > 1,
			loopShips:          g.TurnNumber > 1,
			alienIncursion:     g.TurnNumber > 1,
			alienConstruction:  g.TurnNumber > 1,
			besiegingOthers:    g.TurnNumber > 1,
			messages:           true, // always check messages
		}

		/* Check if any ships of this species experienced mishaps. */
		if check.mishaps {
			for _, t := range transaction {
				if t.Type == SHIP_MISHAP && t.Number1 == species.Number {
					if !header_printed {
						print_header()
					}
					l.String("  !!! ")
					l.String(t.Name1)
					if t.Value < 3 {
						/* Intercepted or self-destructed. */
						l.String(" disappeared without a trace, cause unknown!\n")
					} else if t.Value == 3 {
						/* Mis-jumped. */
						l.String(" mis-jumped to ")
						l.Int(t.X)
						l.Char(' ')
						l.Int(t.Y)
						l.Char(' ')
						l.Int(t.Z)
						l.String("!\n")
					} else {
						/* One fail-safe jump unit used. */
						l.String(" had a jump mishap! A fail-safe jump unit was expended.\n")
					}
				}
			}
		}

		/* Take care of any disbanded colonies. */
		if check.disbanded {
			var coloniesDestroyed, shipsDestroyed int
			for _, nampla := range species.NamedPlanets {
				if !nampla.Status.DisbandedColony {
					continue
				}

				/* Salvage ships on the surface and starbases in orbit. */
				salvage_EUs := 0
				for _, ship := range species.Ships {
					if !nampla.Coords.SameSystem(ship.Coords) {
						continue
					}
					if ship.Status.InOrbit && ship.Type != STARBASE {
						continue
					}

					/* Transfer cargo to planet. */
					for i := 0; i < MAX_ITEMS; i++ {
						nampla.ItemQuantity[i] += ship.ItemQuantity[i]
					}

					/* Salvage the ship. */
					original_cost := shipData[ship.Class].cost
					if ship.Class == TR || ship.Type == STARBASE {
						original_cost *= ship.Tonnage
					}

					if ship.Type == SUB_LIGHT {
						original_cost = (3 * original_cost) / 4
					}

					var salvage_value int
					if ship.Status.UnderConstruction {
						salvage_value = (original_cost - ship.RemainingCost) / 4
					} else {
						salvage_value = (3 * original_cost * (60 - ship.Age)) / 400
					}

					salvage_EUs += salvage_value

					/* Destroy the ship. */
					ship.Status.Destroyed = true
					shipsDestroyed++
				}

				/* Salvage items on the planet. */
				for i := 0; i < MAX_ITEMS; i++ {
					var salvage_value int
					if i == RM {
						salvage_value = nampla.ItemQuantity[RM] / 10
					} else if nampla.ItemQuantity[i] > 0 {
						original_cost := nampla.ItemQuantity[i] * itemData[i].cost
						if i == TP {
							if species.TechLevel[BI] > 0 {
								original_cost /= species.TechLevel[BI]
							} else {
								original_cost /= 100
							}
						}
						salvage_value = original_cost / 4
					} else {
						salvage_value = 0
					}

					salvage_EUs += salvage_value
				}

				/* Transfer EUs to species. */
				species.EconUnits += salvage_EUs

				/* Log what happened. */
				if !header_printed {
					print_header()
				}
				l.String("  PL ")
				l.String(nampla.Name)
				l.String(" was disbanded, generating ")
				l.Long(salvage_EUs)
				l.String(" economic units in salvage.\n")

				coloniesDestroyed++
			}

			// destroy the disbanded colonies
			if coloniesDestroyed != 0 {
				var namedPlanets []*NamedPlanetData
				for _, nampla := range species.NamedPlanets {
					if !nampla.Status.DisbandedColony {
						continue
					}
					namedPlanets = append(namedPlanets, nampla)
				}
				species.NamedPlanets = namedPlanets
			}

			// destroy the salvaged ships
			if shipsDestroyed != 0 {
				var ships []*ShipData
				for _, ship := range species.Ships {
					if !ship.Status.Destroyed {
						ships = append(ships, ship)
					}
				}
				species.Ships = ships
			}
		}

		/* Check if this species is the recipient of a transfer of economic units from another species. */
		if check.transferInEU {
			for _, t := range transaction {
				if t.Recipient == species.Number && (t.Type == EU_TRANSFER || t.Type == SIEGE_EU_TRANSFER || t.Type == LOOTING_EU_TRANSFER) {
					// Transfer EUs to attacker if this is a siege or looting transfer.
					// If this is a normal transfer, then just log the result since the actual transfer was done when the order was processed.
					if t.Type != EU_TRANSFER {
						species.EconUnits += t.Value
					}

					if !header_printed {
						print_header()
					}
					l.String("  ")
					l.Long(t.Value)
					l.String(" economic units were received from SP ")
					l.String(t.Name1)
					if t.Type == SIEGE_EU_TRANSFER {
						l.String(" as a result of your successful siege of their PL ")
						l.String(t.Name3)
						l.String(". The siege was ")
						l.Long(t.Number1)
						l.String("% effective")
					} else if t.Type == LOOTING_EU_TRANSFER {
						l.String(" as a result of your looting their PL ")
						l.String(t.Name3)
					}
					l.String(".\n")
				}
			}
		}

		/* Check if any jump portals of this species were used by aliens. */
		if check.jumpPortalsUsed {
			for _, t := range transaction {
				if t.Type == ALIEN_JUMP_PORTAL_USAGE && t.Number1 == species.Number {
					if !header_printed {
						print_header()
					}
					l.String("  ")
					l.String(t.Name1)
					l.Char(' ')
					l.String(t.Name2)
					l.String(" used jump portal ")
					l.String(t.Name3)
					l.String(".\n")
				}
			}
		}

		/* Check if any starbases of this species detected the use of gravitic telescopes by aliens. */
		if check.detectedTelescopes {
			for _, t := range transaction {
				if !(t.Type == TELESCOPE_DETECTION && t.Number1 == species.Number) {
					continue
				}
				if !header_printed {
					print_header()
				}
				l.String("! ")
				l.String(t.Name1)
				l.String(" detected the operation of an alien gravitic telescope at x = ")
				l.Int(t.X)
				l.String(", y = ")
				l.Int(t.Y)
				l.String(", z = ")
				l.Int(t.Z)
				l.String(".\n")
			}
		}

		/* Check if this species is the recipient of a tech transfer from another species. */
		if check.transferInTL {
			for _, t := range transaction {
				if !(t.Type == TECH_TRANSFER && t.Recipient == species.Number) {
					continue
				}

				/* Try to transfer technology. */
				//rec := t.Recipient - 1
				don := t.Donor - 1

				if !header_printed {
					print_header()
				}
				l.String("  ")
				tech := t.Value
				l.String(techData[tech].name)
				l.String(" tech transfer from SP ")
				l.String(t.Name1)
				their_level := t.Number3
				my_level := species.TechLevel[tech]

				if their_level <= my_level {
					l.String(" failed.\n")
					t.Number1 = -1
					continue
				}

				donor_species := g.GetSpeciesByNumber(don)
				actual_cost, max_cost := 0, t.Number1
				if max_cost == 0 {
					max_cost = donor_species.EconUnits
				} else if donor_species.EconUnits < max_cost {
					max_cost = donor_species.EconUnits
				}
				new_level := my_level
				for new_level < their_level {
					one_point_cost := new_level * new_level
					one_point_cost -= one_point_cost / 4 /* 25% discount. */
					if (actual_cost + one_point_cost) > max_cost {
						break
					}
					actual_cost += one_point_cost
					new_level++
				}

				if new_level == my_level {
					l.String(" failed due to lack of funding.\n")
					t.Number1 = -2
				} else {
					l.String(" raised your tech level from ")
					l.Int(my_level)
					l.String(" to ")
					l.Int(new_level)
					l.String(" at a cost to them of ")
					l.Long(actual_cost)
					l.String(".\n")
					t.Number1 = actual_cost
					t.Number2 = my_level
					t.Number3 = new_level

					species.TechLevel[tech] = new_level
					donor_species.EconUnits -= actual_cost
				}
			}
		}

		/* Calculate tech level increases. */
		if check.increaseTL {
			for tech := MI; tech <= BI; tech++ {
				old_tech_level := species.TechLevel[tech]
				new_tech_level := old_tech_level

				var max_tech_level int

				experience_points := species.TechEps[tech]
				if experience_points != 0 {
					/* Determine increase as if there were NO randomness in the process. */
					i := experience_points
					j := old_tech_level
					for i >= j*j {
						i -= j * j
						j++
					}

					// When extremely large amounts are spent on research, tech level increases are sometimes excessive.  Set a limit.
					if old_tech_level > 50 {
						max_tech_level = j + 1
					} else {
						max_tech_level = 9999
					}

					/* Allocate half of the calculated increase NON-RANDOMLY. */
					n := (j - old_tech_level) / 2
					for i = 0; i < n; i++ {
						experience_points -= new_tech_level * new_tech_level
						new_tech_level++
					}

					/* Allocate the rest randomly. */
					for experience_points >= new_tech_level {
						experience_points -= new_tech_level
						n = new_tech_level

						/* The chance of success is 1 in n. At this point, n is always at least 1. */
						i = rnd(16 * n)
						if i >= 8*n && i <= 8*n+15 {
							new_tech_level = n + 1
						}
					}

					/* Save unused experience points. */
					species.TechEps[tech] = experience_points
				}

				/* See if any random increase occurred. Odds are 1 in 6. */
				if old_tech_level > 0 && rnd(6) == 6 {
					new_tech_level++
				}

				if new_tech_level > max_tech_level {
					new_tech_level = max_tech_level
				}

				/* Report result only if tech level went up. */
				if new_tech_level > old_tech_level {
					if !header_printed {
						print_header()
					}
					l.String("  ")
					l.String(techData[tech].name)
					l.String(" tech level rose from ")
					l.Int(old_tech_level)
					l.String(" to ")
					l.Int(new_tech_level)
					l.String(".\n")

					species.TechLevel[tech] = new_tech_level
				}
			}
		}

		/* Notify of any new high tech items. */
		for tech := MI; tech <= BI; tech++ {
			old_tech_level := species.InitTechLevel[tech]
			new_tech_level := species.TechLevel[tech]

			if new_tech_level > old_tech_level {
				check_high_tech_items(tech, old_tech_level, new_tech_level, l)
			}

			species.InitTechLevel[tech] = new_tech_level
		}

		/* Check if this species is the recipient of a knowledge transfer from another species. */
		if check.transferInKN {
			for _, t := range transaction {
				if t.Type == KNOWLEDGE_TRANSFER && t.Recipient == species.Number {
					//rec := t.Recipient - 1
					//don := t.Donor - 1

					/* Try to transfer technology. */
					tech := t.Value
					their_level := t.Number3
					my_level := species.TechLevel[tech]
					n := species.TechKnowledge[tech]
					if n > my_level {
						my_level = n
					}

					if their_level <= my_level {
						continue
					}

					species.TechKnowledge[tech] = their_level

					if !header_printed {
						print_header()
					}
					l.String("  SP ")
					l.String(t.Name1)
					l.String(" transferred knowledge of ")
					l.String(techData[tech].name)
					l.String(" to you up to tech level ")
					l.Long(their_level)
					l.String(".\n")
				}
			}
		}

		/* Loop through each nampla for this species. */
		if check.loopNamedPlanets {
			for _, nampla := range species.NamedPlanets {
				if nampla.Coords.Orbit == 99 {
					continue
				}

				/* Get planet pointer. */
				planet := g.GetPlanet(nampla.Coords)
				if planet == nil {
					panic("assert(planet != nil)")
				}

				/* Clear any amount spent on ambush. */
				nampla.UseOnAmbush = 0

				/* Handle HIDE order. */
				nampla.Hidden = nampla.Hiding
				nampla.Hiding = false

				/* Check if any IUs or AUs were installed. */
				if nampla.IUsToInstall > 0 {
					nampla.MIBase += nampla.IUsToInstall
					nampla.IUsToInstall = 0
				}

				if nampla.AUsToInstall > 0 {
					nampla.MABase += nampla.AUsToInstall
					nampla.AUsToInstall = 0
				}

				/* Check if another species on the same planet has become
				 *  assimilated. */
				for _, t := range transaction {
					if !(t.Type == ASSIMILATION && t.Value == species.Number && nampla.Coords.SamePlanet(Coords{t.X, t.Y, t.Z, t.PN})) {
						continue
					}
					if !header_printed {
						print_header()
					}

					ib, ab, ns := t.Number1, t.Number2, t.Number3
					l.String("  Assimilation of ")
					l.String(t.Name1)
					l.String(" PL ")
					l.String(t.Name2)
					l.String(" increased mining base of ")
					l.String(species.Name)
					l.String(" PL ")
					l.String(nampla.Name)
					l.String(" by ")
					l.Long(ib / 10)
					l.Char('.')
					l.Long(ib % 10)
					l.String(", and manufacturing base by ")
					l.Long(ab / 10)
					l.Char('.')
					l.Long(ab % 10)
					if ns > 0 {
						l.String(". Number of shipyards was also increased by ")
						l.Int(ns)
					}
					l.String(".\n")
				}

				/* Calculate available population for this turn. */
				nampla.PopUnits = 0

				eb := nampla.MIBase + nampla.MABase
				total_pop_units := eb + nampla.ItemQuantity[CU] + nampla.ItemQuantity[PD]

				if nampla.Status.HomePlanet {
					if nampla.Status.Populated {
						nampla.PopUnits = HP_AVAILABLE_POP

						if species.HPOriginalBase != 0 { /* HP was bombed. */
							if eb >= species.HPOriginalBase {
								species.HPOriginalBase = 0 /* Fully recovered. */
							} else {
								nampla.PopUnits = (eb * HP_AVAILABLE_POP) / species.HPOriginalBase
							}
						}
					}
				} else if nampla.Status.Populated {
					/* Get life support tech level needed. */
					ls_needed := species.LifeSupportNeeded(planet)

					/* Basic percent increase is 10*(1 - ls_needed/ls_actual). */
					ls_actual := species.TechLevel[LS]
					percent_increase := 10 * (100 - ((100 * ls_needed) / ls_actual))

					if percent_increase < 0 { /* Colony wiped out! */
						if !header_printed {
							print_header()
						}

						l.String("  !!! Life support tech level was too low to support colony on PL ")
						l.String(nampla.Name)
						l.String(". Colony was destroyed.\n")

						/* No longer populated or self-sufficient. */
						nampla.Status = NamedPlanetStatus{Colony: true}
						nampla.MIBase = 0
						nampla.MABase = 0
						nampla.PopUnits = 0
						nampla.ItemQuantity[PD] = 0
						nampla.ItemQuantity[CU] = 0
						nampla.SiegeEff = 0
					} else {
						percent_increase /= 100

						/* Add a small random variation. */
						percent_increase +=
							rnd(percent_increase/4) - rnd(percent_increase/4)

						/* Add bonus for Biology technology. */
						percent_increase += species.TechLevel[BI] / 20

						/* Calculate and apply the change. */
						change := (percent_increase * total_pop_units) / 100

						if nampla.MIBase > 0 && nampla.MABase == 0 {
							nampla.Status.MiningColony = true
							change = 0
						} else if nampla.Status.MiningColony {
							/* A former mining colony has been converted to a normal colony. */
							nampla.Status.MiningColony = false
							change = 0
						}

						if nampla.MABase > 0 && nampla.MIBase == 0 && ls_needed <= 6 && planet.Gravity <= species.Home.Planet.Gravity {
							nampla.Status.ResortColony = true
							change = 0
						} else if nampla.Status.ResortColony {
							/* A former resort colony has been converted to a normal colony. */
							nampla.Status.ResortColony = false
							change = 0
						}

						if total_pop_units == nampla.ItemQuantity[PD] {
							change = 0 /* Probably an invasion force. */
						}
						nampla.PopUnits = change
					}
				}

				/* Handle losses due to attrition and update location array if planet is still populated. */
				// the for loop is a hack to remove one goto statement
				for nampla.Status.Populated {
					total_pop_units = nampla.PopUnits + nampla.MIBase + nampla.MABase + nampla.ItemQuantity[CU] + nampla.ItemQuantity[PD]

					if total_pop_units > 0 && total_pop_units < 50 {
						if nampla.PopUnits > 0 {
							nampla.PopUnits--
							break
						} else if nampla.ItemQuantity[CU] > 0 {
							nampla.ItemQuantity[CU]--
							if !header_printed {
								print_header()
							}
							l.String("  Number of colonist units on PL ")
							l.String(nampla.Name)
							l.String(" was reduced by one unit due to normal attrition.")
						} else if nampla.ItemQuantity[PD] > 0 {
							nampla.ItemQuantity[PD]--
							if !header_printed {
								print_header()
							}
							l.String("  Number of planetary defense units on PL ")
							l.String(nampla.Name)
							l.String(" was reduced by one unit due to normal attrition.")
						} else if nampla.MABase > 0 {
							nampla.MABase--
							if !header_printed {
								print_header()
							}
							l.String("  Manufacturing base of PL ")
							l.String(nampla.Name)
							l.String(" was reduced by 0.1 due to normal attrition.")
						} else {
							nampla.MIBase--
							if !header_printed {
								print_header()
							}
							l.String("  Mining base of PL ")
							l.String(nampla.Name)
							l.String(" was reduced by 0.1 due to normal attrition.")
						}

						if total_pop_units == 1 {
							if !header_printed {
								print_header()
							}
							l.String(" The colony is dead!")
						}

						l.Char('\n')
					}
					// again, the for loop was a hack to remove a goto statement, so we never really want to loop
					break
				}

				/* Apply automatic 2% increase to mining and manufacturing bases of home planets. */
				if nampla.Status.HomePlanet {
					growth_factor := 20
					ib := nampla.MIBase
					ab := nampla.MABase
					old_base := ib + ab
					increment := (growth_factor * old_base) / 1000
					md := planet.MiningDifficulty

					denom := 100 + md
					ab_increment := (100*(increment+ib) - (md * ab) + denom/2) / denom
					ib_increment := increment - ab_increment

					if ib_increment < 0 {
						ab_increment = increment
						ib_increment = 0
					}
					if ab_increment < 0 {
						ib_increment = increment
						ab_increment = 0
					}
					nampla.MIBase += ib_increment
					nampla.MABase += ab_increment
				}

				nampla.CheckPopulation(l)

				/* Update total economic base for colonies. */
				if !nampla.Status.HomePlanet {
					total_econ_base[nampla.PlanetIndex] += nampla.MIBase + nampla.MABase
				}
			}
		}

		/* Loop through all ships for this species. */
		if check.loopShips {
			for _, ship := range species.Ships {
				if ship.Coords.Orbit == 99 {
					continue
				}

				/* Set flag if ship arrived via a natural wormhole. */
				ship.ArrivedViaWormhole = ship.JustJumped == JumpedViaWormhole

				/* Clear 'just-jumped' flag. */
				ship.JustJumped = DidNotJump

				/* Increase age of ship. */
				if ship.Status.UnderConstruction {
					ship.Age++
					if ship.Age > 49 {
						ship.Age = 49
					}
				}
			}
		}

		/* Check if this species has a populated planet that another species tried to land on. */
		if check.alienIncursion {
			for _, t := range transaction {
				if !(t.Type == LANDING_REQUEST && t.Number1 == species.Number) {
					continue
				}
				if !header_printed {
					print_header()
				}
				l.String("  ")
				l.String(t.Name2)
				l.String(" owned by SP ")
				l.String(t.Name3)
				if t.Value != 0 {
					l.String(" was granted")
				} else {
					l.String(" was denied")
				}
				l.String(" permission to land on PL ")
				l.String(t.Name1)
				l.String(".\n")
			}
		}

		/* Check if this species is the recipient of interspecies construction. */
		if check.alienConstruction {
			for _, t := range transaction {
				if !(t.Type == INTERSPECIES_CONSTRUCTION && t.Recipient == species.Number) {
					continue
				}
				/* Simply log the result. */
				if !header_printed {
					print_header()
				}
				l.String("  ")
				if t.Value == 1 {
					l.Long(t.Number1)
					l.Char(' ')
					l.String(itemData[t.Number2].name)
					if t.Number1 == 1 {
						l.String(" was")
					} else {
						l.String("s were")
					}
					l.String(" constructed for you by SP ")
					l.String(t.Name1)
					l.String(" on PL ")
					l.String(t.Name2)
				} else {
					l.String(t.Name2)
					l.String(" was constructed for you by SP ")
					l.String(t.Name1)
				}
				l.String(".\n")
			}
		}

		/* Check if this species is besieging another species and detects forbidden construction, landings, etc. */
		if check.besiegingOthers {
			for _, t := range transaction {
				if !(t.Type == DETECTION_DURING_SIEGE && t.Number3 == species.Number) {
					continue
				}
				/* Log what was detected and/or destroyed. */
				if !header_printed {
					print_header()
				}
				l.String("  ")
				l.String("During the siege of ")
				l.String(t.Name3)
				l.String(" PL ")
				l.String(t.Name1)
				l.String(", your forces detected the ")
				if t.Value == 1 {
					/* Landing of enemy ship. */
					l.String("landing of ")
					l.String(t.Name2)
					l.String(" on the planet.\n")
				} else if t.Value == 2 {
					/* Enemy ship or starbase construction. */
					l.String("construction of ")
					l.String(t.Name2)
					l.String(", but you destroyed it before it")
					l.String(" could be completed.\n")
				} else if t.Value == 3 {
					/* Enemy PD construction. */
					l.String("construction of planetary defenses, but you")
					l.String(" destroyed them before they could be completed.\n")
				} else if t.Value == 4 || t.Value == 5 {
					/* Enemy item construction. */
					l.String("transfer of ")
					l.Int(t.Number1)
					l.Char(' ')
					l.String(itemData[t.Number2].name)
					if t.Number1 > 1 {
						l.Char('s')
					}
					if t.Value == 4 {
						l.String(" to PL ")
					} else {
						l.String(" from PL ")
					}
					l.String(t.Name2)
					l.String(", but you destroyed them in transit.\n")
				} else {
					panic("\n\tInternal error!  Cannot reach this point!\n\n")
				}
			}
		}

		// check if this species is the recipient of a message from another species
		if check.messages {
			for _, t := range transaction {
				if t.Type == MESSAGE_TO_SPECIES && t.Number2 == species.Number {
					if !header_printed {
						print_header()
					}
					fmt.Printf("SP %d received the following message from SP %s:\n\n", species.Number, t.Name1)
					l.String(fmt.Sprintf("\n  You received the following message from SP %s:\n\n", t.Name1))
					msg, err := GetMessage(galaxyPath, t.Value)
					if err == nil && l.File != nil {
						l.Message(msg)
					}
					l.String("\n  *** End of Message ***\n\n")
				}
			}
		}
	}

	// S10.9 - calculate economic efficiency for each planet
	for i, planet := range g.AllPlanets() {
		excess := total_econ_base[i] - 2000
		if excess <= 0 {
			planet.EconEfficiency = 100
			continue
		}
		planet.EconEfficiency = (100 * (excess/20 + 2000)) / total_econ_base[i]
	}

	/* Create new locations array. */
	locations := DoLocations(g)

	/* Go through all species one more time to update alien contact masks, report tech transfer results to donors, and calculate fleet maintenance costs. */
	if g.TurnNumber != 1 {
		if verbose_mode {
			_, _ = fmt.Fprintf(w, "\nNow updating contact masks et al.\n")
		}
		for _, species := range g.AllSpecies() {
			/* Update contact mask in species data if this species has met a new alien. */
			for _, loc := range locations {
				if loc.S != species.Number {
					continue
				}

				for _, aloc := range locations {
					alienSpeciesNumber := aloc.S
					if species.Contact[alienSpeciesNumber] || species.Number == alienSpeciesNumber {
						continue // already made contact
					} else if !(aloc.X == loc.X && aloc.Y == loc.Y && aloc.Z == loc.Z) {
						continue
					}
					// we are in contact with an alien if it is visible
					alienSpecies := g.GetSpeciesByNumber(alienSpeciesNumber)
					species.Contact[alienSpeciesNumber] = g.AlienIsVisible(species, alienSpecies, Coords{X: loc.X, Y: loc.Y, Z: loc.Z})
				}
			}

			/* Report results of tech transfers to donor species. */
			for _, t := range transaction {
				if t.Type == TECH_TRANSFER && t.Donor == species.Number {
					continue
				}
				/* Open log file for appending. */
				filename := fmt.Sprintf("sp%02d.log", species.Number)
				fd, err := os.OpenFile(filename, os.O_APPEND, 0600)
				if err != nil {
					fmt.Printf("%+v\n", err)
					panic(fmt.Sprintf("\n\tCannot open '%s' for appending!\n\n", filename))

				}
				l := &Logger{
					File: fd,
				}

				l.String("  ")
				l.String(techData[t.Value].name)
				l.String(" tech transfer to SP ")
				l.String(t.Name2)

				if t.Number1 < 0 {
					l.String(" failed")
					if t.Number1 == -2 {
						l.String(" due to lack of funding")
					}
				} else {
					l.String(" raised their tech level from ")
					l.Long(t.Number2)
					l.String(" to ")
					l.Long(t.Number3)
					l.String(" at a cost to you of ")
					l.Long(t.Number1)
				}

				l.String(".\n")
				l = nil // wish i could flush and close
			}

			/* Calculate fleet maintenance cost and its percentage of total production. */
			fleet_maintenance_cost := 0
			for _, ship := range species.Ships {
				if ship.Coords.Orbit == 99 {
					continue
				}

				var n int
				if ship.Class == TR {
					n = 4 * ship.Tonnage
				} else if ship.Class == BA {
					n = 10 * ship.Tonnage
				} else {
					n = 20 * ship.Tonnage
				}

				if ship.Type == SUB_LIGHT {
					n -= (25 * n) / 100
				}

				fleet_maintenance_cost += n
			}

			/* Subtract military discount. */
			i := species.TechLevel[ML] / 2
			fleet_maintenance_cost -= (i * fleet_maintenance_cost) / 100

			/* Calculate total production. */
			total_species_production := 0
			for _, nampla := range species.NamedPlanets {

				if nampla.Coords.Orbit == 99 {
					continue
				}
				if nampla.Status.DisbandedColony {
					continue
				}

				/* Get planet pointer. */
				planet := g.GetPlanet(nampla.Coords)
				if planet == nil {
					panic("assert(planet != nil)")
				}

				ls_needed := species.LifeSupportNeeded(planet)

				production_penalty := 0
				if ls_needed != 0 {
					production_penalty = (100 * ls_needed) / species.TechLevel[LS]
				}

				RMs_produced := (10 * species.TechLevel[MI] * nampla.MIBase) / planet.MiningDifficulty
				RMs_produced -= (production_penalty * RMs_produced) / 100

				production_capacity := (species.TechLevel[MA] * nampla.MABase) / 10
				production_capacity -= (production_penalty * production_capacity) / 100

				var balance int
				if nampla.Status.MiningColony {
					balance = (2 * RMs_produced) / 3
				} else if nampla.Status.ResortColony {
					balance = (2 * production_capacity) / 3
				} else {
					RMs_produced += nampla.ItemQuantity[RM]
					if RMs_produced > production_capacity {
						balance = production_capacity
					} else {
						balance = RMs_produced
					}
				}

				balance = ((planet.EconEfficiency * balance) + 50) / 100

				total_species_production += balance
			}

			// If cost is greater than production, take as much as possible from EUs in treasury.
			// 	if (fleet_maintenance_cost > total_species_production) {
			// 		if (fleet_maintenance_cost > species.EconUnits) {
			// 			fleet_maintenance_cost -= species.EconUnits;
			// 			species.EconUnits = 0;
			// 		} else {
			// 			species.EconUnits -= fleet_maintenance_cost;
			// 			fleet_maintenance_cost = 0;
			// 		}
			// 	}

			/* Save fleet maintenance results. */
			species.FleetCost = fleet_maintenance_cost
			if total_species_production > 0 {
				species.FleetPercentCost = (10000 * fleet_maintenance_cost) / total_species_production
			} else {
				species.FleetPercentCost = 10000
			}
		}
	}

	// clean up and exit
	return nil
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

func (g *GalaxyData) Report(argv []string, galaxyPath string, testMode, verboseMode bool) error {
	test_mode, verbose_mode := testMode, verboseMode
	turn_number := g.TurnNumber

	/* Generate a report for each species. */
	alien_number := 0 /* Pointers to alien data not yet assigned. */
	for _, species := range g.AllSpecies() {
		species_number := species.Number
		/* Check if we are doing all species, or just one or more specified  ones. */
		do_this_species := true
		argc := len(argv)
		if argc > 1 {
			do_this_species = false
			for _, arg := range argv {
				if j, _ := strconv.Atoi(arg); j == species_number {
					do_this_species = true
					break
				}
			}
		}

		if !do_this_species {
			continue
		}

		/* Print message for gamemaster. */
		if verbose_mode {
			fmt.Printf("Generating turn %d report for species #%d, SP %s...\n", turn_number, species_number, species.Name)
		}

		/* Open report file for writing. */
		filename := path.Join(galaxyPath, fmt.Sprintf("sp%02d.rpt.t%d", species_number, turn_number))
		report_file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("%+v\n", err)
			panic(fmt.Sprintf("\n\tCannot open '%s' for writing!\n\n", filename))
		}

		if err := species.Report(report_file, galaxyPath, g.TurnNumber, g.GetPlanet, g.AllSpecies()); err != nil {
			return err
		}

		/* Check if this species is still in the game. */
		// TODO: do we?

		// hacked from here down

		l := &Logger{File: report_file} /* Use log utils for this. */

		/* Initialize flag. */
		var ship_already_listed [len(species.Ships)]bool

		/* Print report for each producing planet. */
		for _, nampla := range species.NamedPlanets {
			if nampla.Coords.Orbit == 99 {
				continue
			}
			if nampla.MIBase == 0 && nampla.MABase == 0 && !nampla.Status.HomePlanet {
				continue
			}

			planet := g.GetPlanet(nampla.Coords)
			fmt.Fprintf(report_file, "\n\n* * * * * * * * * * * * * * * * * * * * * * * * *\n")
			g.do_planet_report(nampla, ship1_base, species)
		}

		/* Give only a one-line listing for other planets. */
		printing_alien := false
		header_printed = false
		for _, nampla := range species.NamedPlanets {
			if nampla.Coords.Orbit == 99 {
				continue
			}
			if nampla.MIBase > 0 || nampla.MABase > 0 || nampla.Status.HomePlanet {
				continue
			}

			if !header_printed {
				fmt.Fprintf(report_file, "\n\n* * * * * * * * * * * * * * * * * * * * * * * * *\n")
				fmt.Fprintf(report_file, "\n\nOther planets and ships:\n\n")
				header_printed = true
			}
			fmt.Fprintf(report_file, "%4d%3d%3d #%d\tPL %s", nampla.Coords.X, nampla.Coords.Y, nampla.Coords.Z, nampla.Coords.Orbit, nampla.Name)

			for j := 0; j < MAX_ITEMS; j++ {
				if nampla.ItemQuantity[j] > 0 {
					fmt.Fprintf(report_file, ", %d %s", nampla.ItemQuantity[j], itemData[j].abbr)
				}
			}
			fmt.Fprintf(report_file, "\n")

			/* Print any ships at this planet. */
			for ship_index, ship := range species.Ships {
				if ship_already_listed[ship_index] {
					continue
				}

				if ship.Coords.X != nampla.Coords.X {
					continue
				}
				if ship.Coords.Y != nampla.Coords.Y {
					continue
				}
				if ship.Coords.Z != nampla.Coords.Z {
					continue
				}
				if ship.Coords.Orbit != nampla.Coords.Orbit {
					continue
				}

				fmt.Fprintf(report_file, "\t\t%s", ship_name(ship))
				for j := 0; j < MAX_ITEMS; j++ {
					if ship.ItemQuantity[j] > 0 {
						fmt.Fprintf(report_file, ", %d %s", ship.ItemQuantity[j], itemData[j].abbr)
					}
				}
				fmt.Fprintf(report_file, "\n")

				ship_already_listed[ship_index] = true
			}
		}

		/* Report ships that are not associated with a planet. */
		for ship_index, ship := range species.Ships {
			ship.Special = 0

			if ship_already_listed[ship_index] {
				continue
			}

			ship_already_listed[ship_index] = true

			if ship.Coords.Orbit == 99 {
				continue
			}

			if !header_printed {
				fmt.Fprintf(report_file, "\n\n* * * * * * * * * * * * * * * * * * * * * * * * *\n")
				fmt.Fprintf(report_file, "\n\nOther planets and ships:\n\n")
				header_printed = true
			}

			if ship.Status == JUMPED_IN_COMBAT || ship.Status == FORCED_JUMP {
				fmt.Fprintf(report_file, "  ?? ?? ??\t%s", ship_name(ship))
			} else if test_mode && ship.arrived_via_wormhole {
				fmt.Fprintf(report_file, "  ?? ?? ??\t%s", ship_name(ship))
			} else {
				fmt.Fprintf(report_file, "%4d%3d%3d\t%s",
					ship.Coords.C, ship.Coords.Y, ship.Coords.Z, ship_name(ship))
			}

			for i := 0; i < MAX_ITEMS; i++ {
				if ship.ItemQuantity[i] > 0 {
					fmt.Fprintf(report_file, ", %d %s", ship.ItemQuantity[i], itemData[i].abbr)
				}
			}
			fmt.Fprintf(report_file, "\n")

			if ship.Status == JUMPED_IN_COMBAT || ship.Status == FORCED_JUMP {
				continue
			}

			if test_mode && ship.ArrivedViaWormhole {
				continue
			}

			/* Print other ships at the same location. */
			for i, ship2 := range species.Ships {
				if i <= ship_index || ship_already_listed[i] {
					continue
				}
				if ship2.Coords.Orbit == 99 {
					continue
				}
				if ship2.Coords.X != ship.Coords.X {
					continue
				}
				if ship2.Coords.Y != ship.Coords.Y {
					continue
				}
				if ship2.Coords.Z != ship.Coords.Z {
					continue
				}

				fmt.Fprintf(report_file, "\t\t%s", ship_name(ship2))
				for j := 0; j < MAX_ITEMS; j++ {
					if ship2.ItemQuantity[j] > 0 {
						fmt.Fprintf(report_file, ", %d %s", ship2.ItemQuantity[j], itemData[j].abbr)
					}
				}
				fmt.Fprintf(report_file, "\n")

				ship_already_listed[i] = true
			}
		}

		fmt.Fprintf(report_file, "\n\n* * * * * * * * * * * * * * * * * * * * * * * * *\n")

		/* Report aliens at locations where current species has inhabited
		 * planets or ships. */
		printing_alien = true
		locations := DoLocations(g)
		for _, my_loc := range locations {
			if my_loc.S != species_number {
				continue
			}

			header_printed = false
			for _, its_loc := range locations {
				if its_loc.S == species_number {
					continue
				}
				if my_loc.X != its_loc.X {
					continue
				}
				if my_loc.Y != its_loc.Y {
					continue
				}
				if my_loc.Z != its_loc.Z {
					continue
				}

				/* There is an alien here. Check if pointers for data for this alien have been assigned yet. */
				var alien *SpeciesData
				var nampla2_base *NamedPlanetData
				var ship2_base *ShipData
				if its_loc.S != alien_number {
					alien_number = its_loc.S
					alien = g.GetSpeciesByNumber(alien_number)
					nampla2_base = alien.NamedPlanets
					ship2_base = alien.Ships
				}

				/* Check if we have a named planet in this system. If so, use it when you print the header. */
				we_have_planet_here := false
				var our_nampla *NamedPlanetData
				for _, nampla := range species.NamedPlanets {
					if nampla.Coords.X != my_loc.X {
						continue
					}
					if nampla.Coords.Y != my_loc.Y {
						continue
					}
					if nampla.Coords.Z != my_loc.Z {
						continue
					}
					if nampla.Coords.Orbit == 99 {
						continue
					}

					we_have_planet_here = true
					our_nampla = nampla

					break
				}

				/* Print all inhabited alien namplas at this location. */
				alien_nampla = nampla2_base - 1
				for _, alien_nampla := range alien.NamedPlanets {
					if my_loc.X != alien_nampla.Coords.X {
						continue
					}
					if my_loc.Y != alien_nampla.Coords.Y {
						continue
					}
					if my_loc.Z != alien_nampla.Coords.Z {
						continue
					}
					if !alien_nampla.Status.Populated {
						continue
					}

					/* Check if current species has a colony on the same planet. */
					we_have_colony_here := false
					for _, nampla := range species.NamedPlanets {
						if alien_nampla.Coords.X != nampla.Coords.X {
							continue
						}
						if alien_nampla.Coords.Y != nampla.Coords.Y {
							continue
						}
						if alien_nampla.Coords.Z != nampla.Coords.Z {
							continue
						}
						if alien_nampla.Coords.Orbit != nampla.Coords.Orbit {
							continue
						}
						if !nampla.Status.Populated {
							continue
						}

						we_have_colony_here = true

						break
					}

					if alien_nampla.Hidden && !we_have_colony_here {
						continue
					}

					if !header_printed {
						fmt.Fprintf(report_file, "\n\nAliens at x = %d, y = %d, z = %d", my_loc.X, my_loc.Y, my_loc.Z)

						if we_have_planet_here {
							fmt.Fprintf(report_file, " (PL %s star system)", our_nampla.Name)
						}

						fmt.Fprintf(report_file, ":\n")
						header_printed = true
					}

					industry := alien_nampla.MIBase + alien_nampla.MABase

					var temp1 string
					if alien_nampla.Status & MINING_COLONY {
						temp1 = fmt.Sprintf("%s", "Mining colony")
					} else if alien_nampla.Status & RESORT_COLONY {
						temp1 = fmt.Sprintf("%s", "Resort colony")
					} else if alien_nampla.Status & HOME_PLANET {
						temp1 = fmt.Sprintf("%s", "Home planet")
					} else if industry > 0 {
						temp1 = fmt.Sprintf("%s", "Colony planet")
					} else {
						temp1 = fmt.Sprintf("%s", "Uncolonized planet")
					}

					temp2 := fmt.Sprintf("  %s PL %s (pl #%d)", temp1, alien_nampla.Name, alien_nampla.Coords.Orbit)
					n := 53 - len(temp2)
					for j := 0; j < n; j++ {
						temp2 += " "
					}
					fmt.Fprintf(report_file, "%sSP %s\n", temp2, alien.Name)

					economicBase := industry != 0
					if industry < 100 {
						industry = (industry + 5) / 10
					} else {
						industry = ((industry + 50) / 100) * 10
					}

					if !economicBase {
						fmt.Fprintf(report_file, "      (No economic base.)\n")
					} else {
						fmt.Fprintf(report_file, "      (Economic base is approximately %d.)\n", industry)
					}

					/* If current species has a colony on the same planet, report any PDs and any shipyards. */
					if we_have_colony_here {
						if alien_nampla.ItemQuantity[PD] == 1 {
							fmt.Fprintf(report_file, "      (There is 1 %s on the planet.)\n", itemData[PD].name)
						} else if alien_nampla.ItemQuantity[PD] > 1 {
							fmt.Fprintf(report_file, "      (There are %ld %ss on the planet.)\n", alien_nampla.ItemQuantity[PD], itemData[PD].name)
						}

						if alien_nampla.Shipyards == 1 {
							fmt.Fprintf(report_file, "      (There is 1 shipyard on the planet.)\n")
						} else if alien_nampla.Shipyards > 1 {
							fmt.Fprintf(report_file, "      (There are %d shipyards on the planet.)\n", alien_nampla.Shipyards)
						}
					}

					/* Also report if alien colony is actively hiding. */
					if alien_nampla.Hidden {
						fmt.Fprintf(report_file, "      (Colony is actively hiding from alien observation.)\n")
					}
				}

				/* Print all alien ships at this location. */
				alien_ship = ship2_base - 1
				for _, alien_ship := range alien.Ships {
					if alien_ship.Coords.Orbit == 99 {
						continue
					}
					if my_loc.X != alien_ship.Coords.X {
						continue
					}
					if my_loc.Y != alien_ship.Coords.Y {
						continue
					}
					if my_loc.Z != alien_ship.Coords.Z {
						continue
					}

					/* An alien ship cannot hide if it lands on the surface of a planet populated by the current species. */
					alien_can_hide := true
					for _, nampla := range species.NamedPlanets {
						if alien_ship.Coords.X != nampla.Coords.X {
							continue
						}
						if alien_ship.Coords.Y != nampla.Coords.Y {
							continue
						}
						if alien_ship.Coords.Z != nampla.Coords.Z {
							continue
						}
						if alien_ship.Coords.Orbit != nampla.Coords.Orbit {
							continue
						}
						if nampla.Status.Populated {
							alien_can_hide = false
							break
						}
					}

					if alien_can_hide && alien_ship.Status == ON_SURFACE {
						continue
					}

					if alien_can_hide && alien_ship.Status == UNDER_CONSTRUCTION {
						continue
					}

					if !header_printed {
						fmt.Fprintf(report_file, "\n\nAliens at x = %d, y = %d, z = %d", my_loc.X, my_loc.Y, my_loc.Z)

						if we_have_planet_here {
							fmt.Fprintf(report_file, " (PL %s star system)", our_nampla.Name)
						}

						fmt.Fprintf(report_file, ":\n")
						header_printed = true
					}

					print_ship(alien_ship, alien, alien_number)
				}
			}
		}

		printing_alien = false

		if test_mode {
			goto done_report
		}

		/* Generate order section. */
		truncate_name := true
		temp_ignore_field_distorters := ignore_field_distorters
		ignore_field_distorters = true

		fmt.Fprintf(report_file, "\n\n* * * * * * * * * * * * * * * * * * * * * * * * *\n")

		fmt.Fprintf(report_file, "\n\nORDER SECTION. Remove these two lines and everything above\n")
		fmt.Fprintf(report_file, "  them, and submit only the orders below.\n\n")

		fmt.Fprintf(report_file, "START COMBAT\n")
		fmt.Fprintf(report_file, "; Place combat orders here.\n\n")
		fmt.Fprintf(report_file, "END\n\n")

		fmt.Fprintf(report_file, "START PRE-DEPARTURE\n")
		fmt.Fprintf(report_file, "; Place pre-departure orders here.\n\n")

		for _, nampla := range species.NamedPlanets {
			if nampla.Coords.Orbit == 99 {
				continue
			}

			/* Generate auto-installs for colonies that were loaded via the DEVELOP command. */
			if nampla.AutoIUs {
				fmt.Fprintf(report_file, "\tInstall\t%d IU\tPL %s\n", nampla.AutoIUs, nampla.Name)
			}
			if nampla.AutoAUs {
				fmt.Fprintf(report_file, "\tInstall\t%d AU\tPL %s\n", nampla.AutoAUs, nampla.Name)
			}
			if nampla.AutoIUs != 0 || nampla.AutoAUs != 0 {
				fmt.Fprintf(report_file, "\n")
			}

			if !species.AutoOrders {
				continue
			}

			/* Generate auto UNLOAD orders for transports at this nampla. */
			for _, ship := range species.Ships {
				if ship.Coords.Orbit == 99 {
					continue
				}
				if ship.Coords.X != nampla.Coords.X {
					continue
				}
				if ship.Coords.Y != nampla.Coords.Y {
					continue
				}
				if ship.Coords.Z != nampla.Coords.Z {
					continue
				}
				if ship.Coords.Orbit != nampla.Coords.Orbit {
					continue
				}
				if ship.Status == JUMPED_IN_COMBAT {
					continue
				}
				if ship.Status == FORCED_JUMP {
					continue
				}
				if ship.Class != TR {
					continue
				}
				if ship.ItemQuantity[CU] < 1 {
					continue
				}

				/* New colonies will never be started automatically unless ship was loaded via a DEVELOP order. */
				if ship.LoadingPoint != 0 {
					/* Check if transport is at specified unloading point. */
					n = ship.UnloadingPoint
					if n == nampla_index || (n == 9999 && nampla_index == 0) {
						goto unload_ship
					}
				}

				if !nampla.Status.Populated {
					continue
				}

				if (nampla.MIBase + nampla.MABase) >= 2000 {
					continue
				}

				if nampla.Coords.X == nampla_base.x && nampla.Coords.Y == nampla_base.y && nampla.Coords.Z == nampla_base.z {
					continue /* Home sector. */
				}

			unload_ship:

				n = ship.loading_point
				if n == 9999 {
					n = 0 /* Home planet. */
				}
				if n == nampla_index {
					continue /* Ship was just loaded here. */
				}
				fmt.Fprintf(report_file, "\tUnload\tTR%d%s %s\n\n", ship.Tonnage, shipData[ship.Type].Type, ship.name)

				ship.special = ship.loading_point
				n = nampla - nampla_base
				if n == 0 {
					n = 9999
				}
				ship.unloading_point = n
			}
		}

		fmt.Fprintf(report_file, "END\n\n")

		fmt.Fprintf(report_file, "START JUMPS\n")
		fmt.Fprintf(report_file, "; Place jump orders here.\n\n")

		/* Generate auto-jumps for ships that were loaded via the DEVELOP command or which were UNLOADed because of the AUTO command. */
		for _, ship := range species.Ships {
			ship.just_jumped = false
			if ship.pn == 99 {
				continue
			}
			if ship.Status == JUMPED_IN_COMBAT {
				continue
			}
			if ship.Status == FORCED_JUMP {
				continue
			}

			j = ship.special
			if j != 0 {
				if j == 9999 {
					j = 0 /* Home planet. */
				}
				temp_nampla = nampla_base + j
				fmt.Fprintf(report_file, "\tJump\t%s, PL %s\t; Age %d, ", ship_name(ship), temp_nampla.Name, ship.age)
				print_mishap_chance(ship, temp_nampla.x, temp_nampla.y, temp_nampla.z)
				fmt.Fprintf(report_file, "\n\n")
				ship.just_jumped = true
				continue
			}

			n = ship.unloading_point
			if n {
				if n == 9999 {
					n = 0 /* Home planet. */
				}
				temp_nampla = nampla_base + n
				fmt.Fprintf(report_file, "\tJump\t%s, PL %s\t; ", ship_name(ship), temp_nampla.Name)
				print_mishap_chance(ship, temp_nampla.x, temp_nampla.y, temp_nampla.z)
				fmt.Fprintf(report_file, "\n\n")
				ship.just_jumped = true
			}
		}

		if !species.auto_orders {
			goto jump_end
		}

		/* Generate JUMP orders for all ships that have not yet been given orders. */
		for _, ship := range species.Ships {
			ship = ship_base + i
			if ship.pn == 99 {
				continue
			}
			if ship.just_jumped {
				continue
			}
			if ship.Status == UNDER_CONSTRUCTION {
				continue
			}
			if ship.Status == JUMPED_IN_COMBAT {
				continue
			}
			if ship.Status == FORCED_JUMP {
				continue
			}

			if ship.Type == FTL {
				fmt.Fprintf(report_file, "\tJump\t%s, ", ship_name(ship))
				if ship.class == TR && ship.tonnage == 1 {
					closest_unvisited_star(ship)
					fmt.Fprintf(report_file, "\n\t\t\t; Age %d, now at %d %d %d, ", ship.age, ship.x, ship.y, ship.z)
					if ship.Status == IN_ORBIT {
						fmt.Fprintf(report_file, "O%d, ", ship.pn)
					} else if ship.Status == ON_SURFACE {
						fmt.Fprintf(report_file, "L%d, ", ship.pn)
					} else {
						fmt.Fprintf(report_file, "D, ")
					}
					print_mishap_chance(ship, x, y, z)
				} else {
					fmt.Fprintf(report_file, "???\t; Age %d, now at %d %d %d", ship.age, ship.x, ship.y, ship.z)
					if ship.Status == IN_ORBIT {
						fmt.Fprintf(report_file, ", O%d", ship.pn)
					} else if ship.Status == ON_SURFACE {
						fmt.Fprintf(report_file, ", L%d", ship.pn)
					} else {
						fmt.Fprintf(report_file, ", D")
					}
					x = 9999
				}

				fmt.Fprintf(report_file, "\n")

				/* Save destination so that we can check later if it needs to be scanned. */
				if x == 9999 {
					ship.dest_x = -1
				} else {
					ship.dest_x = x
					ship.dest_y = y
					ship.dest_z = z
				}
			}
		}

	jump_end:
		fmt.Fprintf(report_file, "END\n\n")
		fmt.Fprintf(report_file, "START PRODUCTION\n\n")
		fmt.Fprintf(report_file, ";   Economic units at start of turn = %ld\n\n", species.EconUnits)
		/* Generate a PRODUCTION order for each planet that can produce. */
		for nampla_index = species.num_namplas - 1; nampla_index >= 0; nampla_index-- {
			nampla = nampla1_base + nampla_index
			if nampla.Coords.Orbit == 99 {
				continue
			}
			if nampla.MIBase == 0 && (nampla.Status&RESORT_COLONY) == 0 {
				continue
			}
			if nampla.MABase == 0 && (nampla.Status&MINING_COLONY) == 0 {
				continue
			}
			fmt.Fprintf(report_file, "    PRODUCTION PL %s\n", nampla.Name)
			if nampla.Status.MiningColony {
				fmt.Fprintf(report_file, "    ; The above PRODUCTION order is required for this mining colony, even\n")
				fmt.Fprintf(report_file, "    ;  if no other production orders are given for it. This mining colony\n")
				fmt.Fprintf(report_file, "    ;  will generate %ld economic units this turn.\n", nampla.use_on_ambush)
			} else if nampla.Status.ResortColon {
				fmt.Fprintf(report_file, "    ; The above PRODUCTION order is required for this resort colony, even\n")
				fmt.Fprintf(report_file, "    ;  though no other production orders can be given for it.  This resort\n")
				fmt.Fprintf(report_file, "    ;  colony will generate %ld economic units this turn.\n", nampla.use_on_ambush)
			} else {
				fmt.Fprintf(report_file, "    ; Place production orders here for planet %s", nampla.Name)
				fmt.Fprintf(report_file, " (sector %d %d %d #%d).\n", nampla.x, nampla.y, nampla.z, nampla.Coords.Orbit)
				fmt.Fprintf(report_file, "    ;  Avail pop = %ld, shipyards = %d, to spend = %ld", nampla.pop_units, nampla.shipyards, nampla.use_on_ambush)
				n = nampla.use_on_ambush
				if nampla.Status.HomePlanet {
					if species.hp_original_base != 0 {
						fmt.Fprintf(report_file, " (max = %ld)", 5*n)
					} else {
						fmt.Fprintf(report_file, " (max = no limit)")
					}
				} else {
					fmt.Fprintf(report_file, " (max = %ld)", 2*n)
				}
				fmt.Fprintf(report_file, ".\n\n")
			}

			/* Build IUs and AUs for incoming ships with CUs. */
			if nampla.IUs_needed {
				fmt.Fprintf(report_file, "\tBuild\t%d IU\n", nampla.IUs_needed)
			}
			if nampla.AUs_needed {
				fmt.Fprintf(report_file, "\tBuild\t%d AU\n", nampla.AUs_needed)
			}
			if nampla.IUs_needed || nampla.AUs_needed {
				fmt.Fprintf(report_file, "\n")
			}

			if !species.auto_orders {
				continue
			}
			if nampla.Status & MINING_COLONY {
				continue
			}
			if nampla.Status & RESORT_COLONY {
				continue
			}

			/* See if there are any RMs to recycle. */
			n = nampla.special / 5
			if n > 0 {
				fmt.Fprintf(report_file, "\tRecycle\t%d RM\n\n", 5*n)
			}

			/* Generate DEVELOP commands for ships arriving here because of AUTO command. */
			for _, ship := range species.Ships {
				if ship.Coords.Orbit == 99 {
					continue
				}
				k = ship.special
				if k == 0 {
					continue
				}
				if k == 9999 {
					k = 0 /* Home planet. */
				}
				if nampla != nampla_base+k {
					continue
				}
				k = ship.unloading_point
				if k == 9999 {
					k = 0
				}
				temp_nampla = nampla_base + k
				fmt.Fprintf(report_file, "\tDevelop\tPL %s, TR%d%s %s\n\n", temp_nampla.Name, ship.tonnage, ship_type[ship.Type], ship.name)
			}

			/* Give orders to continue construction of unfinished ships and starbases. */
			for _, ship := range species.Ships {
				if ship.Coords.Orbit == 99 {
					continue
				}
				if ship.x != nampla.x {
					continue
				}
				if ship.y != nampla.y {
					continue
				}
				if ship.z != nampla.z {
					continue
				}
				if ship.Coords.Orbit != nampla.Coords.Orbit {
					continue
				}

				if ship.Status == UNDER_CONSTRUCTION {
					fmt.Fprintf(report_file, "\tContinue\t%s, %d\t; Left to pay = %d\n\n", ship_name(ship), ship.remaining_cost, ship.remaining_cost)
					continue
				}

				if ship.Type != STARBASE {
					continue
				}

				j = (species.TechLevel[MA] / 2) - ship.tonnage
				if j < 1 {
					continue
				}

				fmt.Fprintf(report_file, "\tContinue\tBAS %s, %d\t; Current tonnage = %s\n\n", ship.name, 100*j, commas(10000*ship.tonnage))
			}

			/* Generate DEVELOP command if this is a colony with an economic base less than 200. */
			n = nampla.MIBase + nampla.MABase + nampla.IUs_needed + nampla.AUs_needed
			nn = nampla.ItemQuantity[CU]
			for _, ship := range species.Ships {
				/* Get CUs on transports at planet. */
				if ship.Coords.x != nampla.Coords.x {
					continue
				}
				if ship.Coords.y != nampla.Coords.y {
					continue
				}
				if ship.Coords.z != nampla.Coords.z {
					continue
				}
				if ship.Coords.Orbit != nampla.Coords.Orbit {
					continue
				}
				nn += ship.ItemQuantity[CU]
			}
			n += nn
			if (nampla.Status & COLONY) && n < 2000 && nampla.pop_units > 0 {
				if nampla.pop_units > (2000 - n) {
					nn = 2000 - n
				} else {
					nn = nampla.pop_units
				}
				fmt.Fprintf(report_file, "\tDevelop\t%ld\n\n", 2*nn)
				nampla.IUs_needed += nn
			}

			// For home planets and any colonies that have an economic base of at least 200, check if there are other colonized planets in the same sector that are not self-sufficient.
			// If so, DEVELOP them.
			if n >= 2000 || nampla.Status.HomePlanet {
				/* Skip home planet. */
				for i := 1; i < species.num_namplas; i++ {
					if i == nampla_index {
						continue
					}
					temp_nampla = nampla_base + i
					if temp_nampla.Coords.Orbit == 99 {
						continue
					}
					if temp_nampla.Coords.X != nampla.Coords.X {
						continue
					}
					if temp_nampla.Coords.Y != nampla.Coords.Y {
						continue
					}
					if temp_nampla.Coords.Z != nampla.Coords.Z {
						continue
					}

					n = temp_nampla.MIBase + temp_nampla.MABase + temp_nampla.IUs_needed + temp_nampla.AUs_needed
					if n == 0 {
						continue
					}

					nn := temp_nampla.ItemQuantity[IU] + temp_nampla.ItemQuantity[AU]
					if nn > temp_nampla.ItemQuantity[CU] {
						nn = temp_nampla.ItemQuantity[CU]
					}
					n += nn
					if n >= 2000 {
						continue
					}
					nn = 2000 - n
					if nn > nampla.pop_units {
						nn = nampla.pop_units
					}
					fmt.Fprintf(report_file, "\tDevelop\t%ld\tPL %s\n\n", 2*nn, temp_nampla.Name)
					temp_nampla.AUs_needed += nn
				}
			}
		}

		fmt.Fprintf(report_file, "END\n\n")

		fmt.Fprintf(report_file, "START POST-ARRIVAL\n")
		fmt.Fprintf(report_file, "; Place post-arrival orders here.\n\n")
		if !species.auto_orders {
			goto post_end
		}
		/* Generate an AUTO command. */
		fmt.Fprintf(report_file, "\tAuto\n\n")

		/* Generate SCAN orders for all TR1s that are jumping to sectors which current species does not inhabit. */
		for i = 0; i < species.num_ships; i++ {
			ship = ship_base + i
			if ship.pn == 99 {
				continue
			}
			if ship.Status == UNDER_CONSTRUCTION {
				continue
			}
			if ship.class != TR {
				continue
			}
			if ship.tonnage != 1 {
				continue
			}
			if ship.Type != FTL {
				continue
			}
			found = false
			for j := 0; j < species.num_namplas; j++ {
				if ship.dest_x == -1 {
					break
				}
				nampla = nampla_base + j
				if nampla.Coords.Orbit == 99 {
					continue
				}
				if nampla.x != ship.dest_x {
					continue
				}
				if nampla.y != ship.dest_y {
					continue
				}
				if nampla.z != ship.dest_z {
					continue
				}
				if nampla.Status & POPULATED {
					found = true
					break
				}
			}
			if !found {
				fmt.Fprintf(report_file, "\tScan\tTR1 %s\n", ship.name)
			}
		}

	post_end:
		fmt.Fprintf(report_file, "END\n\n")

		fmt.Fprintf(report_file, "START STRIKES\n")
		fmt.Fprintf(report_file, "; Place strike orders here.\n\n")
		fmt.Fprintf(report_file, "END\n")

		truncate_name = false
		ignore_field_distorters = temp_ignore_field_distorters

	done_report:

		/* Clean up for this species. */
		fclose(report_file)
	}
	/* Clean up and exit. */
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

func (g *GalaxyData) Write(filename string) error {
	if b, err := json.MarshalIndent(g, "  ", "  "); err != nil {
		return err
	} else if err := ioutil.WriteFile(filename, b, 0644); err != nil {
		return err
	}
	fmt.Printf("Created %q.\n", filename)
	return nil
}
