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

package jsondb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

func Read(filename string) (*Store, error) {
	var ds Store
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	} else if err = json.Unmarshal(b, &ds); err != nil {
		return nil, err
	}
	if ds.Locations == nil {
		ds.Locations = make(map[string][]string)
	}

	// TODO: get better about the zero vs one based indexing
	ds.Planets = convertZeroBasedPlanetId(ds.Planets)

	// add coordinates to the planet structure
	// (they're not present in the dat2json export)
	for _, system := range ds.Systems {
		for orbit, planet := range system.Planets {
			ds.Planets[planet].Coords = system.Coords
			ds.Planets[planet].Orbit = orbit + 1
		}
	}

	// validate planet data
	for _, planet := range ds.Planets {
		if planet.Id < 1 { // ignore the zero-th planet
			continue
		}
		var total int
		for gas, percentage := range planet.Gases {
			switch gas {
			case "NH3": // Ammonia
			case "CO2": // Carbon Dioxide
			case "Cl2": // Chlorine
			case "F2": // Fluorine
			case "He": // Helium
			case "H2": // Hydrogen
			case "HCl": // Hydrogen Chloride
			case "H2S": // Hydrogen Sulfide
			case "CH4": // Methane
			case "N2": // Nitrogen
			case "O2": // Oxygen
			case "SO2": // Sulfur Dioxide
			case "H2O": // Water or Steam
			default:
				return nil, fmt.Errorf("unknown gas %q on planet %d", gas, planet.Id)
			}
			if percentage < 1 || percentage > 100 {
				return nil, fmt.Errorf("invalid percentage %d for gas %q on planet %d", percentage, gas, planet.Id)
			}
			total += percentage
		}
		if total > 0 && total != 100 {
			return nil, fmt.Errorf("invalid percentage %d for gases on planet %d", total, planet.Id)
		}
	}

	// validate and normalize species data
	for key, sp := range ds.Species {
		if sp.Id < 1 {
			return nil, fmt.Errorf("invalid species id %d", sp.Id)
		}
		if key != fmt.Sprintf("SP%02d", sp.Id) {
			return nil, fmt.Errorf("invalid key %q for species %d", key, sp.Id)
		}
		sp.Key = key
		sp.Aliens = make(map[int]string)

		if sp.Homeworld.Id == 0 {
			for _, p := range ds.Planets {
				if p.Coords == sp.Homeworld.Coords && p.Orbit == sp.Homeworld.Orbit {
					sp.Homeworld.Id = p.Id
					break
				}
			}
		}
		// parse Contact, then Ally, then Neutral, then Enemy on the off-chance
		// that someone edits the JSON and assigns the species to multiple categories.
		for _, a := range sp.Contacts {
			if !strings.HasPrefix(a, "SP") {
				continue
			}
			id, err := strconv.Atoi(a[2:])
			if err != nil {
				continue
			} else if id < 1 || MAX_SPECIES < id {
				continue
			}
			sp.Aliens[id] = "neutral"
		}
		for _, a := range sp.Allies {
			if !strings.HasPrefix(a, "SP") {
				continue
			}
			id, err := strconv.Atoi(a[2:])
			if err != nil {
				continue
			} else if id < 1 || MAX_SPECIES < id {
				continue
			}
			sp.Aliens[id] = "ally"
		}
		for _, a := range sp.Enemies {
			if !strings.HasPrefix(a, "SP") {
				continue
			}
			id, err := strconv.Atoi(a[2:])
			if err != nil {
				continue
			} else if id < 1 || MAX_SPECIES < id {
				continue
			}
			sp.Aliens[id] = "enemy"
		}
	}

	return &ds, nil
}

func (ds *Store) Write(filename string) error {
	b, err := json.MarshalIndent(ds, "", "  ")
	if err != nil {
		return err
	}
	if filename == "*stdout*" {
		fmt.Println(string(b))
		return nil
	}
	err = ioutil.WriteFile(filename, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

// convertZeroBasedPlanetId changes the planets array to be 1 based.
// side effect: deletes planets with invalid id.
func convertZeroBasedPlanetId(zero []*Planet) (one []*Planet) {
	maxPlanetId := len(zero)
	for _, p := range zero {
		if maxPlanetId < p.Id {
			maxPlanetId = p.Id
		}
	}
	one = make([]*Planet, maxPlanetId+1, maxPlanetId+1)
	for _, p := range zero {
		if p.Id > 0 {
			one[p.Id] = p
		}
	}
	for i := range one {
		if one[i] == nil {
			one[i] = &Planet{Id: i}
		}
	}
	return one
}
