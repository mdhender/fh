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

type Species struct {
	Id                  int
	Name                string
	AutoOrders          bool
	BankedEconomicUnits int
	FleetCost           int
	FleetPercentCost    float64
	Government          struct {
		Name string
		Type string
	}
	Homeworld struct {
		Colony       *Colony
		OriginalBase int
	}
	Relationships map[int]Relationship
	Tech          map[string]*Tech
}

type Relationship int

const (
	None Relationship = iota
	Ally
	Enemy
	Neutral
)

type Tech struct {
	Level     int
	Init      int
	Knowledge int
	BankedXp  int
}

//func (ds *Store) SpeciesMap(id string) []*Species {
//	var result []*Species
//	for _, s := range ds.Sorted.Species {
//		if id == "*" || id == s.Name || id == fmt.Sprintf("%d", s.Id) || id == fmt.Sprintf("SP%d", s.Id) {
//			result = append(result, s)
//		}
//	}
//	if result == nil {
//		return []*Species{}
//	}
//	return result
//}
