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

/* Generate order section. */
func GenerateOrders(w io.Writer, g *GalaxyData, s *SpeciesData, ignore_field_distorters, truncate_name bool) bool {
	temp_ignore_field_distorters := ignore_field_distorters
	ignore_field_distorters = true

	fmt.Fprintf(w, "\n\n* * * * * * * * * * * * * * * * * * * * * * * * *\n")
	fmt.Fprintf(w, "\n\nORDER SECTION. Remove these two lines and everything above\n")
	fmt.Fprintf(w, "  them, and submit only the orders below.\n\n")

	GenerateCombatOrders(w, s)
	GeneratePreDepartureOrders(w, s)
	GenerateJumpOrders(w, s, g.AllStars(), ignore_field_distorters, truncate_name)
	GenerateProductionOrders(w, s, ignore_field_distorters, truncate_name)
	GeneratePostArrivalOrders(w, s)
	GenerateStrikeOrders(w, s)

	truncate_name = false
	ignore_field_distorters = temp_ignore_field_distorters
	return ignore_field_distorters
}

/* Generate JUMP orders for all ships that have not yet been given orders. */
func GenerateAutoJumpOrders(w io.Writer, s *SpeciesData, stars []*StarData, ignore_field_distorters, truncate_name bool) {
	for _, ship := range s.Ships {
		// TODO: what is so special about orbit 99?
		if ship.Coords.Orbit == 99 || ship.JustJumped != DidNotJump || ship.Status.UnderConstruction || ship.Status.JumpedInCombat || ship.Status.ForcedJump {
			continue
		}

		if ship.Type == FTL {
			fmt.Fprintf(w, "\tJump\t%s, ", ship.GetName(ignore_field_distorters, truncate_name))
			if ship.Class == TR && ship.Tonnage == 1 {
				closestUnvisitedSystem := s.ClosestUnvisitedSystem(ship, stars)
				if closestUnvisitedSystem == nil {
					fmt.Fprintf(w, "???")
				} else {
					fmt.Fprintf(w, "%d %d %d", closestUnvisitedSystem.Coords.X, closestUnvisitedSystem.Coords.Y, closestUnvisitedSystem.Coords.Z)
					/* So that we don't send more than one ship to the same place. */
					closestUnvisitedSystem.VisitedBy[s.Name] = true
				}
				fmt.Fprintf(w, "\n\t\t\t; Age %d, now at %d %d %d, ", ship.Age, ship.Coords.X, ship.Coords.Y, ship.Coords.Z)
				if ship.Status.InOrbit {
					fmt.Fprintf(w, "O%d, ", ship.Coords.Orbit)
				} else if ship.Status.OnSurface {
					fmt.Fprintf(w, "L%d, ", ship.Coords.Orbit)
				} else {
					fmt.Fprintf(w, "D, ")
				}
				s.ReportMishapChance(w, ship, closestUnvisitedSystem.Coords)
				ship.Dest = closestUnvisitedSystem.Coords
			} else {
				fmt.Fprintf(w, "???\t; Age %d, now at %d %d %d", ship.Age, ship.Coords.X, ship.Coords.Y, ship.Coords.Z)
				if ship.Status.InOrbit {
					fmt.Fprintf(w, ", O%d", ship.Coords.Orbit)
				} else if ship.Status.OnSurface {
					fmt.Fprintf(w, ", L%d", ship.Coords.Orbit)
				} else {
					fmt.Fprintf(w, ", D")
				}
				/* Save destination so that we can check later if it needs to be scanned. */
				ship.Dest = Coords{X: -1, Y: -1, Z: -1} // TODO: this is supposed to signal something?
			}
			fmt.Fprintf(w, "\n")
		}
	}
}

func GenerateCombatOrders(w io.Writer, s *SpeciesData) {
	fmt.Fprintf(w, "START COMBAT\n")
	fmt.Fprintf(w, "; Place combat orders here.\n\n")
	fmt.Fprintf(w, "END\n\n")
}

