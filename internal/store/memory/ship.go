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

import "fmt"

type Ship struct {
	Id                 int
	Species            *Species
	Name               string
	Coords             Coords // current location
	Age                int
	Capacity           int
	Code               string
	DeepSpace          bool
	Destination        *Coords // set only when there is a jump target
	ForcedJump         bool
	FTL                bool
	Hiding             bool
	Landed             bool
	MaintenanceCost    int
	MALevel            int
	Orbiting           bool
	WithdrewFromCombat bool
	Inventory          map[string]*Item
}

func (s *Ship) Key() string {
	return fmt.Sprintf("%d:%d", s.Species.Id, s.Id)
}
