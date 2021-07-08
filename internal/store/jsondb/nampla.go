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

type NamedPlanet struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Coords   Coords `json:"coords"`
	Orbit    int    `json:"orbit"`
	Status   struct {
		HomePlanet      bool `json:"home_planet"`
		Colony          bool `json:"colony"`
		Populated       bool `json:"populated"`
		MiningColony    bool `json:"mining_colony"`
		ResortColony    bool `json:"resort_colony"`
		DisbandedColony bool `json:"disbanded_colony"`
	} `json:"status"`
	Hiding       bool           `json:"hiding"`
	Hidden       int            `json:"hidden"`
	PlanetIndex  int            `json:"planet_index"`
	SiegeEff     int            `json:"siege_eff"`
	Shipyards    int            `json:"shipyards"`
	IUsNeeded    int            `json:"IUs_needed"`
	AUsNeeded    int            `json:"AUs_needed"`
	AutoIUs      int            `json:"auto_IUs"`
	AutoAUs      int            `json:"auto_AUs"`
	IUsToInstall int            `json:"IUs_to_install"`
	AUsToInstall int            `json:"AUs_to_install"`
	MiBase       int            `json:"mi_base"`
	MaBase       int            `json:"ma_base"`
	PopUnits     int            `json:"pop_units"`
	UseOnAmbush  int            `json:"use_on_ambush"`
	Message      int            `json:"message"`
	Inventory    map[string]int `json:"inventory"`
}