/* Generate auto-jumps for ships that were loaded via the DEVELOP command or which were UNLOADed because of the AUTO command. */
func GenerateJumpOrders(w io.Writer, s *SpeciesData, stars []*StarData, ignore_field_distorters, truncate_name bool) {
	fmt.Fprintf(w, "START JUMPS\n")
	fmt.Fprintf(w, "; Place jump orders here.\n\n")

	for _, ship := range s.Ships {
		ship.JustJumped = DidNotJump
		if ship.Coords.Orbit == 99 {
			continue
		}
		if ship.Status.JumpedInCombat {
			continue
		}
		if ship.Status.ForcedJump {
			continue
		}

		if ship.Special.AutoJumpTarget.IsSet() {
			// TODO: removed the special logic for 9999 == HomePlanet
			target := s.GetNamedPlanetAt(ship.Special.AutoJumpTarget)
			fmt.Fprintf(w, "\tJump\t%s, PL %s\t; Age %d, ", ship.GetName(ignore_field_distorters, truncate_name), target.Name, ship.Age)
			s.ReportMishapChance(w, ship, target.Planet.Coords)
			fmt.Fprintf(w, "\n\n")
			ship.JustJumped = JustJumped
			continue
		}

		fmt.Printf("TODO: was n := ship.UnloadingPoint\n")
		if ship.UnloadingPoint.IsSet() {
			// TODO: removed the special logic for 9999 == HomePlanet
			target := s.GetNamedPlanetAt(ship.UnloadingPoint)
			fmt.Fprintf(w, "\tJump\t%s, PL %s\t; ", ship.GetName(ignore_field_distorters, truncate_name), target.Name)
			s.ReportMishapChance(w, ship, target.Planet.Coords)
			fmt.Fprintf(w, "\n\n")
			ship.JustJumped = JustJumped
		}
	}

	if s.AutoOrders {
		GenerateAutoJumpOrders(w, s, stars, ignore_field_distorters, truncate_name)
	}

	fmt.Fprintf(w, "END\n\n")
}

func GeneratePostArrivalOrders(w io.Writer, s *SpeciesData) {
	fmt.Fprintf(w, "START POST-ARRIVAL\n")
	fmt.Fprintf(w, "; Place post-arrival orders here.\n\n")
	GenerateScanOrders(w, s)
	fmt.Fprintf(w, "END\n\n")
}

func GeneratePreDepartureOrders(w io.Writer, s *SpeciesData) {
	fmt.Fprintf(w, "START PRE-DEPARTURE\n")
	fmt.Fprintf(w, "; Place pre-departure orders here.\n\n")
	for _, nampla := range s.NamedPlanets {
		if nampla.Planet.Coords.Orbit == 99 {
			continue
		}

		/* Generate auto-installs for colonies that were loaded via the DEVELOP command. */
		if nampla.AutoIUs == 0 && nampla.AutoAUs == 0 {
			fmt.Fprintf(w, "\n")
		} else {
			if nampla.AutoIUs != 0 {
				fmt.Fprintf(w, "\tInstall\t%d IU\tPL %s\n", nampla.AutoIUs, nampla.Name)
			}
			if nampla.AutoAUs != 0 {
				fmt.Fprintf(w, "\tInstall\t%d AU\tPL %s\n", nampla.AutoAUs, nampla.Name)
			}
		}
		if !s.AutoOrders {
			continue
		}

		/* Generate auto UNLOAD orders for transports at this nampla. */
		for _, ship := range s.Ships {
			if ship.Coords.Orbit == 99 {
				continue
			}
			if ship.Coords.X != nampla.Planet.Coords.X {
				continue
			}
			if ship.Coords.Y != nampla.Planet.Coords.Y {
				continue
			}
			if ship.Coords.Z != nampla.Planet.Coords.Z {
				continue
			}
			if ship.Coords.Orbit != nampla.Planet.Coords.Orbit {
				continue
			}
			if ship.Status.JumpedInCombat {
				continue
			}
			if ship.Status.ForcedJump {
				continue
			}
			if ship.Class != TR {
				continue
			}
			if ship.ItemQuantity[CU] < 1 {
				continue
			}

			/* New colonies will never be started automatically unless ship was loaded via a DEVELOP order. */
			if ship.LoadingPoint.IsSet() {
				/* Check if transport is at specified unloading point. */
				// TODO: is this right?
				if nampla.Planet.Coords.SamePlanet(ship.LoadingPoint) {
					goto unload_ship
				}
			}

			if !nampla.Status.Populated {
				continue
			} else if (nampla.MIBase + nampla.MABase) >= 2000 {
				continue
			} else if nampla.Planet.Coords.SameSystem(s.Home.World.Planet.Coords) {
				continue /* Home sector. */
			}

		unload_ship:

			if ship.LoadingPoint.SamePlanet(nampla.Planet.Coords) {
				// TODO: this is planet, not system, right?
				continue /* Ship was just loaded here. */
			}
			fmt.Fprintf(w, "\tUnload\tTR%d%s %s\n\n", ship.Tonnage, shipType[ship.Type], ship.Name)

			// TODO: is this right?
			ship.Special.LoadingPoint = ship.LoadingPoint
			// TODO: is this right?
			ship.UnloadingPoint = nampla.Planet.Coords
		}
	}

	fmt.Fprintf(w, "END\n\n")

}

