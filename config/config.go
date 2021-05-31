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

// Package config defines a struct for configuration data along with
// a routine to load that data from a JSON file.
package config

import (
	"encoding/json"
	"fmt"
	"github.com/mdhender/fh/stdlib"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	RESOURCE_MIN = 0
	RESOURCE_MAX = 9
	SEPARATION_MAX = 20
	SPECIES_MIN  = 1
	SPECIES_MAX  = 18
)

type Config struct {
	IsVerbose bool `json:"is_verbose,omitempty"`
	Galaxy    struct {
		Path            string `json:"path"`
		Name            string `json:"name"`
		Location        string `json:"location,omitempty"`       // not case sensitive
		ResourceLevel   int    `json:"resource_level,omitempty"` // range 0..9
		Separation struct {
			HomeSystems int `json:"home_systems,omitempty"`
			Wormholes   int `json:"wormholes,omitempty"`
		} `json:"separation,omitempty"`
	} `json:"galaxy"`
	Species struct {
		Number int `json:"number"`
	} `json:"species"`
}

func Get(file string, defaultGalaxyPath string, isVerbose bool) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err = stdlib.IsValidName(cfg.Galaxy.Name); err != nil {
		return nil, fmt.Errorf("galaxy: %w", err)
	}

	// try to use the location of the setup file to get a default path for the galaxy
	if cfg.Galaxy.Path != strings.TrimSpace(cfg.Galaxy.Path) {
		return nil, fmt.Errorf("galaxy.path can't have leading or trailing spaces")
	} else if cfg.Galaxy.Path == "" || cfg.Galaxy.Path == "." || cfg.Galaxy.Path == "*" {
		cfg.Galaxy.Path = defaultGalaxyPath
	} else if galaxyPath := filepath.Clean(cfg.Galaxy.Path); galaxyPath != cfg.Galaxy.Path {
		return nil, fmt.Errorf("galaxy.path %q cleaned to %q", cfg.Galaxy.Path, galaxyPath)
	}

	if cfg.Galaxy.Location == "" {
		cfg.Galaxy.Location = "rim"
	}
	cfg.Galaxy.Location = strings.ToLower(cfg.Galaxy.Location)
	switch cfg.Galaxy.Location {
	case "outer rim", "rim", "inner rim", "outer core":
		break
	default:
		return nil, fmt.Errorf("location must be one of 'outer rim', 'rim', 'inner rim', 'outer core'")
	}

	if cfg.Galaxy.ResourceLevel < RESOURCE_MIN || cfg.Galaxy.ResourceLevel > RESOURCE_MAX {
		return nil, fmt.Errorf("resource_level must be between %d and %d", RESOURCE_MIN, RESOURCE_MAX)
	}

	if cfg.Galaxy.Separation.HomeSystems < 0 {
		return nil, fmt.Errorf("minimum distance from home systems must be at least 0")
	} else if cfg.Galaxy.Separation.HomeSystems > SEPARATION_MAX {
		return nil, fmt.Errorf("minimum distance from home systems must be less than %d", SEPARATION_MAX)
	}

	if cfg.Galaxy.Separation.Wormholes < 0 {
		return nil, fmt.Errorf("minimum distance from wormholes must be at least 0")
	} else if cfg.Galaxy.Separation.Wormholes > SEPARATION_MAX {
		return nil, fmt.Errorf("minimum distance from wormholes must be less than %d", SEPARATION_MAX)
	}

	if cfg.Species.Number == 0 {
		cfg.Species.Number = SPECIES_MIN
	} else if cfg.Species.Number < SPECIES_MIN || cfg.Species.Number > SPECIES_MAX {
		return nil, fmt.Errorf("species.number must be between %d and %d", SPECIES_MIN, SPECIES_MAX)
	}

	return &cfg, nil
}
