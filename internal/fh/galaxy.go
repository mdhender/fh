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

func GenerateGalaxy(logFile io.Writer, setupData *SetupData) (*GalaxyData, error) {
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

		var check struct {
			mishaps         bool /* Check if any ships of this species experienced mishaps. */
			disbanded       bool /* Take care of any disbanded colonies. */
			transferInEU    bool /* Check if this species is the recipient of a transfer of economic units from another species. */
			jumpPortalsUsed bool /* Check if any jump portals of this species were used by aliens. */
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
			messages bool // check if this species is the recipient of a message from another species
		}
		// TODO: try to get straight on Turn 0 being setup and Turn 1 being first turn orders are processed
		check.mishaps = g.TurnNumber > 1         /* Check if any ships of this species experienced mishaps. */
		check.disbanded = g.TurnNumber > 1       /* Take care of any disbanded colonies. */
		check.transferInEU = g.TurnNumber > 1    /* Check if this species is the recipient of a transfer of economic units from another species. */
		check.jumpPortalsUsed = g.TurnNumber > 1 /* Check if any jump portals of this species were used by aliens. */
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
		check.messages = true // always check messages

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
		/* Check if this species is the recipient of a tech transfer from another species. */
		/* Calculate tech level increases. */
		/* Notify of any new high tech items. */
		/* Check if this species is the recipient of a knowledge transfer from another species. */
		/* Loop through each nampla for this species. */
		/* Loop through all ships for this species. */
		/* Check if this species has a populated planet that another species tried to land on. */
		/* Check if this species is the recipient of interspecies construction. */
		/* Check if this species is besieging another species and detects forbidden construction, landings, etc. */

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
	/* Go through all species one more time to update alien contact masks, report tech transfer results to donors, and calculate fleet maintenance costs. */
	if verbose_mode {
		_, _ = fmt.Fprintf(w, "\nNow updating contact masks et al.\n")
	}

	/* Clean up and exit. */
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

func (g *GalaxyData) AddHomePlanets(w io.Writer, galaxyPath string, setupData *SetupData, player *PlayerData, s *SpeciesData) error {
	home_nampla := &NamedPlanetData{ID: player.HomePlanetName}
	home_nampla.Name = player.HomePlanetName
	g.NamedPlanets[home_nampla.ID] = home_nampla

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

	_, _ = fmt.Fprintf(w, "\tTech levels: %s = %d,  %s = %d,  %s = %d\n", TechName[MI], s.TechLevel[MI], TechName[MA], s.TechLevel[MA], TechName[ML], s.TechLevel[ML])
	_, _ = fmt.Fprintf(w, "\t             %s = %d,  %s = %d,  %s = %d\n", TechName[MI], s.TechLevel[GV], TechName[MA], s.TechLevel[LS], TechName[ML], s.TechLevel[BI])

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

func (g *GalaxyData) Write(filename string) error {
	if b, err := json.MarshalIndent(g, "  ", "  "); err != nil {
		return err
	} else if err := ioutil.WriteFile(filename, b, 0644); err != nil {
		return err
	}
	fmt.Printf("Created %q.\n", filename)
	return nil
}
