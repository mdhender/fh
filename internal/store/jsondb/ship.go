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

type Ship struct {
	Id             int            `json:"id"`
	Name           string         `json:"name"`
	Location       string         `json:"location"`
	Coords         Coords         `json:"coords"`
	Orbit          int            `json:"orbit"`
	Class          string         `json:"class"`
	Type           string         `json:"type"`
	Tonnage        int            `json:"tonnage"`
	Age            int            `json:"age"`
	Status         string         `json:"status"`
	Dest           Coords         `json:"dest"`
	LoadingPoint   int            `json:"loading_point"`
	UnloadingPoint int            `json:"unloading_point"`
	RemainingCost  int            `json:"remaining_cost"`
	Message        int            `json:"message"`
	Inventory      map[string]int `json:"inventory"`
}

type ShipData struct {
	Class                 string `json:"class"`
	MinManufacturingLevel int    `json:"min_ma"`
	CostFtl               int    `json:"cost_ftl"`
	CostSublight          int    `json:"cost_sublight"`
	Tonnage               int    `json:"tonnage"`
	CarryingCapacity      int    `json:"carrying_capacity"`
}
