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

package command

import (
	"fmt"
	"github.com/mdhender/fh/internal/fh"
	"github.com/mdhender/fh/internal/prng"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// finishCmd implements the finish command
var finishCmd = &cobra.Command{
	Use:   "finish",
	Short: "Finish out a turn",
	Long: `The finish command creates the 'locations.dat' file, updates
populations, handle inter-species transactions, and does some
housekeeping chores.

This command should be run immediately before running the
Report command; i.e. immediately after the last run of AddSpecies
in the very first turn, or immediately after running PostArrival
on all subsequent turns.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		started := time.Now()
		prng.Seed(randomSeed) // seed random number generator

		if isVerbose {
			fmt.Printf("[finish] %-30s == %v\n", "TEST_MODE", isTest)
			fmt.Printf("[finish] %-30s == %v\n", "VERBOSE_MODE", isVerbose)
		}

		game, err := fh.GetGame(galaxyPath, isVerbose)
		if err != nil {
			return err
		}

		turnPath := filepath.Join(galaxyPath, game.TurnDir())
		if isVerbose {
			fmt.Printf("[finish] all output will be created in %s\n", turnPath)
		}

		logFile, err := os.Create(filepath.Join(turnPath, "finish.log"))
		if err != nil {
			return err
		}

		err = game.Finish(logFile, galaxyPath, isTest, isVerbose)
		if err != nil {
			panic(err)
			return err
		}

		fmt.Printf("Finished file %q in %v\n", filepath.Join(galaxyPath, game.TurnDir(), "galaxy.json"), time.Now().Sub(started))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(finishCmd)
}
