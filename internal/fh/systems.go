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
	"sort"
)

type Systems struct {
	Data []*StarData
}

func GetSystems(inputPath string, isVerbose bool) (map[string]*StarData, error) {
	data, err := ioutil.ReadFile(filepath.Join(inputPath, "systems.json"))
	if err != nil {
		return nil, err
	}

	var systems []*struct {
		key                 int
		ID                  string    `json:"id"`
		Coords              Coords    `json:"coords"`
		Type                StarType  `json:"type"`
		Color               StarColor `json:"color"`
		Size                int       `json:"size"`
		PotentialHomeSystem bool      `json:"potential_home_system,omitempty"`
		Wormhole            *Coords   `json:"wormhole_exit,omitempty"`
		Message             int       `json:"message,omitempty"`
		VisitedBy           []string  `json:"visited_by,omitempty"`
	}
	if err := json.Unmarshal(data, &systems); err != nil {
		return nil, err
	}

	stars := make(map[string]*StarData)
	for _, system := range systems {
		star := &StarData{
			ID:     system.ID,
			Coords: system.Coords,
			Type:   system.Type,
			Color:  system.Color,
			Size:   system.Size,
			//Wormhole: system.Wormhole,
			Message: system.Message,
		}
		stars[star.ID] = star
	}
	return stars, nil
}

func (s *Systems) Write(outputPath string, isVerbose bool) error {
	type System struct {
		key                 int
		ID                  string    `json:"id"`
		Coords              Coords    `json:"coords"`
		Type                StarType  `json:"type"`
		Color               StarColor `json:"color"`
		Size                int       `json:"size"`
		PotentialHomeSystem bool      `json:"potential_home_system,omitempty"`
		Wormhole            *Coords   `json:"wormhole_exit,omitempty"`
		Message             int       `json:"message,omitempty"`
		VisitedBy           []string  `json:"visited_by,omitempty"`
	}
	var data []*System
	for _, star := range s.Data {
		o := &System{
			key:                 star.Coords.SystemID(),
			ID:                  star.Coords.String(),
			Coords:              Coords{X: star.Coords.X, Y: star.Coords.Y, Z: star.Coords.Z},
			Type:                star.Type,
			Color:               star.Color,
			Size:                star.Size,
			PotentialHomeSystem: star.HomeSystem,
			Message:             star.Message,
		}
		if star.Wormhole != nil {
			o.Wormhole = &Coords{X: star.Wormhole.Coords.X, Y: star.Wormhole.Coords.Y, Z: star.Wormhole.Coords.Z}
		}
		for visitor := range star.VisitedBy {
			o.VisitedBy = append(o.VisitedBy, visitor)
		}
		sort.Strings(o.VisitedBy)
		data = append(data, o)
	}
	for i := 0; i < len(data); i++ {
		for j := i + 1; j < len(data); j++ {
			if data[i].key > data[j].key {
				data[i], data[j] = data[j], data[i]
			}
		}
	}

	systemsFile := filepath.Join(outputPath, "systems.json")
	if isVerbose {
		fmt.Printf("[system] %-30s == %q\n", "SYSTEMS_FILE", systemsFile)
	}
	if b, err := json.MarshalIndent(&data, "  ", "  "); err != nil {
		return err
	} else if err := ioutil.WriteFile(systemsFile, b, 0644); err != nil {
		return err
	}
	fmt.Printf("[systems] wrote %d/%d stars\n", len(s.Data), len(data))

	return nil
}
