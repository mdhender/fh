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

package memory

type Planet struct {
	Id                       int
	System                   *System
	Coords                   Coords
	Diameter                 int            `json:"diameter"`
	EconEfficiency           float64        `json:"econ_efficiency"`
	Gases                    map[string]int `json:"gases"`
	Gravity                  float64        `json:"gravity"`
	Message                  int            `json:"message"`
	MiningDifficulty         float64        `json:"mining_difficulty"`
	MiningDifficultyIncrease float64        `json:"md_increase"`
	PressureClass            int            `json:"pressure_class"`
	TemperatureClass         int            `json:"temperature_class"`
	Colonies                 []*Colony
	Ships                    []*Ship
}

// Less is a helper for sorting
func (p *Planet) Less(p2 *Planet) bool {
	return p.Coords.Less(p2.Coords)
}

type GasType int

const (
	GTNone GasType = iota
)
