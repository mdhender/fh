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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type GameData struct {
	CurrentTurn int `json:"current_turn"`
}

func GetGame(galaxyPath string, verbose bool) (*GameData, error) {
	if galaxyPath == "" {
		return nil, fmt.Errorf("galaxy-path can't be empty")
	} else if galaxyPath != strings.TrimSpace(galaxyPath) {
		return nil, fmt.Errorf("galaxy-path can't have leading or trailing spaces")
	}
	file := filepath.Join(galaxyPath, "game.json")
	if verbose {
		fmt.Printf("[game] loading game data from %q\n", file)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var game GameData
	if err := json.Unmarshal(data, &game); err != nil {
		return nil, err
	}
	if game.CurrentTurn < 0 || game.CurrentTurn > 999999 {
		return nil, fmt.Errorf("current_turn must be between 0 and 999999")
	}
	return &game, nil
}

// Finish completes a turn
func (game *GameData) Finish(w io.Writer, galaxyPath string, test_mode, isVerbose bool) error {
	if isVerbose {
		_, _ = fmt.Fprintf(w, "\nFinishing up for all species...\n")
	}

	turnPath := filepath.Join(galaxyPath, game.TurnDir())
	fmt.Printf("[finish] %-30s == %q\n", "TURN_PATH", turnPath)
	outputPath := filepath.Join(galaxyPath, fmt.Sprintf("t%06d", game.CurrentTurn+1))
	fmt.Printf("[finish] %-30s == %q\n", "OUTPUT_PATH", outputPath)

	g, err := GetGalaxy(turnPath)
	if err != nil {
		return err
	}

	l := &Logger{Stdout: os.Stdout}

	var header_printed bool
	print_header := func() {
		l.String("\nOther events:\n")
		header_printed = true
	}

	transaction, err := GetTransactionData(galaxyPath)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(w, "Loaded %d transactions\n", len(transaction))

	// Total economic base includes all of the colonies on the planet, not just the one species.
	total_econ_base := make(map[string]int)

	// add mining difficulty increases for each planet, use the increase calcuated on the prior turn
	for _, planet := range g.AllPlanets() {
		planet.MiningDifficulty += planet.MDIncrease
		planet.MDIncrease = 0
	}

	/* Main loop. For each species, take appropriate action. */
	for _, species := range g.AllSpecies() {
		if isVerbose {
		}

		// check if player submitted orders for this turn.
		var ordersReceived bool
		if game.IsSetupTurn() {
			// pretend that we received orders to prevent the error message from being logged
			ordersReceived = true
		} else {
			orders, err := ioutil.ReadFile(filepath.Join(turnPath, fmt.Sprintf("sp%02d.ord", species.Number)))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			ordersReceived = err == nil && len(orders) != 0
		}
		if isVerbose {
			_, _ = fmt.Fprintf(w, "  Now doing SP %s...", species.Name)
			if !ordersReceived {
				_, _ = fmt.Fprintf(w, " WARNING: player did not submit orders this turn!")
			}
			_, _ = fmt.Fprintf(w, "\n")
		}

		// open log file
		var err error
		l.File, err = os.Create(filepath.Join(turnPath, fmt.Sprintf("sp%02d.log.txt", species.Number)))
		if err != nil {
			return err
		}
		l.Stdout = nil

		check := struct {
			mishaps            bool /* Check if any ships of this species experienced mishaps. */
			disbanded          bool /* Take care of any disbanded colonies. */
			transferInEU       bool /* Check if this species is the recipient of a transfer of economic units from another species. */
			jumpPortalsUsed    bool /* Check if any jump portals of this species were used by aliens. */
			detectedTelescopes bool /* Check if any starbases of this species detected the use of gravitic telescopes by aliens. */
			transferInTL       bool /* Check if this species is the recipient of a tech transfer from another species. */
			increaseTL         bool /* Calculate tech level increases. */
			transferInKN       bool /* Check if this species is the recipient of a knowledge transfer from another species. */
			loopNamedPlanets   bool /* Loop through each nampla for this species. */
			loopShips          bool /* Loop through all ships for this species. */
			alienIncursion     bool /* Check if this species has a populated planet that another species tried to land on. */
			alienConstruction  bool /* Check if this species is the recipient of interspecies construction. */
			besiegingOthers    bool /* Check if this species is besieging another species and detects forbidden construction, landings, etc. */
			messages           bool // check if this species is the recipient of a message from another species
		}{
			mishaps:            !game.IsSetupTurn(),
			disbanded:          !game.IsSetupTurn(),
			transferInEU:       !game.IsSetupTurn(),
			jumpPortalsUsed:    !game.IsSetupTurn(),
			detectedTelescopes: !game.IsSetupTurn(),
			transferInTL:       !game.IsSetupTurn(),
			increaseTL:         !game.IsSetupTurn(),
			transferInKN:       !game.IsSetupTurn(),
			loopNamedPlanets:   !game.IsSetupTurn(),
			loopShips:          !game.IsSetupTurn(),
			alienIncursion:     !game.IsSetupTurn(),
			alienConstruction:  !game.IsSetupTurn(),
			besiegingOthers:    !game.IsSetupTurn(),
			messages:           true, // always check messages
		}

		/* Check if any ships of this species experienced mishaps. */
		if check.mishaps {
			for _, t := range transaction {
				if t.Type == SHIP_MISHAP && t.Number1 == species.Number {
					if !header_printed {
						print_header()
					}
					l.String("  !!! ")
					l.String(t.Name1)
					if t.Value < 3 {
						/* Intercepted or self-destructed. */
						l.String(" disappeared without a trace, cause unknown!\n")
					} else if t.Value == 3 {
						/* Mis-jumped. */
						l.String(" mis-jumped to ")
						l.Int(t.X)
						l.Char(' ')
						l.Int(t.Y)
						l.Char(' ')
						l.Int(t.Z)
						l.String("!\n")
					} else {
						/* One fail-safe jump unit used. */
						l.String(" had a jump mishap! A fail-safe jump unit was expended.\n")
					}
				}
			}
		}

		/* Take care of any disbanded colonies. */
		if check.disbanded {
			var coloniesDestroyed, shipsDestroyed int
			for _, nampla := range species.NamedPlanets {
				if !nampla.Status.DisbandedColony {
					continue
				}

				/* Salvage ships on the surface and starbases in orbit. */
				salvage_EUs := 0
				for _, ship := range species.Ships {
					if !nampla.Planet.Coords.SameSystem(ship.Coords) {
						continue
					}
					if ship.Status.InOrbit && ship.Type != STARBASE {
						continue
					}

					/* Transfer cargo to planet. */
					for i := 0; i < MAX_ITEMS; i++ {
						nampla.ItemQuantity[i] += ship.ItemQuantity[i]
					}

					/* Salvage the ship. */
					original_cost := shipData[ship.Class].cost
					if ship.Class == TR || ship.Type == STARBASE {
						original_cost *= ship.Tonnage
					}

					if ship.Type == SUB_LIGHT {
						original_cost = (3 * original_cost) / 4
					}

					var salvage_value int
					if ship.Status.UnderConstruction {
						salvage_value = (original_cost - ship.RemainingCost) / 4
					} else {
						salvage_value = (3 * original_cost * (60 - ship.Age)) / 400
					}

					salvage_EUs += salvage_value

					/* Destroy the ship. */
					ship.Status.Destroyed = true
					shipsDestroyed++
				}

				/* Salvage items on the planet. */
				for i := 0; i < MAX_ITEMS; i++ {
					var salvage_value int
					if i == RM {
						salvage_value = nampla.ItemQuantity[RM] / 10
					} else if nampla.ItemQuantity[i] > 0 {
						original_cost := nampla.ItemQuantity[i] * itemData[i].cost
						if i == TP {
							if species.TechLevel[BI] > 0 {
								original_cost /= species.TechLevel[BI]
							} else {
								original_cost /= 100
							}
						}
						salvage_value = original_cost / 4
					} else {
						salvage_value = 0
					}

					salvage_EUs += salvage_value
				}

				/* Transfer EUs to species. */
				species.EconUnits += salvage_EUs

				/* Log what happened. */
				if !header_printed {
					print_header()
				}
				l.String("  PL ")
				l.String(nampla.Name)
				l.String(" was disbanded, generating ")
				l.Long(salvage_EUs)
				l.String(" economic units in salvage.\n")

				coloniesDestroyed++
			}

			// destroy the disbanded colonies
			if coloniesDestroyed != 0 {
				var namedPlanets []*NamedPlanetData
				for _, nampla := range species.NamedPlanets {
					if !nampla.Status.DisbandedColony {
						continue
					}
					namedPlanets = append(namedPlanets, nampla)
				}
				species.NamedPlanets = namedPlanets
			}

			// destroy the salvaged ships
			if shipsDestroyed != 0 {
				var ships []*ShipData
				for _, ship := range species.Ships {
					if !ship.Status.Destroyed {
						ships = append(ships, ship)
					}
				}
				species.Ships = ships
			}
		}

		/* Check if this species is the recipient of a transfer of economic units from another species. */
		if check.transferInEU {
			for _, t := range transaction {
				if t.Recipient == species.Number && (t.Type == EU_TRANSFER || t.Type == SIEGE_EU_TRANSFER || t.Type == LOOTING_EU_TRANSFER) {
					// Transfer EUs to attacker if this is a siege or looting transfer.
					// If this is a normal transfer, then just log the result since the actual transfer was done when the order was processed.
					if t.Type != EU_TRANSFER {
						species.EconUnits += t.Value
					}

					if !header_printed {
						print_header()
					}
					l.String("  ")
					l.Long(t.Value)
					l.String(" economic units were received from SP ")
					l.String(t.Name1)
					if t.Type == SIEGE_EU_TRANSFER {
						l.String(" as a result of your successful siege of their PL ")
						l.String(t.Name3)
						l.String(". The siege was ")
						l.Long(t.Number1)
						l.String("% effective")
					} else if t.Type == LOOTING_EU_TRANSFER {
						l.String(" as a result of your looting their PL ")
						l.String(t.Name3)
					}
					l.String(".\n")
				}
			}
		}

		/* Check if any jump portals of this species were used by aliens. */
		if check.jumpPortalsUsed {
			for _, t := range transaction {
				if t.Type == ALIEN_JUMP_PORTAL_USAGE && t.Number1 == species.Number {
					if !header_printed {
						print_header()
					}
					l.String("  ")
					l.String(t.Name1)
					l.Char(' ')
					l.String(t.Name2)
					l.String(" used jump portal ")
					l.String(t.Name3)
					l.String(".\n")
				}
			}
		}

		/* Check if any starbases of this species detected the use of gravitic telescopes by aliens. */
		if check.detectedTelescopes {
			for _, t := range transaction {
				if !(t.Type == TELESCOPE_DETECTION && t.Number1 == species.Number) {
					continue
				}
				if !header_printed {
					print_header()
				}
				l.String("! ")
				l.String(t.Name1)
				l.String(" detected the operation of an alien gravitic telescope at x = ")
				l.Int(t.X)
				l.String(", y = ")
				l.Int(t.Y)
				l.String(", z = ")
				l.Int(t.Z)
				l.String(".\n")
			}
		}

		/* Check if this species is the recipient of a tech transfer from another species. */
		if check.transferInTL {
			for _, t := range transaction {
				if !(t.Type == TECH_TRANSFER && t.Recipient == species.Number) {
					continue
				}

				/* Try to transfer technology. */
				//rec := t.Recipient - 1
				don := t.Donor - 1

				if !header_printed {
					print_header()
				}
				l.String("  ")
				tech := t.Value
				l.String(techData[tech].name)
				l.String(" tech transfer from SP ")
				l.String(t.Name1)
				their_level := t.Number3
				my_level := species.TechLevel[tech]

				if their_level <= my_level {
					l.String(" failed.\n")
					t.Number1 = -1
					continue
				}

				donor_species := g.GetSpeciesByNumber(don)
				actual_cost, max_cost := 0, t.Number1
				if max_cost == 0 {
					max_cost = donor_species.EconUnits
				} else if donor_species.EconUnits < max_cost {
					max_cost = donor_species.EconUnits
				}
				new_level := my_level
				for new_level < their_level {
					one_point_cost := new_level * new_level
					one_point_cost -= one_point_cost / 4 /* 25% discount. */
					if (actual_cost + one_point_cost) > max_cost {
						break
					}
					actual_cost += one_point_cost
					new_level++
				}

				if new_level == my_level {
					l.String(" failed due to lack of funding.\n")
					t.Number1 = -2
				} else {
					l.String(" raised your tech level from ")
					l.Int(my_level)
					l.String(" to ")
					l.Int(new_level)
					l.String(" at a cost to them of ")
					l.Long(actual_cost)
					l.String(".\n")
					t.Number1 = actual_cost
					t.Number2 = my_level
					t.Number3 = new_level

					species.TechLevel[tech] = new_level
					donor_species.EconUnits -= actual_cost
				}
			}
		}

		/* Calculate tech level increases. */
		if check.increaseTL {
			for tech := MI; tech <= BI; tech++ {
				old_tech_level := species.TechLevel[tech]
				new_tech_level := old_tech_level

				var max_tech_level int

				experience_points := species.TechEps[tech]
				if experience_points != 0 {
					/* Determine increase as if there were NO randomness in the process. */
					i := experience_points
					j := old_tech_level
					for i >= j*j {
						i -= j * j
						j++
					}

					// When extremely large amounts are spent on research, tech level increases are sometimes excessive.  Set a limit.
					if old_tech_level > 50 {
						max_tech_level = j + 1
					} else {
						max_tech_level = 9999
					}

					/* Allocate half of the calculated increase NON-RANDOMLY. */
					n := (j - old_tech_level) / 2
					for i = 0; i < n; i++ {
						experience_points -= new_tech_level * new_tech_level
						new_tech_level++
					}

					/* Allocate the rest randomly. */
					for experience_points >= new_tech_level {
						experience_points -= new_tech_level
						n = new_tech_level

						/* The chance of success is 1 in n. At this point, n is always at least 1. */
						i = rnd(16 * n)
						if i >= 8*n && i <= 8*n+15 {
							new_tech_level = n + 1
						}
					}

					/* Save unused experience points. */
					species.TechEps[tech] = experience_points
				}

				/* See if any random increase occurred. Odds are 1 in 6. */
				if old_tech_level > 0 && rnd(6) == 6 {
					new_tech_level++
				}

				if new_tech_level > max_tech_level {
					new_tech_level = max_tech_level
				}

				/* Report result only if tech level went up. */
				if new_tech_level > old_tech_level {
					if !header_printed {
						print_header()
					}
					l.String("  ")
					l.String(techData[tech].name)
					l.String(" tech level rose from ")
					l.Int(old_tech_level)
					l.String(" to ")
					l.Int(new_tech_level)
					l.String(".\n")

					species.TechLevel[tech] = new_tech_level
				}
			}
		}

		/* Notify of any new high tech items. */
		for tech := MI; tech <= BI; tech++ {
			old_tech_level := species.InitTechLevel[tech]
			new_tech_level := species.TechLevel[tech]

			if new_tech_level != old_tech_level {
				fmt.Printf("[debug] species %q old_tech_level %2d new_tech_level %2d\n", old_tech_level, new_tech_level)
				if new_tech_level > old_tech_level {
					check_high_tech_items(tech, old_tech_level, new_tech_level, l)
				}
			}

			species.InitTechLevel[tech] = new_tech_level
		}

		/* Check if this species is the recipient of a knowledge transfer from another species. */
		if check.transferInKN {
			for _, t := range transaction {
				if t.Type == KNOWLEDGE_TRANSFER && t.Recipient == species.Number {
					//rec := t.Recipient - 1
					//don := t.Donor - 1

					/* Try to transfer technology. */
					tech := t.Value
					their_level := t.Number3
					my_level := species.TechLevel[tech]
					n := species.TechKnowledge[tech]
					if n > my_level {
						my_level = n
					}

					if their_level <= my_level {
						continue
					}

					species.TechKnowledge[tech] = their_level

					if !header_printed {
						print_header()
					}
					l.String("  SP ")
					l.String(t.Name1)
					l.String(" transferred knowledge of ")
					l.String(techData[tech].name)
					l.String(" to you up to tech level ")
					l.Long(their_level)
					l.String(".\n")
				}
			}
		}

		/* Loop through each nampla for this species. */
		if check.loopNamedPlanets {
			for _, nampla := range species.NamedPlanets {
				if nampla.Planet.Coords.Orbit == 99 {
					continue
				}

				/* Get planet pointer. */
				planet := g.GetPlanet(nampla.Planet.Coords)
				if planet == nil {
					panic("assert(planet != nil)")
				}

				/* Clear any amount spent on ambush. */
				nampla.UseOnAmbush = 0

				/* Handle HIDE order. */
				nampla.Hidden = nampla.Hiding
				nampla.Hiding = false

				/* Check if any IUs or AUs were installed. */
				if nampla.IUsToInstall > 0 {
					nampla.MIBase += nampla.IUsToInstall
					nampla.IUsToInstall = 0
				}

				if nampla.AUsToInstall > 0 {
					nampla.MABase += nampla.AUsToInstall
					nampla.AUsToInstall = 0
				}

				/* Check if another species on the same planet has become
				 *  assimilated. */
				for _, t := range transaction {
					if !(t.Type == ASSIMILATION && t.Value == species.Number && nampla.Planet.Coords.SamePlanet(Coords{t.X, t.Y, t.Z, t.PN})) {
						continue
					}
					if !header_printed {
						print_header()
					}

					ib, ab, ns := t.Number1, t.Number2, t.Number3
					l.String("  Assimilation of ")
					l.String(t.Name1)
					l.String(" PL ")
					l.String(t.Name2)
					l.String(" increased mining base of ")
					l.String(species.Name)
					l.String(" PL ")
					l.String(nampla.Name)
					l.String(" by ")
					l.Long(ib / 10)
					l.Char('.')
					l.Long(ib % 10)
					l.String(", and manufacturing base by ")
					l.Long(ab / 10)
					l.Char('.')
					l.Long(ab % 10)
					if ns > 0 {
						l.String(". Number of shipyards was also increased by ")
						l.Int(ns)
					}
					l.String(".\n")
				}

				/* Calculate available population for this turn. */
				nampla.PopUnits = 0

				eb := nampla.MIBase + nampla.MABase
				total_pop_units := eb + nampla.ItemQuantity[CU] + nampla.ItemQuantity[PD]

				if nampla.Status.HomePlanet {
					if nampla.Status.Populated {
						nampla.PopUnits = HP_AVAILABLE_POP

						if species.HPOriginalBase != 0 { /* HP was bombed. */
							if eb >= species.HPOriginalBase {
								species.HPOriginalBase = 0 /* Fully recovered. */
							} else {
								nampla.PopUnits = (eb * HP_AVAILABLE_POP) / species.HPOriginalBase
							}
						}
					}
				} else if nampla.Status.Populated {
					/* Get life support tech level needed. */
					ls_needed := species.LifeSupportNeeded(planet)

					/* Basic percent increase is 10*(1 - ls_needed/ls_actual). */
					ls_actual := species.TechLevel[LS]
					percent_increase := 10 * (100 - ((100 * ls_needed) / ls_actual))

					if percent_increase < 0 { /* Colony wiped out! */
						if !header_printed {
							print_header()
						}

						l.String("  !!! Life support tech level was too low to support colony on PL ")
						l.String(nampla.Name)
						l.String(". Colony was destroyed.\n")

						/* No longer populated or self-sufficient. */
						nampla.Status = NamedPlanetStatus{Colony: true}
						nampla.MIBase = 0
						nampla.MABase = 0
						nampla.PopUnits = 0
						nampla.ItemQuantity[PD] = 0
						nampla.ItemQuantity[CU] = 0
						nampla.SiegeEff = 0
					} else {
						percent_increase /= 100

						/* Add a small random variation. */
						percent_increase +=
							rnd(percent_increase/4) - rnd(percent_increase/4)

						/* Add bonus for Biology technology. */
						percent_increase += species.TechLevel[BI] / 20

						/* Calculate and apply the change. */
						change := (percent_increase * total_pop_units) / 100

						if nampla.MIBase > 0 && nampla.MABase == 0 {
							nampla.Status.MiningColony = true
							change = 0
						} else if nampla.Status.MiningColony {
							/* A former mining colony has been converted to a normal colony. */
							nampla.Status.MiningColony = false
							change = 0
						}

						if nampla.MABase > 0 && nampla.MIBase == 0 && ls_needed <= 6 && planet.Gravity <= species.Home.World.Planet.Gravity {
							nampla.Status.ResortColony = true
							change = 0
						} else if nampla.Status.ResortColony {
							/* A former resort colony has been converted to a normal colony. */
							nampla.Status.ResortColony = false
							change = 0
						}

						if total_pop_units == nampla.ItemQuantity[PD] {
							change = 0 /* Probably an invasion force. */
						}
						nampla.PopUnits = change
					}
				}

				/* Handle losses due to attrition and update location array if planet is still populated. */
				// the for loop is a hack to remove one goto statement
				for nampla.Status.Populated {
					total_pop_units = nampla.PopUnits + nampla.MIBase + nampla.MABase + nampla.ItemQuantity[CU] + nampla.ItemQuantity[PD]

					if total_pop_units > 0 && total_pop_units < 50 {
						if nampla.PopUnits > 0 {
							nampla.PopUnits--
							break
						} else if nampla.ItemQuantity[CU] > 0 {
							nampla.ItemQuantity[CU]--
							if !header_printed {
								print_header()
							}
							l.String("  Number of colonist units on PL ")
							l.String(nampla.Name)
							l.String(" was reduced by one unit due to normal attrition.")
						} else if nampla.ItemQuantity[PD] > 0 {
							nampla.ItemQuantity[PD]--
							if !header_printed {
								print_header()
							}
							l.String("  Number of planetary defense units on PL ")
							l.String(nampla.Name)
							l.String(" was reduced by one unit due to normal attrition.")
						} else if nampla.MABase > 0 {
							nampla.MABase--
							if !header_printed {
								print_header()
							}
							l.String("  Manufacturing base of PL ")
							l.String(nampla.Name)
							l.String(" was reduced by 0.1 due to normal attrition.")
						} else {
							nampla.MIBase--
							if !header_printed {
								print_header()
							}
							l.String("  Mining base of PL ")
							l.String(nampla.Name)
							l.String(" was reduced by 0.1 due to normal attrition.")
						}

						if total_pop_units == 1 {
							if !header_printed {
								print_header()
							}
							l.String(" The colony is dead!")
						}

						l.Char('\n')
					}
					// again, the for loop was a hack to remove a goto statement, so we never really want to loop
					break
				}

				/* Apply automatic 2% increase to mining and manufacturing bases of home planets. */
				if nampla.Status.HomePlanet {
					growth_factor := 20
					ib := nampla.MIBase
					ab := nampla.MABase
					old_base := ib + ab
					increment := (growth_factor * old_base) / 1000
					md := planet.MiningDifficulty

					denom := 100 + md
					ab_increment := (100*(increment+ib) - (md * ab) + denom/2) / denom
					ib_increment := increment - ab_increment

					if ib_increment < 0 {
						ab_increment = increment
						ib_increment = 0
					}
					if ab_increment < 0 {
						ib_increment = increment
						ab_increment = 0
					}
					nampla.MIBase += ib_increment
					nampla.MABase += ab_increment
				}

				nampla.CheckPopulation(l)

				/* Update total economic base for colonies. */
				if !nampla.Status.HomePlanet {
					total_econ_base[nampla.Planet.Coords.String()] = total_econ_base[nampla.Planet.Coords.String()] + nampla.MIBase + nampla.MABase
				}
			}
		}

		/* Loop through all ships for this species. */
		if check.loopShips {
			for _, ship := range species.Ships {
				if ship.Coords.Orbit == 99 {
					continue
				}

				/* Set flag if ship arrived via a natural wormhole. */
				ship.ArrivedViaWormhole = ship.JustJumped == JumpedViaWormhole

				/* Clear 'just-jumped' flag. */
				ship.JustJumped = DidNotJump

				/* Increase age of ship. */
				if ship.Status.UnderConstruction {
					ship.Age++
					if ship.Age > 49 {
						ship.Age = 49
					}
				}
			}
		}

		/* Check if this species has a populated planet that another species tried to land on. */
		if check.alienIncursion {
			for _, t := range transaction {
				if !(t.Type == LANDING_REQUEST && t.Number1 == species.Number) {
					continue
				}
				if !header_printed {
					print_header()
				}
				l.String("  ")
				l.String(t.Name2)
				l.String(" owned by SP ")
				l.String(t.Name3)
				if t.Value != 0 {
					l.String(" was granted")
				} else {
					l.String(" was denied")
				}
				l.String(" permission to land on PL ")
				l.String(t.Name1)
				l.String(".\n")
			}
		}

		/* Check if this species is the recipient of interspecies construction. */
		if check.alienConstruction {
			for _, t := range transaction {
				if !(t.Type == INTERSPECIES_CONSTRUCTION && t.Recipient == species.Number) {
					continue
				}
				/* Simply log the result. */
				if !header_printed {
					print_header()
				}
				l.String("  ")
				if t.Value == 1 {
					l.Long(t.Number1)
					l.Char(' ')
					l.String(itemData[t.Number2].name)
					if t.Number1 == 1 {
						l.String(" was")
					} else {
						l.String("s were")
					}
					l.String(" constructed for you by SP ")
					l.String(t.Name1)
					l.String(" on PL ")
					l.String(t.Name2)
				} else {
					l.String(t.Name2)
					l.String(" was constructed for you by SP ")
					l.String(t.Name1)
				}
				l.String(".\n")
			}
		}

		/* Check if this species is besieging another species and detects forbidden construction, landings, etc. */
		if check.besiegingOthers {
			for _, t := range transaction {
				if !(t.Type == DETECTION_DURING_SIEGE && t.Number3 == species.Number) {
					continue
				}
				/* Log what was detected and/or destroyed. */
				if !header_printed {
					print_header()
				}
				l.String("  ")
				l.String("During the siege of ")
				l.String(t.Name3)
				l.String(" PL ")
				l.String(t.Name1)
				l.String(", your forces detected the ")
				if t.Value == 1 {
					/* Landing of enemy ship. */
					l.String("landing of ")
					l.String(t.Name2)
					l.String(" on the planet.\n")
				} else if t.Value == 2 {
					/* Enemy ship or starbase construction. */
					l.String("construction of ")
					l.String(t.Name2)
					l.String(", but you destroyed it before it")
					l.String(" could be completed.\n")
				} else if t.Value == 3 {
					/* Enemy PD construction. */
					l.String("construction of planetary defenses, but you")
					l.String(" destroyed them before they could be completed.\n")
				} else if t.Value == 4 || t.Value == 5 {
					/* Enemy item construction. */
					l.String("transfer of ")
					l.Int(t.Number1)
					l.Char(' ')
					l.String(itemData[t.Number2].name)
					if t.Number1 > 1 {
						l.Char('s')
					}
					if t.Value == 4 {
						l.String(" to PL ")
					} else {
						l.String(" from PL ")
					}
					l.String(t.Name2)
					l.String(", but you destroyed them in transit.\n")
				} else {
					panic("\n\tInternal error!  Cannot reach this point!\n\n")
				}
			}
		}

		// check if this species is the recipient of a message from another species
		if check.messages {
			for _, t := range transaction {
				if t.Type == MESSAGE_TO_SPECIES && t.Number2 == species.Number {
					if !header_printed {
						print_header()
					}
					fmt.Printf("SP %d received the following message from SP %s:\n\n", species.Number, t.Name1)
					l.String(fmt.Sprintf("\n  You received the following message from SP %s:\n\n", t.Name1))
					msg, err := GetMessage(galaxyPath, t.Value)
					if err == nil && l.File != nil {
						l.Message(msg)
					}
					l.String("\n  *** End of Message ***\n\n")
				}
			}
		}
	}

	// S10.9 - calculate economic efficiency for each planet
	for _, planet := range g.AllPlanets() {
		excess := total_econ_base[planet.Coords.String()] - 2000
		if excess <= 0 {
			planet.EconEfficiency = 100
			continue
		}
		planet.EconEfficiency = (100 * (excess/20 + 2000)) / total_econ_base[planet.Coords.String()]
	}

	/* Create new locations array. */
	locations := DoLocations(g)

	/* Go through all species one more time to update alien contact masks, report tech transfer results to donors, and calculate fleet maintenance costs. */
	if !game.IsSetupTurn() {
		if isVerbose {
			_, _ = fmt.Fprintf(w, "\nNow updating contact masks et al.\n")
		}
		for _, species := range g.AllSpecies() {
			/* Update contact mask in species data if this species has met a new alien. */
			for _, loc := range locations {
				if loc.S != species.Number {
					continue
				}

				for _, aloc := range locations {
					alienSpeciesNumber := aloc.S
					if species.Contact[alienSpeciesNumber] || species.Number == alienSpeciesNumber {
						continue // already made contact
					} else if !(aloc.X == loc.X && aloc.Y == loc.Y && aloc.Z == loc.Z) {
						continue
					}
					// we are in contact with an alien if it is visible
					alienSpecies := g.GetSpeciesByNumber(alienSpeciesNumber)
					species.Contact[alienSpeciesNumber] = g.AlienIsVisible(species, alienSpecies, Coords{X: loc.X, Y: loc.Y, Z: loc.Z})
				}
			}

			/* Report results of tech transfers to donor species. */
			for _, t := range transaction {
				if t.Type == TECH_TRANSFER && t.Donor == species.Number {
					continue
				}
				/* Open log file for appending. */
				filename := filepath.Join(turnPath, fmt.Sprintf("sp%02d.log.txt", species.Number))
				fd, err := os.OpenFile(filename, os.O_APPEND, 0600)
				if err != nil {
					fmt.Printf("%+v\n", err)
					panic(fmt.Sprintf("\n\tCannot open '%s' for appending!\n\n", filename))

				}
				l := &Logger{
					File: fd,
				}

				l.String("  ")
				l.String(techData[t.Value].name)
				l.String(" tech transfer to SP ")
				l.String(t.Name2)

				if t.Number1 < 0 {
					l.String(" failed")
					if t.Number1 == -2 {
						l.String(" due to lack of funding")
					}
				} else {
					l.String(" raised their tech level from ")
					l.Long(t.Number2)
					l.String(" to ")
					l.Long(t.Number3)
					l.String(" at a cost to you of ")
					l.Long(t.Number1)
				}

				l.String(".\n")
				l = nil // wish i could flush and close
			}

			/* Calculate fleet maintenance cost and its percentage of total production. */
			fleet_maintenance_cost := 0
			for _, ship := range species.Ships {
				if ship.Coords.Orbit == 99 {
					continue
				}

				var n int
				if ship.Class == TR {
					n = 4 * ship.Tonnage
				} else if ship.Class == BA {
					n = 10 * ship.Tonnage
				} else {
					n = 20 * ship.Tonnage
				}

				if ship.Type == SUB_LIGHT {
					n -= (25 * n) / 100
				}

				fleet_maintenance_cost += n
			}

			/* Subtract military discount. */
			i := species.TechLevel[ML] / 2
			fleet_maintenance_cost -= (i * fleet_maintenance_cost) / 100

			/* Calculate total production. */
			total_species_production := 0
			for _, nampla := range species.NamedPlanets {

				if nampla.Planet.Coords.Orbit == 99 {
					continue
				}
				if nampla.Status.DisbandedColony {
					continue
				}

				/* Get planet pointer. */
				planet := nampla.Planet
				if planet == nil {
					panic("assert(planet != nil)")
				}

				ls_needed := species.LifeSupportNeeded(planet)

				production_penalty := 0
				if ls_needed != 0 {
					production_penalty = (100 * ls_needed) / species.TechLevel[LS]
				}

				RMs_produced := (10 * species.TechLevel[MI] * nampla.MIBase) / planet.MiningDifficulty
				RMs_produced -= (production_penalty * RMs_produced) / 100

				production_capacity := (species.TechLevel[MA] * nampla.MABase) / 10
				production_capacity -= (production_penalty * production_capacity) / 100

				var balance int
				if nampla.Status.MiningColony {
					balance = (2 * RMs_produced) / 3
				} else if nampla.Status.ResortColony {
					balance = (2 * production_capacity) / 3
				} else {
					RMs_produced += nampla.ItemQuantity[RM]
					if RMs_produced > production_capacity {
						balance = production_capacity
					} else {
						balance = RMs_produced
					}
				}

				balance = ((planet.EconEfficiency * balance) + 50) / 100

				total_species_production += balance
			}

			// If cost is greater than production, take as much as possible from EUs in treasury.
			// 	if (fleet_maintenance_cost > total_species_production) {
			// 		if (fleet_maintenance_cost > species.EconUnits) {
			// 			fleet_maintenance_cost -= species.EconUnits;
			// 			species.EconUnits = 0;
			// 		} else {
			// 			species.EconUnits -= fleet_maintenance_cost;
			// 			fleet_maintenance_cost = 0;
			// 		}
			// 	}

			/* Save fleet maintenance results. */
			species.FleetCost = fleet_maintenance_cost
			if total_species_production > 0 {
				species.FleetPercentCost = (10000 * fleet_maintenance_cost) / total_species_production
			} else {
				species.FleetPercentCost = 10000
			}
		}
	}

	// bump the turn number
	game.CurrentTurn++

	// save the results
	err = g.Write(galaxyPath, isVerbose)
	if err != nil {
		return err
	}
	err = game.Write(galaxyPath, isVerbose)
	if err != nil {
		return err
	}

	return nil
}

// TODO: try to get straight on Turn 0 being setup and Turn 1 being first turn orders are processed
func (game *GameData) IsSetupTurn() bool {
	return game.CurrentTurn == 0
}

func (game *GameData) Report(g *GalaxyData, argv []string, galaxyPath, outputPath string, isTest, isVerbose bool) error {
	/* Check if we are doing all species, or just one or more specified  ones. */
	list := make(map[int]*SpeciesData)
	for _, arg := range argv {
		if id, err := strconv.Atoi(arg); err == nil {
			if species := g.GetSpeciesByNumber(id); species != nil {
				list[id] = species
			}
		}
	}
	// we want to keep the sort order from the galaxy's list, so copy it in that order
	var speciesFilter []*SpeciesData
	for _, species := range g.AllSpecies() {
		if len(list) == 0 {
			// if no species specified, use the entire list
			speciesFilter = append(speciesFilter, species)
		} else if _, ok := list[species.Number]; ok {
			// otherwise, just the ones specified
			speciesFilter = append(speciesFilter, species)
		}
	}

	/* Generate a report for each species in the filter */
	for _, species := range speciesFilter {
		/* Print message for gamemaster. */
		if isVerbose {
			fmt.Printf("[report] generating turn %d report for species #%d, SP %s...\n", game.CurrentTurn, species.Number, species.Name)
		}

		/* Open report file for writing. */
		filename := filepath.Join(outputPath, fmt.Sprintf("sp%02d.rpt", species.Number))
		report_file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("%+v\n", err)
			panic(fmt.Sprintf("\n\tCannot open '%s' for writing!\n\n", filename))
		}
		fmt.Printf("[report] created turn %d report %s\n", game.CurrentTurn, filename)

		// TODO: track down ignore_field_distorters and truncate_name initialization
		ignore_field_distorters, truncate_name := false, false

		if err := species.Report(report_file, galaxyPath, game.CurrentTurn, isTest, ignore_field_distorters, truncate_name, DoLocations(g), g.GetPlanet, g.GetSpeciesByNumber, g.AllSpecies()); err != nil {
			return err
		}

		if !isTest {
			/* Generate order section. */
			GenerateOrders(report_file, g, species, ignore_field_distorters, truncate_name)
		}

		/* Clean up for this species. */
	}
	/* Clean up and exit. */
	return nil
}

func (game *GameData) TurnDir() string {
	return fmt.Sprintf("t%06d", game.CurrentTurn)
}

func (game *GameData) Write(outputPath string, isVerbose bool) error {
	gameFile := filepath.Join(outputPath, "game.json")
	if isVerbose {
		fmt.Printf("[game] %-30s == %q\n", "GAME_FILE", gameFile)
	}
	if b, err := json.MarshalIndent(game, "  ", "  "); err != nil {
		return err
	} else if err := ioutil.WriteFile(gameFile, b, 0644); err != nil {
		return err
	}
	return nil
}