/* Generate a PRODUCTION order for each planet that can produce. */
func GenerateProductionOrders(w io.Writer, s *SpeciesData, ignore_field_distorters, truncate_name bool) {
	fmt.Fprintf(w, "START PRODUCTION\n\n")
	fmt.Fprintf(w, ";   Economic units at start of turn = %d\n\n", s.EconUnits)
	// TODO: why do this in reverse order?
	for _, nampla := range s.NamedPlanetsReversed() {
		// TODO: what is so special about orbit 99?
		if nampla.Planet.Coords.Orbit == 99 {
			continue
		} else if nampla.MIBase == 0 && !nampla.Status.ResortColony {
			continue
		} else if nampla.MABase == 0 && !nampla.Status.MiningColony {
			continue
		}
		fmt.Fprintf(w, "    PRODUCTION PL %s\n", nampla.Name)
		if nampla.Status.MiningColony {
			fmt.Fprintf(w, "    ; The above PRODUCTION order is required for this mining colony, even\n")
			fmt.Fprintf(w, "    ;  if no other production orders are given for it. This mining colony\n")
			// TODO: is this really use_on_ambush?
			fmt.Fprintf(w, "    ;  will generate %d economic units this turn.\n", nampla.UseOnAmbush)
		} else if nampla.Status.ResortColony {
			fmt.Fprintf(w, "    ; The above PRODUCTION order is required for this resort colony, even\n")
			fmt.Fprintf(w, "    ;  though no other production orders can be given for it.  This resort\n")
			// TODO: is this really use_on_ambush?
			fmt.Fprintf(w, "    ;  colony will generate %d economic units this turn.\n", nampla.UseOnAmbush)
		} else {
			fmt.Fprintf(w, "    ; Place production orders here for planet %s (sector %d %d %d #%d).\n", nampla.Name, nampla.Planet.Coords.X, nampla.Planet.Coords.Y, nampla.Planet.Coords.Z, nampla.Planet.Coords.Orbit)
			fmt.Fprintf(w, "    ;  Avail pop = %d, shipyards = %d, to spend = %d", nampla.PopUnits, nampla.Shipyards, nampla.UseOnAmbush)
			n := nampla.UseOnAmbush
			if nampla.Status.HomePlanet {
				if s.HPOriginalBase != 0 {
					fmt.Fprintf(w, " (max = %d)", 5*n)
				} else {
					fmt.Fprintf(w, " (max = no limit)")
				}
			} else {
				fmt.Fprintf(w, " (max = %d)", 2*n)
			}
			fmt.Fprintf(w, ".\n\n")
		}

		/* Build IUs and AUs for incoming ships with CUs. */
		if nampla.IUsNeeded != 0 {
			fmt.Fprintf(w, "\tBuild\t%d IU\n", nampla.IUsNeeded)
		}
		if nampla.AUsNeeded != 0 {
			fmt.Fprintf(w, "\tBuild\t%d AU\n", nampla.AUsNeeded)
		}
		if nampla.IUsNeeded != 0 || nampla.AUsNeeded != 0 {
			fmt.Fprintf(w, "\n")
		}

		if !s.AutoOrders {
			continue
		} else if nampla.Status.MiningColony || nampla.Status.ResortColony {
			continue
		}

		/* See if there are any RMs to recycle. */
		n := nampla.Special / 5
		if n > 0 {
			fmt.Fprintf(w, "\tRecycle\t%d RM\n\n", 5*n)
		}

		/* Generate DEVELOP commands for ships arriving here because of AUTO command. */
		for _, ship := range s.Ships {
			if ship.Coords.Orbit == 99 {
				continue
			}
			k := ship.Special.AutoJumpTarget
			fmt.Println("TODO: should this be SamePlanet or SameSystem")
			if !k.IsSet() || !nampla.Planet.Coords.SamePlanet(k) {
				continue
			}
			planet := s.GetNamedPlanetAt(ship.UnloadingPoint)
			fmt.Fprintf(w, "\tDevelop\tPL %s, TR%d%s %s\n\n", planet.Name, ship.Tonnage, shipType[ship.Type], ship.Name)
		}

		/* Give orders to continue construction of unfinished ships and starbases. */
		for _, ship := range s.Ships {
			if ship.Coords.Orbit == 99 {
				continue
			}
			if ship.Coords.X != nampla.Planet.Coords.X {
				continue
			}
			if ship.Coords.Y != nampla.Planet.Coords.Y {
				continue
			}
			if ship.Coords.Z != nampla.Planet.Coords.Z {
				continue
			}
			if ship.Coords.Orbit != nampla.Planet.Coords.Orbit {
				continue
			}

			if ship.Status.UnderConstruction {
				fmt.Fprintf(w, "\tContinue\t%s, %d\t; Left to pay = %d\n\n", ship.GetName(ignore_field_distorters, truncate_name), ship.RemainingCost, ship.RemainingCost)
				continue
			}

			if ship.Type != STARBASE {
				continue
			}

			j := (s.TechLevel[MA] / 2) - ship.Tonnage
			if j < 1 {
				continue
			}

			fmt.Fprintf(w, "\tContinue\tBAS %s, %d\t; Current tonnage = %s\n\n", ship.Name, 100*j, Commas(10000*ship.Tonnage))
		}

		/* Generate DEVELOP command if this is a colony with an economic base less than 200. */
		n = nampla.MIBase + nampla.MABase + nampla.IUsNeeded + nampla.AUsNeeded
		nn := nampla.ItemQuantity[CU]
		for _, ship := range s.Ships {
			/* Get CUs on transports at planet. */
			if !ship.Coords.SamePlanet(nampla.Planet.Coords) {
				continue
			}
			nn += ship.ItemQuantity[CU]
		}
		n += nn
		if (nampla.Status.Colony) && n < 2000 && nampla.PopUnits > 0 {
			if nampla.PopUnits > (2000 - n) {
				nn = 2000 - n
			} else {
				nn = nampla.PopUnits
			}
			fmt.Fprintf(w, "\tDevelop\t%d\n\n", 2*nn)
			nampla.IUsNeeded += nn
		}

		// For home planets and any colonies that have an economic base of at least 200,
		// check if there are other colonized planets in the same sector that are not
		// self-sufficient. If so, DEVELOP them.
		if n >= 2000 || nampla.Status.HomePlanet {
			/* Skip home planet. */
			for _, temp_nampla := range s.NamedPlanets {
				if nampla == temp_nampla {
					continue
				}
				// TODO: what is so special about orbit 99
				if temp_nampla.Planet.Coords.Orbit == 99 || !nampla.Planet.Coords.SameSystem(temp_nampla.Planet.Coords) {
					continue
				}

				n = temp_nampla.MIBase + temp_nampla.MABase + temp_nampla.IUsNeeded + temp_nampla.AUsNeeded
				if n == 0 {
					continue
				}

				nn = temp_nampla.ItemQuantity[IU] + temp_nampla.ItemQuantity[AU]
				if nn > temp_nampla.ItemQuantity[CU] {
					nn = temp_nampla.ItemQuantity[CU]
				}
				n += nn
				if n >= 2000 {
					continue
				}
				nn = 2000 - n
				if nn > nampla.PopUnits {
					nn = nampla.PopUnits
				}
				fmt.Fprintf(w, "\tDevelop\t%d\tPL %s\n\n", 2*nn, temp_nampla.Name)
				temp_nampla.AUsNeeded += nn
			}
		}
	}

	fmt.Fprintf(w, "END\n\n")
}

