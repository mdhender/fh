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
	"github.com/mdhender/fh/internal/orders"
	"github.com/mdhender/fh/internal/store/jsondb"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC) // force logs to be UTC

	cfg := DefaultConfig()
	err := cfg.Load()
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(2)
	}

	if errors := run(cfg); errors != nil {
		for _, err := range errors {
			fmt.Printf("%+v\n", err)
		}
		os.Exit(2)
	}
}

type TurnData struct {
	Turn         int
	EconomicBase struct {
		PerPlanet  []int
		PerSpecies []int
	}
	Species []*SpeciesTurnData
}
type SpeciesTurnData struct {
	Id        string
	Species   *jsondb.Species
	OrderFile string
	Orders    *orders.Orders
}

func run(cfg *Config) []error {
	jdb, err := jsondb.Read(filepath.Join(cfg.Data.JDB, "galaxy.json"))
	if err != nil {
		return []error{err}
	}
	if jdb == nil {
		fmt.Println("jdb is nil?")
	}

	numSpecies := jdb.Galaxy.NumSpecies
	if len(jdb.Species) > numSpecies {
		numSpecies = len(jdb.Species)
	}
	turnData := &TurnData{
		Turn:    jdb.Galaxy.TurnNumber,
		Species: make([]*SpeciesTurnData, numSpecies, numSpecies),
	}

	for i := 1; i <= numSpecies; i++ {
		turnData.Species[i-i] = &SpeciesTurnData{Id: fmt.Sprintf("SP%02d", i)}
		td := turnData.Species[i-i]
		td.Species = jdb.Species[td.Id]
		td.OrderFile = filepath.Join(cfg.Data.Orders, fmt.Sprintf("sp%02d.ord", i))

		log.Printf("orders: loading %q\n", td.OrderFile)
		o := orders.Parse(td.OrderFile)
		if o.Errors == nil {
			fmt.Printf(";; SP%02d TURN %3d\n", i, jdb.Galaxy.TurnNumber)
			for _, section := range []*orders.Section{o.Combat, o.PreDeparture, o.Jumps, o.Production, o.PostArrival, o.Strikes} {
				if section != nil {
					fmt.Printf("START %q\n", section.Name)
					for _, command := range section.Commands {
						fmt.Printf("    %-18s", command.Name)
						for _, arg := range command.Args {
							fmt.Printf(" %q", arg)
						}
						fmt.Printf("\n")
					}
				}
			}
		} else {
			for _, err := range o.Errors {
				log.Printf("%+v\n", err)
			}
		}
	}

	test, verbose := false, true
	if jdb.Galaxy.TurnNumber == 0 {
		test, verbose = !test, !verbose
	}

	// locations
	Locations(jdb, turnData, test, verbose)

	// no-orders if not the first turn
	// combat
	// pre-departure
	// jump
	// production
	// post-arrival
	// locations
	// strike
	// finish
	// reports
	// stats

	return nil
}
