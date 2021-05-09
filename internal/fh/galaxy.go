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
	"io"
	"io/ioutil"
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
	NamedPlanets      map[string]*NamedPlanetData
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

func GenerateGalaxy(setupData *SetupData) (*GalaxyData, error) {
	galaxy := &GalaxyData{
		ID:           setupData.Galaxy.Name,
		Name:         setupData.Galaxy.Name,
		Secret:       "your-private-key-belongs-here",
		Players:      make(map[string]*Player),
		Species:      make(map[string]*SpeciesData),
		Stars:        make(map[string]*StarData),
		NamedPlanets: make(map[string]*NamedPlanetData),
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
			fmt.Printf("Low density option giving, boosting species count to %d\n", adjusted_number_of_species)
			return nil, fmt.Errorf("adjusted number of species must be between %d and %d, inclusive", MIN_SPECIES, MAX_SPECIES)
		}
	}
	galaxy.DNumSpecies = d_num_species

	// get approximate number of star systems to generate
	desired_num_stars := (adjusted_number_of_species * STANDARD_NUMBER_OF_STAR_SYSTEMS) / STANDARD_NUMBER_OF_SPECIES
	fmt.Printf("For %d species, there should be about %d stars.\n", d_num_species, desired_num_stars)
	if setupData.Galaxy.Overrides.UseOverrides {
		if setupData.Galaxy.LowDensity {
			fmt.Printf("For %d species, a low density game needs about %d stars.\n", d_num_species, desired_num_stars)
		} else {
			fmt.Printf("For %d species, a game needs about %d stars.\n", d_num_species, desired_num_stars)
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
		fmt.Printf("For %d stars, the galaxy should have a radius of about %d parsecs.", desired_num_stars, galactic_radius)
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
		if _, exists := galaxy.Stars[XYZToID(x, y, z)]; exists {
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
			dx, dy, dz := star.X-ps.X, star.Y-ps.Y, star.Z-ps.Z
			if distance_squared := (dx * dx) + (dy * dy) + (dz * dz); distance_squared < minWormholeLength*minWormholeLength {
				continue
			}
			worm_star = ps
		}
		if worm_star == nil {
			// wow. none of the existing stars met the criteria
			continue
		}

		star.WormHere = true
		star.WormX, star.WormY, star.WormZ = worm_star.X, worm_star.Y, worm_star.Z

		worm_star.WormHere = true
		worm_star.WormX, worm_star.WormY, worm_star.WormZ = star.X, star.Y, star.Z

		// todo: consider making a number of the wormholes one-way
		galaxy.NumberOfWormHoles++
	}

	for _, star := range galaxy.Stars {
		galaxy.NumberOfPlanets += len(star.Planets)
	}

	fmt.Printf("This galaxy contains a total of %d stars and %d planets.\n", len(galaxy.Stars), galaxy.NumberOfPlanets)
	if galaxy.NumberOfWormHoles == 1 {
		fmt.Printf("The galaxy contains %d natural wormhole.\n\n", galaxy.NumberOfWormHoles)
	} else {
		fmt.Printf("The galaxy contains %d natural wormholes.\n\n", galaxy.NumberOfWormHoles)
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

func (g *GalaxyData) AddSpecies(s *SpeciesData) {
	g.Translate.SpeciesNameToID[s.Name] = s.ID
	g.Translate.IndexToSpeciesID = append(g.Translate.IndexToSpeciesID, s.ID)
	g.Species[s.ID] = s
	g.allSpecies = append(g.allSpecies, s)
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
func (g *GalaxyData) Finish(verbose_mode bool) error {
	if verbose_mode {
		fmt.Printf("\nFinishing up for all species...\n")
	}

	// bump the turn number
	g.TurnNumber++

	// Total economic base includes all of the colonies on the planet, not just the one species.
	total_econ_base := make([]int, len(g.AllPlanets()), len(g.AllPlanets()))

	// add mining difficulty increases for each planet, use the increase calcuated on the prior turn
	for _, planet := range g.AllPlanets() {
		planet.MiningDifficulty += planet.MDIncrease
		planet.MDIncrease = 0
	}

	/* Main loop. For each species, take appropriate action. */
	for _, species := range g.AllSpecies() {
		// check if player submitted orders for this turn.
		var orders_received bool
		if g.TurnNumber == 1 {
			orders_received = true
		} else {
			orderFile := fmt.Sprintf("D:/GoLand/farHorizons/testdata/sp%02d.ord", species.Number)
			_, err := ioutil.ReadFile(orderFile)
			orders_received = err == nil
		}

		// display name of species
		if verbose_mode {
			fmt.Printf("  Now doing SP %s...", species.Name)
			if !orders_received {
				fmt.Printf(" WARNING: player did not submit orders this turn!")
			}
			fmt.Printf("\n")
		}

		// open log file
		//log_file, err := os.Create(fmt.Sprintf("D:/GoLand/farHorizons/testdata/sp%02d.t%04d.log", species.Number, g.TurnNumber))
		//if err != nil {
		//	return err
		//}
		//log_stdout, header_printed := false, false

		if g.TurnNumber == 1 {
			// goto checkForMessage
		}
		/* Check if any ships of this species experienced mishaps. */
		/* Take care of any disbanded colonies. */
		/* Check if this species is the recipient of a transfer of economic units from another species. */
		/* Check if any jump portals of this species were used by aliens. */
		/* Check if any starbases of this species detected the use of gravitic telescopes by aliens. */
		/* Check if this species is the recipient of a tech transfer from another species. */
		/* Calculate tech level increases. */
		/* Notify of any new high tech items. */
		/* Check if this species is the recipient of a knowledge transfer from another species. */
		/* Loop through each nampla for this species. */
		/* Loop through all ships for this species. */
		/* Check if this species has a populated planet that another species tried to land on. */
		/* Check if this species is the recipient of interspecies construction. */
		/* Check if this species is besieging another species and detects forbidden construction, landings, etc. */
		/* Check if this species is the recipient of a message from another species. */
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
	/* Go through all species one more time to update alien contact masks,
	report tech transfer results to donors, and calculate fleet
	maintenance costs. */
	if verbose_mode {
		fmt.Printf("\nNow updating contact masks et al.\n")
	}
	/* Clean up and exit. */
	return nil
}

// GetFirstXYZ returns the first system that is not a home system
// or has a worm hole or is within a given distance of any other home
// system.
func (g *GalaxyData) GetFirstXYZ(d int, forbidWormHoles bool) (int, int, int, error) {
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
			if origin.DistanceSquaredTo(star) < minDSquared {
				nearForbiddenSystem = true
				break
			}
		}
		if !nearForbiddenSystem {
			return origin.X, origin.Y, origin.Z, nil
		}
	}
	return 0, 0, 0, fmt.Errorf("all suitable systems are within %d parsecs of each other", d)
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

func (g *GalaxyData) GetStarAt(x, y, z int) *StarData {
	return g.Stars[XYZToID(x, y, z)]
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
			fmt.Printf("x = %d\ty = %d\tz = %d", star.X, star.Y, star.Z)
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
				fmt.Printf("!!! Natural wormhole from here to %d %d %d\n", star.WormX, star.WormY, star.WormZ)
			} else if listWormholes {
				fmt.Printf("Wormhole #%d: from %d %d %d to %d %d %d\n", total_wormstars, star.X, star.Y, star.Z, star.WormX, star.WormY, star.WormZ)
				// turn off the target's worm flag to avoid double-reporting
				wormSystem := g.GetStarAt(star.WormX, star.WormY, star.WormZ)
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

func (g *GalaxyData) Scan(w io.Writer, x, y, z int) error {
	star := g.GetStarAt(x, y, z)
	if star == nil {
		fmt.Fprintf(w, "Scan Report: There is no star system at x = %d, y = %d, z = %d.\n", x, y, z)
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
