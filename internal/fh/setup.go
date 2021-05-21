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

	return &setup, nil
}
