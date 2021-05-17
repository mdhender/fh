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

package fh

import (
	"fmt"
	"io"
)

/* Status codes for named planets. These are logically ORed together. */
const HOME_PLANET = 1
const COLONY = 2
const POPULATED = 8
const MINING_COLONY = 16
const RESORT_COLONY = 32
const DISBANDED_COLONY = 64

type NamedPlanetData struct {
	ID           string            `json:"id"`
	Name         string            /* Name of planet. */
	Coords       Coords            `json:"coords"` // coordinates
	Status       NamedPlanetStatus `json:"status,omitempty"`
	Hiding       bool              `json:"hide_order_give,omitempty"` /* HIDE order given. */
	Hidden       bool              `json:"hidden,omitempty"`          /* Colony is hidden. */
	PlanetIndex  int               /* Index (starting at zero) into the file "planets.dat" of this planet. */
	SiegeEff     int               /* Siege effectiveness - a percentage between 0 and 99. */
	Shipyards    int               /* Number of shipyards on planet. */
	IUsNeeded    int               /* Incoming ship with only CUs on board. */
	AUsNeeded    int               /* Incoming ship with only CUs on board. */
	AutoIUs      int               /* Number of IUs to be automatically installed. */
	AutoAUs      int               /* Number of AUs to be automatically installed. */
	IUsToInstall int               /* Colonial mining units to be installed. */
	AUsToInstall int               /* Colonial manufacturing units to be installed. */
	MIBase       int               /* Mining base times 10. */
	MABase       int               /* Manufacturing base times 10. */
	PopUnits     int               /* Number of available population units. */
	UseOnAmbush  int               /* Amount to use on ambush. */
	Message      int               /* Message associated with this planet, if any. */
	Special      int               /* Different for each application. */
	ItemQuantity [MAX_ITEMS]int    /* Quantity of each item available. */
}

type NamedPlanetStatus struct {
	HomePlanet      bool `json:"home_planet,omitempty"`
	Colony          bool `json:"colony,omitempty"`
	Populated       bool `json:"populated,omitempty"`
	MiningColony    bool `json:"mining_colony,omitempty"`
	ResortColony    bool `json:"resort_colony,omitempty"`
	DisbandedColony bool `json:"disbanded_colony,omitempty"`
}

// CheckPopulation will set the Populated, MiningColony, and ResortColony flags.
// It will return true if the nampla is populated or false if not.
// It will also check if a message associated with this planet should be logged.
func (n *NamedPlanetData) CheckPopulation(l *Logger) bool {
	was_already_populated := n.Status.Populated
	total_pop := n.MIBase + n.MABase + n.IUsToInstall + n.AUsToInstall + n.ItemQuantity[PD] + n.ItemQuantity[CU] + n.PopUnits

	n.Status.Populated = total_pop > 0
	if total_pop == 0 {
		n.Status.MiningColony = false
		n.Status.ResortColony = false
	}

	if n.Status.Populated && !was_already_populated {
		if n.Message != 0 {
			// There is a message that must be logged whenever this planet becomes populated for the first time.
			filename := fmt.Sprintf("message%d.txt", n.Message)
			l.Message(filename)
		}
	}

	return n.Status.Populated
}

