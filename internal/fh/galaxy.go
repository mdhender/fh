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
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
)

type GalaxyData struct {
	ID                string
	Name              string
	Secret            string
	Players           map[string]*Player
	Center            Coords
	allSpecies        []*SpeciesData
	Species           map[string]*SpeciesData
	DNumSpecies       int
	NumSpecies        int
	Radius            int
	NumberOfWormHoles int
	NumberOfPlanets   int
	allSystems        []*StarData
	Systems           map[int]*StarData
	// SystemTemplates is an attempt to balance out the player's home systems.
	// Ideally, every species will have a nice earth-like planet to start from.
	SystemTemplates [10][]*PlanetData
	Translate       struct {
		EmailToID        map[string]string
		IndexToSpeciesID []string
		SpeciesNameToID  map[string]string
		XYZToSystem      map[string]*StarData
	}
	allPlanets []*PlanetData
	l          *Logger
}

type Player struct {
	ID           string `json:"id"`
	EmailAddress string `json:"email"`
	Species      string `json:"species"`
}

func GenerateGalaxy(l *Logger, setupData *SetupData, galaxyPath string, players []*PlayerData) (*GalaxyData, error) {
	// derive galaxy setup data from number of players
	var desiredNumberOfSystems int
	if setupData.Galaxy.Overrides.UseOverrides {
		typicalNumberOfStars := EstimateNumberOfSystems(len(players), setupData.Galaxy.LowDensity)
		desiredNumberOfSystems = setupData.Galaxy.Overrides.NumberOfStars
		l.Printf("For %d species, overriding normal %d stars to %d stars.\n", typicalNumberOfStars, desiredNumberOfSystems)
	} else {
		desiredNumberOfSystems = EstimateNumberOfSystems(len(players), setupData.Galaxy.LowDensity)
		l.Printf("For %d species, there should be about %d stars.\n", len(players), desiredNumberOfSystems)
	}
	if !(MIN_STARS <= desiredNumberOfSystems) {
		desiredNumberOfSystems = MIN_STARS
		l.Printf("Warning: forcing number of stars to minimum of %d stars.\n", desiredNumberOfSystems)
	} else if !(desiredNumberOfSystems <= MAX_STARS) {
		desiredNumberOfSystems = MAX_STARS
		l.Printf("Warning: forcing number of stars to maximum of %d stars.\n", desiredNumberOfSystems)
	}

	// ensure that we have enough systems for the players to each claim 5 systems
	if !(len(players) < desiredNumberOfSystems) {
		panic(fmt.Sprintf("assert(len(players) < desiredNumberOfSystems (%d < %d))", len(players), desiredNumberOfSystems))
	} else if !(5*len(players) <= desiredNumberOfSystems) {
		panic(fmt.Sprintf("assert(6 * len(players) < desiredNumberOfSystems (%d <= %d))", 5*len(players), desiredNumberOfSystems))
	}

	g := &GalaxyData{
		ID:     setupData.Galaxy.Name,
		Name:   setupData.Galaxy.Name,
		Secret: "your-private-key-belongs-here",
		Radius: setupData.Galaxy.MinimumDistance, // setup data influences the minimum radius
		l:      l,
	}
	defer l.Close()

	// initialize from some player data
	g.Players = make(map[string]*Player)
	g.Translate.EmailToID = make(map[string]string)
	for _, player := range players {
		id := player.Email
		g.Players[id] = &Player{
			ID:           id,
			EmailAddress: player.Email,
			Species:      player.SpeciesName,
		}
		g.Translate.EmailToID[id] = player.Email
	}

	// create the home system templates
	for i := 3; i < 10; i++ {
		g.SystemTemplates[i] = NewSystemTemplates(l, i, 50_000)
		if g.SystemTemplates[i] == nil {
			return nil, fmt.Errorf("unable to generate template for system with %d planets", i)
		}
	}

	g.Species = make(map[string]*SpeciesData)
	g.Translate.SpeciesNameToID = make(map[string]string)
	g.DNumSpecies = len(players)

	// get size of galaxy to generate.
	volume := desiredNumberOfSystems * STANDARD_GALACTIC_RADIUS * STANDARD_GALACTIC_RADIUS * STANDARD_GALACTIC_RADIUS / STANDARD_NUMBER_OF_STAR_SYSTEMS
	for g.Radius*g.Radius*g.Radius < volume {
		g.Radius++
	}
	g.Log("For %d stars, the galaxy should have a radius of about %d parsecs.\n", desiredNumberOfSystems, g.Radius)
	if setupData.Galaxy.Overrides.UseOverrides {
		if setupData.Galaxy.Overrides.Radius != 0 && setupData.Galaxy.Overrides.Radius != g.Radius {
			g.Log("\tBut we are over-riding that to a radius of about %d parsecs.", setupData.Galaxy.Overrides.Radius)
			g.Radius = setupData.Galaxy.Overrides.Radius
		}
	}
	if g.Radius < MIN_RADIUS || g.Radius > MAX_RADIUS {
		return nil, fmt.Errorf("radius %d outside the allowed range of %d to %d parsecs", g.Radius, MIN_RADIUS, MAX_RADIUS)
	}
	galactic_diameter := 2 * g.Radius

	// center translates the coordinates to another frame of reference
	// where 0,0,0 is the "center" of the cluster. it's not needed,
	// but it makes answering the question "is this star within N
	// parsecs of the center of the cluster" easier to answer later on.
	g.Center = Coords{X: g.Radius, Y: g.Radius, Z: g.Radius}

	// get the number of cubic parsecs within a sphere with a radius of galactic_radius parsecs.
	volume = (4 * 314 * g.Radius * g.Radius * g.Radius) / 300

	// the probability of a star system existing at any particular set of x,y,z coordinates is one in chance_of_star.
	chance_of_star := volume / desiredNumberOfSystems
	if chance_of_star < 50 {
		return nil, fmt.Errorf("galactic radius is too small for %d stars (cos %d)", desiredNumberOfSystems, chance_of_star)
	} else if chance_of_star > 3200 {
		return nil, fmt.Errorf("galactic radius is too large for %d stars (cos %d)", desiredNumberOfSystems, chance_of_star)
	}

	// create the star systems and disperse them throughout the cluster
	g.Systems = make(map[int]*StarData)
	g.Translate.XYZToSystem = make(map[string]*StarData)
	locations := make(map[int]*StarData)
	var planetCount [10]int
	mushyDistance, maxDistance := 9+g.Radius*g.Radius, 0
	for i := 0; i < desiredNumberOfSystems; {
		// generate coordinates randomly
		at := Coords{X: rnd(galactic_diameter) - 1, Y: rnd(galactic_diameter) - 1, Z: rnd(galactic_diameter) - 1}
		real_x, real_y, real_z := at.X-g.Center.X, at.Y-g.Center.Y, at.Z-g.Center.Z

		// verify the coordinates are within the galactic boundaries.
		// (sort of - actually just validates the distance from the center)
		sq_distance_from_center := (real_x * real_x) + (real_y * real_y) + (real_z * real_z)
		// TODO: original was >=. Changed this because why?
		if sq_distance_from_center > mushyDistance {
			continue
		}
		// verify that we don't already have a system there
		if _, exists := locations[at.SystemID()]; exists {
			continue
		}
		// all systems need to separated by at least 3 parsecs
		var neighbor *StarData
		for _, o := range g.allSystems {
			if o.Coords.DistanceSquaredTo(at) < 9 {
				neighbor = o
				break
			}
		}
		if neighbor != nil {
			continue
		}

		if sq_distance_from_center > maxDistance {
			maxDistance = sq_distance_from_center
		}

		system, err := NewStar(l, at)
		if err != nil {
			return nil, err
		}
		g.allSystems = append(g.allSystems, system)
		g.Systems[system.key] = system
		g.Translate.XYZToSystem[system.Coords.XYZ()] = system
		g.NumberOfPlanets += len(system.Planets)
		planetCount[len(system.Planets)]++
		locations[system.key] = system

		i++
	}
	l.Printf("Maximum distance from center of cluster is %f parsecs\n", math.Sqrt(float64(maxDistance)))

	for i := 0; i < len(planetCount); i++ {
		if planetCount[i] == 0 {
			continue
		}
		l.Printf("    %3d systems have %d planets\n", planetCount[i], i)
	}
	l.Printf("    %3d planets per system on average\n", g.NumberOfPlanets/len(g.allSystems))

	// create species
	g.Species = make(map[string]*SpeciesData)
	for i, player := range players {
		// player-specified tech levels must sum to 15
		if player.BI+player.GV+player.LS+player.ML != 15 {
			l.Printf("\n\tERROR! ML + GV + LS + BI is not equal to 15!\n\n")
			return nil, fmt.Errorf("species %q: total tech levels must sum up to 15", player.SpeciesName)
		}

		s := &SpeciesData{Number: i + 1, Name: player.SpeciesName}
		s.ID = fmt.Sprintf("SP%02d", s.Number)
		s.Government.Name = player.GovName
		s.Government.Type = player.GovType

		s.Home.System = &NamedSystem{
			Name: player.HomeSystemName,
		}
		s.Home.World = &NamedPlanetData{
			ID:   player.HomePlanetName,
			Name: player.HomePlanetName,
		}

		// set default levels for mining and manufacturing tech
		s.TechLevel[MA], s.TechKnowledge[MA], s.InitTechLevel[MA] = 10, 10, 10
		s.TechLevel[MI], s.TechKnowledge[MI], s.InitTechLevel[MI] = 10, 10, 10

		// set player-specified tech levels
		s.TechLevel[BI], s.TechKnowledge[BI], s.InitTechLevel[BI] = player.BI, player.BI, player.BI
		s.TechLevel[GV], s.TechKnowledge[GV], s.InitTechLevel[GV] = player.GV, player.GV, player.GV
		s.TechLevel[LS], s.TechKnowledge[LS], s.InitTechLevel[LS] = player.LS, player.LS, player.LS
		s.TechLevel[ML], s.TechKnowledge[ML], s.InitTechLevel[ML] = player.ML, player.ML, player.ML

		g.allSpecies = append(g.allSpecies, s)
		g.Species[s.ID] = s
	}

	// randomly assign home systems that aren't too close to other
	// species or wormholes.
	for _, s := range g.allSpecies {
		if s.Home.System.Star != nil {
			continue
		}
		system, err := g.GetRandomSystem(false, false, setupData.Galaxy.MinimumDistance)
		if err != nil {
			return nil, err
		} else if system == nil {
			return nil, fmt.Errorf("assert(getRandomSystem != nil)")
		}
		system.HomeSpecies = s
		s.Home.System.Star = system
	}

	// create home planets from the templates
	for _, s := range g.allSpecies {
		if err := g.ConvertToHomeSystem(l, s); err != nil {
			return nil, err
		}
	}

	// TODO: think about this. put a wormhole into all home systems that points
	// back to the home system. call it an artifact of inventing FTL travel.
	// it would make it simple for aliens to detect that this was a home system.

	// randomly place wormholes between systems.
	// we want about 8% of systems to contain a wormhole.
	desiredNumberOfWormholes := 1 + (8 * (len(g.allSystems) - len(g.Players)) / 100)
	l.Printf("This galaxy wants a total of %d wormholes.\n", desiredNumberOfWormholes)
	// we want wormholes to have a minimum length determined by the cluster size
	minWormholeLength := 20 // galactic_radius + 3 // in parsecs
	// now we actually distribute the wormholes
	for _, system := range g.allSystems {
		if desiredNumberOfWormholes < 1 {
			break
		}
		// don't allow any system to have multiple wormholes
		if system.Wormhole != nil {
			continue
		}
		// randomly fetch a star that doesn't have a wormhole and
		// is at least the minimum distance away
		var endpoint *StarData
		for _, o := range g.allSystems {
			// eliminate endpoints that already have wormholes or are too close
			if o.Wormhole != nil || o.Coords.CloserThan(system.Coords, minWormholeLength) {
				continue
			}
			endpoint = o
			break
		}
		if endpoint == nil { // none of the existing stars met the criteria
			continue
		}
		system.Wormhole, endpoint.Wormhole = endpoint, system
		desiredNumberOfWormholes--
	}

	// create log file for first turn. write home star system data to it.
	for _, s := range g.allSpecies {
		speciesLogFile := filepath.Join(galaxyPath, fmt.Sprintf("sp%02d.log.txt", s.Number))
		fd, err := os.Create(speciesLogFile)
		if err != nil {
			return nil, err
		}
		wl := &Logger{Stdout: fd}
		defer wl.Close()

		wl.Printf("\nScan of home star system for SP %s:\n\n", s.Name)
		_ = s.Home.System.Star.Scan(wl, s)
		wl.Printf("\n")

		l.Printf("Created file %q\n", speciesLogFile)
	}

	l.Printf("This galaxy contains a total of %d stars and %d planets.\n", len(g.allSystems), g.NumberOfPlanets)
	if g.NumberOfWormHoles == 1 {
		l.Printf("The galaxy contains %d natural wormhole.\n\n", g.NumberOfWormHoles)
	} else {
		l.Printf("The galaxy contains %d natural wormholes.\n\n", g.NumberOfWormHoles)
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
		for _, star := range g.AllSystems() {
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

// AllSystems returns a new list containing all of the systems
func (g *GalaxyData) AllSystems() []*StarData {
	if len(g.Systems) != len(g.allSystems) {
		g.Systems = make(map[int]*StarData)
		for _, system := range g.allSystems {
			g.Systems[system.Coords.SystemID()] = system
		}
	}
	systems := make([]*StarData, len(g.allSystems))
	copy(systems, g.allSystems)
	return systems
}

// update the system with values from the system template
func (g *GalaxyData) ConvertToHomeSystem(l *Logger, s *SpeciesData) error {
	homeSystem := s.Home.System.Star
	if homeSystem == nil {
		return fmt.Errorf("assert(homeSystem != nil)")
	}
	fmt.Printf("[convert] %s planets %d %d\n", homeSystem.Coords.XYZ(), len(homeSystem.Planets), len(g.SystemTemplates[len(homeSystem.Planets)]))

	homeSystem.ConvertToHomeSystem(l, s, g.SystemTemplates[len(homeSystem.Planets)])
	pn := s.Home.System.Star.HomePlanetNumber()
	l.Printf("Converted system %s, home planet %d\n", s.Home.System.Star.Coords.XYZ(), pn)

	s.Home.World.Planet = s.Home.System.Star.Planets[homeSystem.HomePlanetIndex()]
	s.AddNamedPlanet(s.Home.World)

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
		l.Printf("\n\tERROR! Planet does not have %s(%s)!\n", s.Gases.Required.Type.String(), s.Gases.Required.Type.Char())
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

	// print summary
	l.Printf("Scan of star system:\n\n")
	_ = s.Home.System.Star.Scan(l, nil)
	l.Printf("\n")
	l.Printf("\n  Summary for species #%d:\n", s.Number)
	l.Printf("\tName of species: %s\n", s.Name)
	l.Printf("\tName of home planet: %s\n", s.Home.World.Name)
	l.Printf("\t\tCoordinates: %s #%d\n", s.Home.World.Planet.Coords, s.Home.World.Planet.Coords.Orbit)
	l.Printf("\tName of government: %s\n", s.Government.Name)
	l.Printf("\tType of government: %s\n\n", s.Government.Type)
	l.Printf("\tTech levels: %s = %d,  %s = %d,  %s = %d\n", techData[MI].name, s.TechLevel[MI], techData[MA].name, s.TechLevel[MA], techData[ML].name, s.TechLevel[ML])
	l.Printf("\t             %s = %d,  %s = %d,  %s = %d\n", techData[GV].name, s.TechLevel[GV], techData[LS].name, s.TechLevel[LS], techData[BI].name, s.TechLevel[BI])
	l.Printf("\n\n\tFor this species, the required gas is %s (%d%%-%d%%).\n", s.Gases.Required.Type.Char(), s.Gases.Required.Min, s.Gases.Required.Max)
	l.Printf("\tGases neutral to species:")
	for _, gasType := range s.Gases.Neutral {
		l.Printf(" %s ", gasType.Char())
	}
	l.Printf("\n\tGases poisonous to species:")
	for _, gasType := range s.Gases.Poison {
		l.Printf(" %s ", gasType.Char())
	}
	l.Printf("\n\n\tInitial mining base = %d.%d. Initial manufacturing base = %d.%d.\n", s.Home.World.MIBase/10, s.Home.World.MIBase%10, s.Home.World.MABase/10, s.Home.World.MABase%10)
	l.Printf("\tIn the first turn, %d raw material units will be produced,\n", (10*s.TechLevel[MI]*s.Home.World.MIBase)/s.Home.World.Planet.MiningDifficulty)
	l.Printf("\tand the total production capacity will be %d.\n\n", (s.TechLevel[MA]*s.Home.World.MABase)/10)

	// set visited_by bit in star data
	s.Home.System.Star.VisitedBy[s.ID] = true

	return nil
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

// GetRandomSystem returns a randomly selected system meeting some criteria.
func (g *GalaxyData) GetRandomSystem(acceptHomeSystem, acceptWormhole bool, noCloserThan int) (*StarData, error) {
	allSystems := g.AllSystems()

	var forbiddenSystems []*StarData
	for _, system := range allSystems {
		if system.HomeSpecies != nil && !acceptHomeSystem {
			forbiddenSystems = append(forbiddenSystems, system)
		} else if system.Wormhole != nil && !acceptWormhole {
			forbiddenSystems = append(forbiddenSystems, system)
		}
	}

	// TODO: shuffle the systems instead of starting at the beginning
	for _, s := range g.allSystems {
		if s == nil {
			continue
		} else if len(s.Planets) < 3 {
			continue
		} else if s.HomeSpecies != nil && !acceptHomeSystem {
			continue
		} else if s.Wormhole != nil && !acceptWormhole {
			continue
		}
		nearForbiddenSystem := false
		for _, system := range forbiddenSystems {
			if s.Coords.CloserThan(system.Coords, noCloserThan) {
				nearForbiddenSystem = true
				break
			}
		}
		if !nearForbiddenSystem {
			return s, nil
		}
	}
	fmt.Printf("[galaxy] getRandomSystem: %d systems (%d forbidden) %d parsecs\n", len(allSystems), len(forbiddenSystems), noCloserThan)
	return nil, fmt.Errorf("all suitable systems are within %d parsecs of each other", noCloserThan)
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

func (g *GalaxyData) GetSystemAt(c Coords) *StarData {
	if len(g.Systems) != len(g.allSystems) {
		g.Systems = make(map[int]*StarData)
		for _, system := range g.allSystems {
			g.Systems[system.Coords.SystemID()] = system
		}
	}
	return g.Systems[c.SystemID()]
}

func (g *GalaxyData) List(listPlanets, listWormholes bool) error {
	// initialize counts
	total_planets := 0
	total_wormstars := 0
	var type_count [10]int
	for i := DWARF; i <= GIANT; i++ {
		type_count[i] = 0
	}

	wormholesListed := make(map[int]bool)

	// for each star, list info
	for _, star := range g.AllSystems() {
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

		if star.Wormhole != nil && !wormholesListed[star.Wormhole.Coords.SystemID()] {
			total_wormstars++
			if listPlanets {
				fmt.Printf("!!! Natural wormhole from here to %s\n", star.Wormhole.Coords.String())
			} else if listWormholes {
				fmt.Printf("Wormhole #%d: from %s to %s\n", total_wormstars, star.Coords, star.Wormhole.Coords.String())
				// log that we've reported on the target to avoid double-reporting
				wormholesListed[star.Coords.SystemID()] = true
				wormholesListed[star.Wormhole.Coords.SystemID()] = true
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
		fmt.Printf("    and %d giant stars, for a total of %d stars.\n", type_count[GIANT], len(g.allSystems))
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
	if len(a) == 0 {
		g.l.Write([]byte(format))
	} else {
		g.l.Printf(format, a...)
	}
}

func (g *GalaxyData) Scan(l *Logger, c Coords) error {
	star := g.GetSystemAt(c)
	if star == nil {
		l.Printf("Scan Report: There is no star system at x = %d, y = %d, z = %d.\n", c.X, c.Y, c.Z)
		return nil
	}
	return star.Scan(l, nil)
}

func (g *GalaxyData) Write(outputPath string, isVerbose bool) error {
	type Ship struct {
		Species string `json:"species"`
		Name    string `json:"name"`
	}
	type Planet struct {
		Orbit                    int               `json:"orbit"`
		Density                  int               `json:"density"`
		Diameter                 int               `json:"diameter"`
		EconomicEfficiency       int               `json:"economic_efficiency"`
		Gases                    []*GasData        `json:"atmosphere,omitempty"`
		Gravity                  int               `json:"gravity"`
		Message                  int               `json:"message_id,omitempty"`
		MiningDifficulty         int               `json:"mining_difficulty"`
		MiningDifficultyIncrease int               `json:"mining_difficulty_increase,omitempty"`
		PressureClass            int               `json:"pressure_class"`
		Special                  PlanetSpecialType `json:"special,omitempty"`
		TemperatureClass         int               `json:"temperature_class"`
		Ships                    []*Ship           `json:"ships,omitempty"`
	}
	type System struct {
		key                 int
		Coords              string    `json:"coords"`
		Message             int       `json:"message_id,omitempty"`
		PotentialHomeSystem bool      `json:"potential_home_system,omitempty"`
		Planets             []*Planet `json:"planets,omitempty"`
		VisitedBy           []string  `json:"visited_by,omitempty"`
	}
	type Wormhole struct {
		key  int
		From int `json:"from"`
		To   int `json:"to"`
	}
	var galaxy struct {
		ID        string      `json:"id"`
		Name      string      `json:"name"`
		Radius    int         `json:"radius"`
		Species   []string    `json:"species"`
		Systems   []*System   `json:"systems"`
		Wormholes []*Wormhole `json:"wormholes"`
	}
	galaxy.ID = g.ID
	galaxy.Name = g.Name
	galaxy.Radius = g.Radius

	// assume that AllSpecies returns a sorted list
	for _, s := range g.AllSpecies() {
		galaxy.Species = append(galaxy.Species, s.Name)
	}

	// assume that AllSystems does not return a sorted list
	allSystems := g.AllSystems()
	for i := 0; i < len(allSystems); i++ {
		for j := i + 1; j < len(allSystems); j++ {
			if allSystems[i].key > allSystems[j].key {
				allSystems[i], allSystems[j] = allSystems[j], allSystems[i]
			}
		}
	}

	for _, star := range allSystems {
		system := &System{key: star.key, Coords: star.Coords.String()}
		if star.Wormhole != nil {
			galaxy.Wormholes = append(galaxy.Wormholes, &Wormhole{
				From: star.key,
				To:   star.Wormhole.key,
			})
		}
		for i, p := range star.Planets {
			planet := &Planet{
				Orbit:                    p.Coords.Orbit,
				Density:                  p.Density,
				Diameter:                 p.Diameter,
				EconomicEfficiency:       p.EconEfficiency,
				Gases:                    p.Gases,
				Gravity:                  p.Gravity,
				Message:                  p.Message,
				MiningDifficulty:         p.MiningDifficulty,
				MiningDifficultyIncrease: p.MDIncrease,
				PressureClass:            p.PressureClass,
				Special:                  p.Special,
				TemperatureClass:         p.TemperatureClass,
			}
			if planet.Orbit != i+1 {
				fmt.Printf("internal error: system %s planet %d has invalid orbit: expected %d: got %d\n", star.Coords.XYZ(), i+1, p.Coords.Orbit, planet.Orbit)
				planet.Orbit = i + 1
			}
			system.PotentialHomeSystem = system.PotentialHomeSystem || p.Special == IDEAL_HOME_PLANET || p.Special == IDEAL_COLONY_PLANET
			system.Planets = append(system.Planets, planet)
		}
		for species, ok := range star.VisitedBy {
			if ok {
				system.VisitedBy = append(system.VisitedBy, species)
			}
		}
		sort.Strings(system.VisitedBy)
		galaxy.Systems = append(galaxy.Systems, system)
	}

	// sort wormholes on output
	for i := 0; i < len(galaxy.Wormholes); i++ {
		for j := i + 1; j < len(galaxy.Wormholes); j++ {
			if galaxy.Wormholes[i].From > galaxy.Wormholes[j].From {
				galaxy.Wormholes[i], galaxy.Wormholes[j] = galaxy.Wormholes[j], galaxy.Wormholes[i]
			}
		}
	}

	galaxyFile := filepath.Join(outputPath, "galaxy.json")
	if isVerbose {
		fmt.Printf("[galaxy] %-30s == %q\n", "GALAXY_FILE", galaxyFile)
	}
	if b, err := json.MarshalIndent(galaxy, "  ", "  "); err != nil {
		return err
	} else if err := ioutil.WriteFile(galaxyFile, b, 0644); err != nil {
		return err
	}

	return nil
}
