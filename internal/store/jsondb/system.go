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

type System struct {
	Id         int      `json:"id"`
	Key        string   `json:"key"`
	Coords     Coords   `json:"coords"`
	Type       string   `json:"type"`
	Color      string   `json:"color"`
	Size       int      `json:"size"`
	HomeSystem bool     `json:"home_system"`
	Planets    []int    `json:"planets"`
	VisitedBy  []string `json:"visited_by"`
	Message    int      `json:"message"`
}