/* Print type of planet, name and coordinates. */
func (n *NamedPlanetData) Report(w io.Writer, s *SpeciesData, planet *PlanetData, ships []*ShipData) {
	fmt.Fprintf(w, "\n\n* * * * * * * * * * * * * * * * * * * * * * * * *\n")
	fmt.Fprintf(w, "\n\n")

	if n.Status.HomePlanet {
		fmt.Fprintf(w, "HOME PLANET")
	} else if n.Status.MiningColony {
		fmt.Fprintf(w, "MINING COLONY")
	} else if n.Status.ResortColony {
		fmt.Fprintf(w, "RESORT COLONY")
	} else if n.Status.Populated {
		fmt.Fprintf(w, "COLONY PLANET")
	} else {
		fmt.Fprintf(w, "PLANET")
	}

	fmt.Fprintf(w, ": PL %s", n.Name)
	fmt.Fprintf(w, "\n   Coordinates: x = %d, y = %d, z = %d, planet number %d\n", n.Coords.X, n.Coords.Y, n.Coords.Z, n.Coords.Orbit)

	if n.Status.HomePlanet && n.MIBase+n.MABase < s.HPOriginalBase {
		current_base := n.MIBase + n.MABase
		cuNeeded := s.HPOriginalBase - current_base /* Number of CUs needed. */
		md := s.Home.Planet.MiningDifficulty
		denom := 100 + md
		auNeeded := (100*(cuNeeded+n.MIBase) - (md * n.MABase) + denom/2) / denom
		iuNeeded := cuNeeded - auNeeded
		if iuNeeded < 0 {
			auNeeded, iuNeeded = cuNeeded, 0
		}
		if auNeeded < 0 {
			iuNeeded, auNeeded = cuNeeded, 0
		}

		fmt.Fprintf(w, "\nWARNING! Home planet has not yet completely recovered from bombardment!\n")
		fmt.Fprintf(w, "         %d IUs and %d AUs will have to be installed for complete recovery.\n", iuNeeded, auNeeded)
	}

	if n.Status.Populated {
		/* Print available population. */
		if !(n.Status.MiningColony || n.Status.ResortColony) {
			fmt.Fprintf(w, "\nAvailable population units = %d\n", n.PopUnits)
		}
		if n.SiegeEff != 0 {
			fmt.Fprintf(w, "\nWARNING!  This planet is currently under siege and will remain\n")
			fmt.Fprintf(w, "  under siege until the combat phase of the next turn!\n")
		}
		if n.UseOnAmbush > 0 {
			fmt.Fprintf(w, "\nIMPORTANT!  This planet has made preparations for an ambush!\n")
		}
		if n.Hidden {
			fmt.Fprintf(w, "\nIMPORTANT!  This planet is actively hiding from alien observation!\n")
		}

		/* Print what will be produced this turn. */
		raw_material_units := (10 * s.TechLevel[MI] * n.MIBase) / planet.MiningDifficulty
		production_capacity := (s.TechLevel[MA] * n.MABase) / 10

		ls_needed, production_penalty := s.LifeSupportNeeded(planet), 0
		if ls_needed != 0 {
			production_penalty = (100 * ls_needed) / s.TechLevel[LS]
		}
		fmt.Fprintf(w, "\nProduction penalty = %d%% (LSN = %d)\n", production_penalty, ls_needed)

		fmt.Fprintf(w, "\nEconomic efficiency = %d%%\n", planet.EconEfficiency)

		raw_material_units -= (production_penalty * raw_material_units) / 100
		raw_material_units = ((planet.EconEfficiency * raw_material_units) + 50) / 100

		production_capacity -= (production_penalty * production_capacity) / 100
		production_capacity = ((planet.EconEfficiency * production_capacity) + 50) / 100

		fleet_percent_cost := s.FleetPercentCost
		if fleet_percent_cost > 10000 {
			fleet_percent_cost = 10000
		}

		if n.MIBase > 0 {
			fmt.Fprintf(w, "\nMining base = %d.%d", n.MIBase/10, n.MIBase%10)
			fmt.Fprintf(w, " (MI = %d, MD = %d.%02d)\n", s.TechLevel[MI], planet.MiningDifficulty/100, planet.MiningDifficulty%100)

			/* For mining colonies, print economic units that will be produced. */
			if n.Status.MiningColony {
				n1 := (2 * raw_material_units) / 3
				n2 := ((fleet_percent_cost * n1) + 5000) / 10000
				n3 := n1 - n2
				fmt.Fprintf(w, "   This mining colony will generate %d - %d = %d economic units this turn.\n", n1, n2, n3)
				fmt.Printf("argh: MIBase hiding n3 in use_on_ambush slot! /* Temporary use only. */\n")
				n.UseOnAmbush = n3 /* Temporary use only. */
			} else {
				fmt.Fprintf(w, "   %d raw material units will be produced this turn.\n", raw_material_units)
			}
		}

		if n.MABase > 0 {
			if n.Status.ResortColony {
				fmt.Fprintf(w, "\n")
			}
			fmt.Fprintf(w, "Manufacturing base = %d.%d", n.MABase/10, n.MABase%10)
			fmt.Fprintf(w, " (MA = %d)\n", s.TechLevel[MA])

			/* For resort colonies, print economic units that will be produced. */
			if n.Status.ResortColony {
				n1 := (2 * production_capacity) / 3
				n2 := ((fleet_percent_cost * n1) + 5000) / 10000
				n3 := n1 - n2
				fmt.Fprintf(w, "   This resort colony will generate %d - %d = %d economic units this turn.\n", n1, n2, n3)
				fmt.Printf("argh: MABase hiding n3 in use_on_ambush slot! /* Temporary use only. */\n")
				n.UseOnAmbush = n3 /* Temporary use only. */
			} else {
				fmt.Fprintf(w, "   Production capacity this turn will be %d.\n", production_capacity)
			}
		}

		if n.ItemQuantity[RM] > 0 {
			fmt.Fprintf(w, "\n%ss (%s,C%d) carried over from last turn = %d\n", itemData[RM].name, itemData[RM].abbr, itemData[RM].carryCapacity, n.ItemQuantity[RM])
		}

		/* Print what can be spent this turn. */
		available_to_spend := 0
		raw_material_units += n.ItemQuantity[RM]
		if raw_material_units > production_capacity {
			/* Excess raw material units that may be recycled in AUTO mode. */
			available_to_spend = production_capacity
			n.Special = raw_material_units - production_capacity
		} else {
			available_to_spend = raw_material_units
			n.Special = 0
		}

		/* Don't print spendable amount for mining and resort colonies. */
		n1 := available_to_spend
		n2 := ((fleet_percent_cost * n1) + 5000) / 10000
		n3 := n1 - n2
		if !(n.Status.MiningColony || n.Status.ResortColony) {
			fmt.Fprintf(w, "\nTotal available for spending this turn = %d - %d = %d\n", n1, n2, n3)
			fmt.Printf("argh: totalAvailable hiding n3 in use_on_ambush slot! /* Temporary use only. */\n")
			n.UseOnAmbush = n3 /* Temporary use only. */
			fmt.Fprintf(w, "\nShipyard capacity = %d\n", n.Shipyards)
		}
	}

	header_printed := false
	for i := 0; i < MAX_ITEMS; i++ {
		if n.ItemQuantity[i] > 0 && i != RM {
			if !header_printed {
				fmt.Fprintf(w, "\nPlanetary inventory:\n")
				header_printed = true
			}
			fmt.Fprintf(w, "   %ss (%s,C%d) = %d", itemData[i].name, itemData[i].abbr, itemData[i].carryCapacity, n.ItemQuantity[i])
			if i == PD {
				fmt.Fprintf(w, " (warship equivalence = %d tons)", 50*n.ItemQuantity[PD])
			}
			fmt.Fprintf(w, "\n")
		}
	}

	/* Print all ships that are under construction on, on the surface of, or in orbit around this planet. */
	var shipList []*ShipData
	// Start with starbases
	for _, ship := range ships {
		if n.Coords.SamePlanet(ship.Coords) && ship.Class == BA {
			shipList = append(shipList, ship)
			ship.alreadyListed = true
		}
	}
	// then transports
	for _, ship := range ships {
		if n.Coords.SamePlanet(ship.Coords) && ship.Class == TR {
			shipList = append(shipList, ship)
			ship.alreadyListed = true
		}
	}
	// then everything else
	for _, ship := range ships {
		if n.Coords.SamePlanet(ship.Coords) && !ship.alreadyListed {
			shipList = append(shipList, ship)
			ship.alreadyListed = true
		}
	}
	// and now report on the "sorted" list of ships
	printing_alien := false
	if len(shipList) != 0 {
		fmt.Fprintf(w, "\nShips at PL %s:\n", n.Name)
		printHeader := true
		for _, ship := range shipList[1:] {
			ship.Report(w, printHeader, printing_alien, s)
			printHeader = false
		}
	}
}
