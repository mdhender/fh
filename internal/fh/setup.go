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
	IsVerbose       bool `json:"is_verbose,omitempty"`
	NumberOfSpecies int  `json:"number_of_species,omitempty"`
	Galaxy          struct {
		Path      string `json:"path"`
		Name      string `json:"name"`
		Overrides struct {
			UseOverrides  bool `json:"use_overrides"`
			Radius        int  `json:"radius"`
			NumberOfStars int  `json:"number_of_stars"`
		}
		LargeCluster          bool   `json:"large_cluster"`
		Density               string `json:"density"` // should be sparse, normal, high
		ForbidNearbyWormholes bool   `json:"forbid_nearby_wormholes"`
		MinimumDistance       int    `json:"minimum_distance"`
		Radius                struct {
			Minimum int `json:"minimum,omitempty"`
			Maximum int `json:"maximum,omitempty"`
		}
	} `json:"galaxy"`
}

func GetSetup(dir, file string, isVerbose bool) (*SetupData, error) {
	if isVerbose {
		fmt.Printf("[setup] loading configuration from %q\n", file)
	}

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
	if isVerbose {
		fmt.Printf("[setup] %-30s == %q\n", "GALAXY_NAME", setup.Galaxy.Name)
	}

	switch setup.Galaxy.Density {
	case "": // default to normal if missing
		setup.Galaxy.Density = "normal"
	case "high", "normal", "sparse":
		// acceptable values
	default:
		return nil, fmt.Errorf("galaxy.density must be sparse, normal, or high")
	}

	// try to use the location of the setup file to get a default path for the galaxy
	if setup.Galaxy.Path != strings.TrimSpace(setup.Galaxy.Path) {
		return nil, fmt.Errorf("galaxy.path can't have leading or trailing spaces")
	} else if setup.Galaxy.Path == "" || setup.Galaxy.Path == "." || setup.Galaxy.Path == "*" {
		setup.Galaxy.Path = dir
	} else if galaxyPath := filepath.Clean(setup.Galaxy.Path); galaxyPath != setup.Galaxy.Path {
		return nil, fmt.Errorf("galaxy.path %q cleaned to %q", setup.Galaxy.Path, galaxyPath)
	}
	if isVerbose {
		fmt.Printf("[setup] %-30s == %q\n", "GALAXY_PATH", setup.Galaxy.Path)
	}

	if setup.Galaxy.MinimumDistance < 1 {
		return nil, fmt.Errorf("minimum distance must be at least 1")
	} else if setup.Galaxy.MinimumDistance > MAX_RADIUS {
		return nil, fmt.Errorf("minimum distance must be less than %d", MAX_RADIUS)
	}

	if setup.Galaxy.Radius.Minimum < 1 {
		setup.Galaxy.Radius.Minimum = 1
	}
	if setup.Galaxy.Radius.Maximum < setup.Galaxy.Radius.Minimum {
		setup.Galaxy.Radius.Maximum = MAX_RADIUS
	} else if MAX_RADIUS < setup.Galaxy.Radius.Minimum {
		setup.Galaxy.Radius.Maximum = MAX_RADIUS
	}

	if setup.NumberOfSpecies < MIN_SPECIES {
		setup.NumberOfSpecies = MIN_SPECIES
	} else if setup.NumberOfSpecies > MAX_SPECIES {
		return nil, fmt.Errorf("maximum number of species is %d", MAX_SPECIES)
	}

	return &setup, nil
}
