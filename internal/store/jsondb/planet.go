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

type Planet struct {
	Id                int            `json:"id"`
	Coords            Coords         `json:"coords"`
	Orbit             int            `json:"orbit"`
	TemperatureClass  int            `json:"temperature_class"`
	PressureClass     int            `json:"pressure_class"`
	Gases             map[string]int `json:"gases"`
	Diameter          int            `json:"diameter"`
	Gravity           int            `json:"gravity"`
	MiningDifficulty  int            `json:"mining_difficulty"`
	EconEfficiency    int            `json:"econ_efficiency"`
	MdIncrease        int            `json:"md_increase"`
	Message           int            `json:"message"`
	TotalEconomicBase int            `json:"total_econ_base"`
}
