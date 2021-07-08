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

type Galaxy struct {
	TurnNumber        int `json:"turn_number"`
	DNumSpecies       int `json:"d_num_species"`
	NumSpecies        int `json:"num_species"`
	Radius            int `json:"radius"`
	MinRadius         int `json:"min_radius"`
	MaxRadius         int `json:"max_radius"`
	StdNumStars       int `json:"std_num_stars"`
	MinStars          int `json:"min_stars"`
	MaxStars          int `json:"max_stars"`
	StdNumSpecies     int `json:"std_num_species"`
	MinSpecies        int `json:"min_species"`
	MaxSpecies        int `json:"max_species"`
	MaxItems          int `json:"max_items"`
	MaxLocations      int `json:"max_locations"`
	MaxTransactions   int `json:"max_transactions"`
	NumCommands       int `json:"num_commands"`
	NumContactWords   int `json:"num_contact_words"`
	NumShipClasses    int `json:"num_ship_classes"`
	SizeofChar        int `json:"sizeof char"`
	SizeofInt         int `json:"sizeof int"`
	SizeofLong        int `json:"sizeof long"`
	SizeofShort       int `json:"sizeof short"`
	SizeofGalaxyData  int `json:"sizeof galaxy_data"`
	SizeofStarData    int `json:"sizeof star_data"`
	SizeofPlanetData  int `json:"sizeof planet_data"`
	SizeofNamplaData  int `json:"sizeof nampla_data"`
	SizeofSpeciesData int `json:"sizeof species_data"`
	SizeofShipData    int `json:"sizeof ship_data"`
	SizeofTransData   int `json:"sizeof trans_data"`
}