func GenerateScanOrders(w io.Writer, s *SpeciesData) {
	if !s.AutoOrders {
		return
	}
	/* Generate an AUTO command. */
	fmt.Fprintf(w, "\tAuto\n\n")
	/* Generate SCAN orders for all TR1s that are jumping to sectors which current species does not inhabit. */
	for _, ship := range s.Ships {
		if ship.Coords.Orbit == 99 {
			continue
		}
		if ship.Status.UnderConstruction {
			continue
		}
		if ship.Class != TR {
			continue
		}
		if ship.Tonnage != 1 {
			continue
		}
		if ship.Type != FTL {
			continue
		}
		found := false
		for _, nampla := range s.NamedPlanets {
			if ship.Dest.X == -1 {
				break
			}
			if nampla.Planet.Coords.Orbit == 99 {
				continue
			}
			if nampla.Planet.Coords.X != ship.Dest.X {
				continue
			}
			if nampla.Planet.Coords.Y != ship.Dest.Y {
				continue
			}
			if nampla.Planet.Coords.Z != ship.Dest.Z {
				continue
			}
			if nampla.Status.Populated {
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(w, "\tScan\tTR1 %s\n", ship.Name)
		}
	}

}

func GenerateStrikeOrders(w io.Writer, s *SpeciesData) {
	fmt.Fprintf(w, "START STRIKES\n")
	fmt.Fprintf(w, "; Place strike orders here.\n\n")
	fmt.Fprintf(w, "END\n")
}
