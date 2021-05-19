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
	"io/ioutil"
	"path/filepath"
	"strings"
)

type SetupData struct {
	Galaxy struct {
		Path      string `json:"path"`
		Name      string `json:"name"`
		Overrides struct {
			UseOverrides  bool `json:"use_overrides"`
			Radius        int  `json:"radius"`
			NumberOfStars int  `json:"number_of_stars"`
		}
		LowDensity            bool `json:"low_density"`
		ForbidNearbyWormholes bool `json:"forbid_nearby_wormholes"`
		MinimumDistance       int  `json:"minimum_distance"`
	} `json:"galaxy"`
	Players []PlayerData `json:"players"`
}

type PlayerData struct {
	Email          string `json:"email"`
	SpeciesName    string `json:"species_name"`
	HomePlanetName string `json:"home_planet_name"`
	GovName        string `json:"government_name"`
	GovType        string `json:"government_type"`
	ML             int    `json:"military_level"`
	GV             int    `json:"gravitics_level"`
	LS             int    `json:"life_support_level"`
	BI             int    `json:"biology_level"`
}

func GetSetup(dir, file string) (*SetupData, error) {
	fmt.Printf("[setup] loading configuration from %q\n", file)

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var setup SetupData
	if err := json.Unmarshal(data, &setup); err != nil {
		return nil, err
	}

	if err = IsValidName(setup.Galaxy.Name); err != nil {
		return nil, fmt.Errorf("galaxy: %w", err)
	}
	fmt.Printf("[setup] %-30s == %q\n", "GALAXY_NAME", setup.Galaxy.Name)

	if setup.Galaxy.Path != strings.TrimSpace(setup.Galaxy.Path) {
		return nil, fmt.Errorf("galaxy.path can't have leading or trailing spaces")
	} else if setup.Galaxy.Path == "" || setup.Galaxy.Path == "." || setup.Galaxy.Path == "*" {
		setup.Galaxy.Path = dir
	} else if galaxyPath := filepath.Clean(setup.Galaxy.Path); galaxyPath != setup.Galaxy.Path {
		return nil, fmt.Errorf("galaxy.path %q cleaned to %q", setup.Galaxy.Path, galaxyPath)
	}
	fmt.Printf("[setup] %-30s == %q\n", "GALAXY_PATH", setup.Galaxy.Path)

	if setup.Galaxy.MinimumDistance < 1 {
		return nil, fmt.Errorf("minimum distance must be at least 1")
	} else if setup.Galaxy.MinimumDistance > MAX_RADIUS {
		return nil, fmt.Errorf("minimum distance must be less than %d", MAX_RADIUS)
	}

	emails := make(map[string]bool)
	homePlanetName := make(map[string]bool)
	speciesNames := make(map[string]bool)
	for i, player := range setup.Players {
		if player.Email != strings.TrimSpace(player.Email) {
			return nil, fmt.Errorf("player %d: email address must not have leading or trailing spaces", i+1)
		} else if exists := emails[player.Email]; exists {
			return nil, fmt.Errorf("player %d: duplicate email address %q", i+1, player.Email)
		}
		emails[player.Email] = true

		if err = IsValidName(player.SpeciesName); err != nil {
			return nil, fmt.Errorf("player %d: species name: %w", i+1, err)
		} else if len(player.SpeciesName) < 5 {
			return nil, fmt.Errorf("player %d: species name %q too short (min 5 chars required)", i+1, player.SpeciesName)
		} else if len(player.SpeciesName) > 31 {
			return nil, fmt.Errorf("player %d: species name %q too long (max 31 chars required)", i+1, player.SpeciesName)
		} else if err = IsValidName(player.SpeciesName); err != nil {
			return nil, fmt.Errorf("player %d: species name: %w", i+1, err)
		} else if i := strings.IndexAny(player.SpeciesName, "$!`\"{}\\"); i != -1 {
			return nil, fmt.Errorf("player %d: invalid character %q in species name", i+1, player.SpeciesName[i])
		} else if exists := speciesNames[player.SpeciesName]; exists {
			return nil, fmt.Errorf("player %d: duplicate species name %q", i+1, player.SpeciesName)
		}
		speciesNames[player.Email] = true

		if err = IsValidName(player.SpeciesName); err != nil {
			return nil, fmt.Errorf("player %d: home planet name: %w", i+1, err)
		} else if len(player.HomePlanetName) > 31 {
			return nil, fmt.Errorf("player %d: home planet name %q too long (max 31 chars required)", i+1, player.HomePlanetName)
		} else if err = IsValidName(player.HomePlanetName); err != nil {
			return nil, fmt.Errorf("player %d: home planet name: %w", i+1, err)
		} else if i := strings.IndexAny(player.HomePlanetName, "$!`\"{}\\"); i != -1 {
			return nil, fmt.Errorf("player %d: invalid character %q in home planet name", i+1, player.HomePlanetName[i])
		} else if exists := homePlanetName[player.HomePlanetName]; exists {
			return nil, fmt.Errorf("player %d: duplicate home planet name %q", i+1, player.HomePlanetName)
		}
		homePlanetName[player.HomePlanetName] = true

		if err = IsValidName(player.SpeciesName); err != nil {
			return nil, fmt.Errorf("player %d: government name: %w", i+1, err)
		} else if len(player.GovName) > 31 {
			return nil, fmt.Errorf("player %d: government name %q too long (max 31 chars required)", i+1, player.GovName)
		} else if err = IsValidName(player.GovName); err != nil {
			return nil, fmt.Errorf("player %d: government name: %w", i+1, err)
		} else if i := strings.IndexAny(player.GovName, "$!`\"{}\\"); i != -1 {
			return nil, fmt.Errorf("player %d: invalid character %q in home planet name", i+1, player.GovName[i])
		}

		if err = IsValidName(player.SpeciesName); err != nil {
			return nil, fmt.Errorf("player %d: government type: %w", i+1, err)
		} else if len(player.GovType) > 31 {
			return nil, fmt.Errorf("player %d: government type %q too long (max 31 chars required)", i+1, player.GovType)
		} else if err = IsValidName(player.GovName); err != nil {
			return nil, fmt.Errorf("player %d: government type: %w", i+1, err)
		} else if i := strings.IndexAny(player.GovType, "$!`\"{}\\"); i != -1 {
			return nil, fmt.Errorf("player %d: invalid character %q in government type", i+1, player.GovType[i])
		}

		if player.BI+player.GV+player.LS+player.ML != 15 {
			return nil, fmt.Errorf("player %d: the tech levels must sum to 15", i+1)
		}
	}
	return &setup, nil
}
