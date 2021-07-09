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

package main

import (
	"fmt"
	"github.com/mdhender/fh/internal/store/jsondb"
)

func Stats(jdb *jsondb.Store) {
	// initialize data
	n_species := 0
	all_production := 0
	min_production := 1000000000
	max_production := 0
	all_warship_tons := 0
	all_starbase_tons := 0
	all_transport_tons := 0
	n_warships := 0
	min_warships := 32000
	max_warships := 0
	n_starbases := 0
	min_starbases := 32000
	max_starbases := 0
	n_transports := 0
	min_transports := 32000
	max_transports := 0
	n_pop_pl := 0
	min_pop_pl := 32000
	max_pop_pl := 0
	n_yards := 0
	min_yards := 32000
	max_yards := 0
	totalBankedEconUnits, minBankedEconUnits, maxBankedEconUnits := 0, 0, 0

	var all_tech_level, max_tech_level, min_tech_level [6]int
	for i := 0; i < 6; i++ {
		all_tech_level[i] = 0
		min_tech_level[i] = 32000
		max_tech_level[i] = 0
	}

	// print header
	fmt.Printf("SP Species               Tech Levels        Total  Num Num  Num  Offen.  Defen.  Econ\n")
	fmt.Printf(" # Name             MI  MA  ML  GV  LS  BI  Prod.  Pls Shps Yrds  Power   Power  Units\n")
	fmt.Printf("----------------------------------------------------------------------------------------\n")

	// main loop. for each species, take appropriate action.
	for species_number := 1; species_number <= jdb.Galaxy.NumSpecies; species_number++ {
		n_species++

		species := jdb.Species[fmt.Sprintf("SP%02d", species_number)]

		// get fleet maintenance cost
		fleet_percent_cost := species.FleetPercentCost

		if fleet_percent_cost > 10000 {
			fleet_percent_cost = 10000
		}

		// print species data
		fmt.Printf("%2d", species_number)
		fmt.Printf(" %-15.15s", species.Name)

		for i := 0; i < 6; i++ {
			var techLevel int
			switch i {
			case 0:
				techLevel = species.Tech.Mining.Level
			case 1:
				techLevel = species.Tech.Manufacturing.Level
			case 2:
				techLevel = species.Tech.Military.Level
			case 3:
				techLevel = species.Tech.Gravitics.Level
			case 4:
				techLevel = species.Tech.LifeSupport.Level
			case 5:
				techLevel = species.Tech.Biology.Level
			}
			fmt.Printf("%4d", techLevel)
			all_tech_level[i] += techLevel
			if techLevel < min_tech_level[i] {
				min_tech_level[i] = techLevel
			}
			if techLevel > max_tech_level[i] {
				max_tech_level[i] = techLevel
			}
		}

		// Get stats for namplas
		total_production := 0
		total_defensive_power := 0
		num_yards := 0
		num_pop_planets := 0
		home_planet := jdb.Planets[species.Homeworld.Id]
		for _, nampla := range species.NamedPlanets {
			if nampla.Orbit == 99 {
				continue
			}

			// TODO: how are these two different
			num_yards += nampla.Shipyards
			n_yards += nampla.Shipyards

			planet := jdb.Planets[nampla.PlanetIndex]

			raw_material_units := (10 * species.Tech.Mining.Level * nampla.MiBase) / planet.MiningDifficulty

			production_capacity := (species.Tech.Manufacturing.Level * (nampla.MaBase)) / 10

			ls_needed := LifeSupportNeeded(species, home_planet, planet)

			production_penalty := 0
			if ls_needed != 0 {
				production_penalty = (100 * ls_needed) / species.Tech.LifeSupport.Level
			}

			raw_material_units -= (production_penalty * raw_material_units) / 100
			raw_material_units = ((planet.EconEfficiency * raw_material_units) + 50) / 100
			production_capacity -= (production_penalty * production_capacity) / 100
			production_capacity = ((planet.EconEfficiency * production_capacity) + 50) / 100

			var n1 int
			if nampla.Status.MiningColony {
				n1 = (2 * raw_material_units) / 3
			} else if nampla.Status.ResortColony {
				n1 = (2 * production_capacity) / 3
			} else if production_capacity > raw_material_units {
				n1 = raw_material_units
			} else {
				n1 = production_capacity
			}
			total_production += n1
			total_production -= ((fleet_percent_cost * n1) + 5000) / 10000

			tons := nampla.Inventory["PD"] / 200
			if tons < 1 && nampla.Inventory["PD"] > 0 {
				tons = 1
			}
			total_defensive_power += power(tons)

			if nampla.Status.Populated {
				n_pop_pl++
				num_pop_planets++
			}
		}

		fmt.Printf("%7d%4d", total_production, num_pop_planets)

		if total_production < min_production {
			min_production = total_production
		}
		if total_production > max_production {
			max_production = total_production
		}

		if num_pop_planets < min_pop_pl {
			min_pop_pl = num_pop_planets
		}
		if num_pop_planets > max_pop_pl {
			max_pop_pl = num_pop_planets
		}

		if num_yards < min_yards {
			min_yards = num_yards
		}
		if num_yards > max_yards {
			max_yards = num_yards
		}

		all_production += total_production

		// get stats for ships
		num_ships := 0 // number of ships and starbases
		ntr := 0       // number of transports
		nba := 0       // number of starbases
		nwa := 0       // number of warships
		total_tonnage := 0
		total_offensive_power := 0
		for _, ship := range species.Ships {
			if ship.Orbit == 99 { // TODO: why exclude 99?
				continue
			}

			if ship.Status == "UNDER_CONSTRUCTION" {
				continue
			}

			num_ships++
			total_tonnage += ship.Tonnage

			// TODO: should this be ship.Class for STARBASE?
			if ship.Type == "STARBASE" {
				total_defensive_power += power(ship.Tonnage)
				all_starbase_tons += ship.Tonnage
				n_starbases++
				nba++
			} else if ship.Class == "TR" {
				all_transport_tons += ship.Tonnage
				n_transports++
				ntr++
			} else {
				if ship.Type == "SUB_LIGHT" {
					total_defensive_power += power(ship.Tonnage)
				} else {
					total_offensive_power += power(ship.Tonnage)
				}
				all_warship_tons += ship.Tonnage
				n_warships++
				nwa++
			}
		}
		fmt.Printf("%5d", num_ships)
		fmt.Printf("%5d", num_yards)

		if nwa < min_warships {
			min_warships = nwa
		}
		if nwa > max_warships {
			max_warships = nwa
		}

		if nba < min_starbases {
			min_starbases = nba
		}
		if nba > max_starbases {
			max_starbases = nba
		}

		if ntr < min_transports {
			min_transports = ntr
		}
		if ntr > max_transports {
			max_transports = ntr
		}

		if species.Tech.Military.Level == 0 {
			total_defensive_power = 0
			total_offensive_power = 0
		} else {
			total_offensive_power = (total_offensive_power + (species.Tech.Military.Level*total_offensive_power)/50) / 10
			total_defensive_power = (total_defensive_power + (species.Tech.Military.Level*total_defensive_power)/50) / 10
		}

		fmt.Printf("%8d%8d", total_offensive_power, total_defensive_power)

		totalBankedEconUnits += species.BankedEconUnits
		if species_number == 1 {
			minBankedEconUnits, maxBankedEconUnits = species.BankedEconUnits, species.BankedEconUnits
		} else {
			if minBankedEconUnits > species.BankedEconUnits {
				minBankedEconUnits = species.BankedEconUnits
			}
			if maxBankedEconUnits < species.BankedEconUnits {
				maxBankedEconUnits = species.BankedEconUnits
			}
		}
		fmt.Printf("%9d\n", species.BankedEconUnits)
	}

	fmt.Printf("\n")
	m := n_species / 2
	var avg_tech_level int
	for i := 0; i < 6; i++ {
		avg_tech_level = (all_tech_level[i] + m) / n_species
		var tech_name string
		switch i {
		case 0:
			tech_name = "Mining"
		case 1:
			tech_name = "Manufacturing"
		case 2:
			tech_name = "Military"
		case 3:
			tech_name = "Gravitics"
		case 4:
			tech_name = "Life Support"
		case 5:
			tech_name = "Biology"
		}
		fmt.Printf("Average %s tech level = %d (min = %d, max = %d)\n", tech_name, avg_tech_level, min_tech_level[i], max_tech_level[i])
	}

	i := ((10 * n_warships) + m) / n_species
	fmt.Printf("\nAverage number of warships per species = %d.%d (min = %d, max = %d)\n", i/10, i%10, min_warships, max_warships)

	if n_warships == 0 {
		n_warships = 1
	}
	avg_warship_tons := (10000 * all_warship_tons) / n_warships
	avg_warship_tons = 1000 * ((avg_warship_tons + 500) / 1000)
	fmt.Printf("Average warship size = %s tons\n", commas(avg_warship_tons))

	avg_warship_tons = (10000 * all_warship_tons) / n_species
	avg_warship_tons = 1000 * ((avg_warship_tons + 500) / 1000)
	fmt.Printf("Average total warship tonnage per species = %s tons\n", commas(avg_warship_tons))

	i = ((10 * n_starbases) + m) / n_species
	fmt.Printf("\nAverage number of starbases per species = %d.%d (min = %d, max = %d)\n", i/10, i%10, min_starbases, max_starbases)

	if n_starbases == 0 {
		n_starbases = 1
	}
	avg_starbase_tons := (10000 * all_starbase_tons) / n_starbases
	avg_starbase_tons = 1000 * ((avg_starbase_tons + 500) / 1000)
	fmt.Printf("Average starbase size = %s tons\n", commas(avg_starbase_tons))

	avg_starbase_tons = (10000 * all_starbase_tons) / n_species
	avg_starbase_tons = 1000 * ((avg_starbase_tons + 500) / 1000)
	fmt.Printf("Average total starbase tonnage per species = %s tons\n", commas(avg_starbase_tons))

	i = ((10 * n_transports) + m) / n_species
	fmt.Printf("\nAverage number of transports per species = %d.%d (min = %d, max = %d)\n", i/10, i%10, min_transports, max_transports)

	if n_transports == 0 {
		n_transports = 1
	}
	avg_transport_tons := (10000 * all_transport_tons) / n_transports
	avg_transport_tons = 1000 * ((avg_transport_tons + 500) / 1000)
	fmt.Printf("Average transport size = %s tons\n", commas(avg_transport_tons))

	avg_transport_tons = (10000 * all_transport_tons) / n_species
	avg_transport_tons = 1000 * ((avg_transport_tons + 500) / 1000)
	fmt.Printf("Average total transport tonnage per species = %s tons\n", commas(avg_transport_tons))

	avg_yards := ((10 * n_yards) + m) / n_species
	fmt.Printf("\nAverage number of shipyards per species = %d.%d (min = %d, max = %d)\n", avg_yards/10, avg_yards%10, min_yards, max_yards)

	avg_pop_pl := ((10 * n_pop_pl) + m) / n_species
	fmt.Printf("\nAverage number of populated planets per species = %d.%d (min = %d, max = %d)\n", avg_pop_pl/10, avg_pop_pl%10, min_pop_pl, max_pop_pl)

	avg_production := (all_production + m) / n_species
	fmt.Printf("Average total production per species = %d (min = %d, max = %d)\n", avg_production, min_production, max_production)

	avgBankedEconUnits := (totalBankedEconUnits + m) / n_species
	fmt.Printf("\nAverage banked economic units per species = %d (min = %d, max = %d)\n", avgBankedEconUnits, minBankedEconUnits, maxBankedEconUnits)
}

func commas(i int) string {
	return fmt.Sprintf("%d", i)
}

func LifeSupportNeeded(species *jsondb.Species, home, colony *jsondb.Planet) int {
	// temperature class
	tcDelta := colony.TemperatureClass - home.TemperatureClass
	if tcDelta < 0 {
		tcDelta *= -1
	}

	// pressure class
	pcDelta := colony.PressureClass - home.PressureClass
	if pcDelta < 0 {
		pcDelta *= -1
	}

	// check gases on planet
	atmoDelta := 3 // assumes that required gases are not present
	for gas, pct := range colony.Gases {
		// if poisonous to species, bump needed life support
		if species.Gases.Poison[gas] {
			atmoDelta += 3
		}
		// if required gas is present and in the right range, decrease needed life support
		if amt, ok := species.Gases.Required[gas]; ok {
			if amt.Min <= pct && pct <= amt.Max {
				atmoDelta -= 3
			}
		}
	}

	return tcDelta * 3 + pcDelta * 3 + atmoDelta
}
