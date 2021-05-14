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
	"path"
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
		prng.Seed(0x00C0FFEE) // seed random number generator

		galaxyPath, err := cmd.Flags().GetString("galaxy-path")
		if err != nil {
			return err
		} else if galaxyPath == "" {
			return fmt.Errorf("you must specify a valid path to read and create galaxy data in")
		}
		testMode, _ := cmd.Flags().GetBool("test")
		verboseMode, _ := cmd.Flags().GetBool("verbose")

		g, err := fh.GetGalaxy(path.Join(galaxyPath, "galaxy.json"))
		if err != nil {
			return err
		}
		tn := g.TurnNumber

		logFile, err := os.Create(path.Join(galaxyPath, fmt.Sprintf("t%06d-finish.log", g.TurnNumber)))
		if err != nil {
			return err
		}

		err = g.Finish(logFile, galaxyPath, verboseMode, testMode)
		if err != nil {
			return err
		}

		g.TurnNumber = tn
		fmt.Printf("warning: forcing turn number back to %d\n", tn)
		err = g.Write(path.Join(galaxyPath, "galaxy.json"))
		if err != nil {
			return err
		}

		fmt.Printf("Finished file %q in %v\n", path.Join(galaxyPath, "galaxy.json"), time.Now().Sub(started))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(finishCmd)
	finishCmd.Flags().BoolP("test", "t", false, "enable test mode")
	finishCmd.Flags().BoolP("verbose", "v", false, "enable verbose mode")
	finishCmd.Flags().StringP("galaxy-path", "g", "", "path to galaxy data")
	_ = finishCmd.MarkFlagRequired("galaxy-path")
}
