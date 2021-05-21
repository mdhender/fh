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
	"strings"
)

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

func GetPlayers(file string, isVerbose bool) ([]*PlayerData, error) {
	if isVerbose {
		fmt.Printf("[players] loading configuration from %q\n", file)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var players []*PlayerData
	if err := json.Unmarshal(data, &players); err != nil {
		return nil, err
	}

	emails := make(map[string]bool)
	homePlanetName := make(map[string]bool)
	speciesNames := make(map[string]bool)

	var errors []error
	for i, player := range players {
		if player.Email != strings.TrimSpace(player.Email) {
			errors = append(errors, fmt.Errorf("player %d: email address must not have leading or trailing spaces", i+1))
		} else if exists := emails[player.Email]; exists {
			errors = append(errors, fmt.Errorf("player %d: duplicate email address %q", i+1, player.Email))
		} else {
			emails[player.Email] = true
		}

		if err = IsValidName(player.SpeciesName); err != nil {
			errors = append(errors, fmt.Errorf("player %d: species name: %w", i+1, err))
		} else if len(player.SpeciesName) < 5 {
			errors = append(errors, fmt.Errorf("player %d: species name %q too short (min 5 chars allowed)", i+1, player.SpeciesName))
		} else if len(player.SpeciesName) > 31 {
			errors = append(errors, fmt.Errorf("player %d: species name %q too long (max 31 chars allowed)", i+1, player.SpeciesName))
		} else if i := strings.IndexAny(player.SpeciesName, "$!`\"{}\\"); i != -1 {
			errors = append(errors, fmt.Errorf("player %d: invalid character %q in species name", i+1, player.SpeciesName[i]))
		} else if exists := speciesNames[player.SpeciesName]; exists {
			errors = append(errors, fmt.Errorf("player %d: duplicate species name %q", i+1, player.SpeciesName))
		} else {
			speciesNames[player.Email] = true
		}

		if err = IsValidName(player.HomePlanetName); err != nil {
			errors = append(errors, fmt.Errorf("player %d: home planet name: %w", i+1, err))
		} else if len(player.HomePlanetName) > 31 {
			errors = append(errors, fmt.Errorf("player %d: home planet name %q too long (max 31 chars allowed)", i+1, player.HomePlanetName))
		} else if err = IsValidName(player.HomePlanetName); err != nil {
			errors = append(errors, fmt.Errorf("player %d: home planet name: %w", i+1, err))
		} else if i := strings.IndexAny(player.HomePlanetName, "$!`\"{}\\"); i != -1 {
			errors = append(errors, fmt.Errorf("player %d: invalid character %q in home planet name", i+1, player.HomePlanetName[i]))
		} else if exists := homePlanetName[player.HomePlanetName]; exists {
			errors = append(errors, fmt.Errorf("player %d: duplicate home planet name %q", i+1, player.HomePlanetName))
		} else {
			homePlanetName[player.HomePlanetName] = true
		}

		if err = IsValidName(player.GovName); err != nil {
			errors = append(errors, fmt.Errorf("player %d: government name: %w", i+1, err))
		} else if len(player.GovName) > 31 {
			errors = append(errors, fmt.Errorf("player %d: government name %q too long (max 31 chars allowed)", i+1, player.GovName))
		} else if err = IsValidName(player.GovName); err != nil {
			errors = append(errors, fmt.Errorf("player %d: government name: %w", i+1, err))
		} else if i := strings.IndexAny(player.GovName, "$!`\"{}\\"); i != -1 {
			errors = append(errors, fmt.Errorf("player %d: invalid character %q in government name", i+1, player.GovName[i]))
		}

		if err = IsValidName(player.GovType); err != nil {
			errors = append(errors, fmt.Errorf("player %d: government type: %w", i+1, err))
		} else if len(player.GovType) > 31 {
			errors = append(errors, fmt.Errorf("player %d: government type %q too long (max 31 chars allowed)", i+1, player.GovType))
		} else if err = IsValidName(player.GovType); err != nil {
			errors = append(errors, fmt.Errorf("player %d: government type: %w", i+1, err))
		} else if i := strings.IndexAny(player.GovType, "$!`\"{}\\"); i != -1 {
			errors = append(errors, fmt.Errorf("player %d: invalid character %q in government type", i+1, player.GovType[i]))
		}

		if player.BI+player.GV+player.LS+player.ML != 15 {
			errors = append(errors, fmt.Errorf("player %d: the tech levels must sum to 15", i+1))
		}
	}

	if len(errors) == 1 {
		return nil, errors[0]
	} else if len(errors) > 1 {
		for _, err := range errors {
			fmt.Printf("[players] %+v\n", err)
		}
		return nil, fmt.Errorf("player file contained too many errors")
	}

	if isVerbose {
		fmt.Printf("[players] loaded %d players\n", len(players))
	}

	return players, nil
}
