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

import "fmt"

const MAX_SPECIES = 100

type Store struct {
	Version   string    `json:"version"`
	Galaxy    *Galaxy   `json:"galaxy"`
	Systems   []*System `json:"systems"`
	Planets   []*Planet `json:"planets"`
	Species   map[string]*Species
	Commands  map[string]string    `json:"commands"`
	Items     map[string]*Item     `json:"items"`
	Ships     map[string]*ShipData `json:"ships"`
	Tech      map[string]string    `json:"tech"`
	Locations map[string][]string  `json:"locations"`
}

type Coords struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

func (c Coords) Key() string {
	return fmt.Sprintf("%d %d %d", c.X, c.Y, c.Z)
}

type GasMinMax struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type Technology struct {
	Level     int `json:"level"`
	Init      int `json:"init"`
	Knowledge int `json:"knowledge"`
	BankedXp  int `json:"xp"`
}
