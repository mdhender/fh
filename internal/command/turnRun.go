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
	"github.com/spf13/cobra"
	"io/ioutil"
	"path/filepath"
)

// turnRunCmd implements the turn run command
var turnRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run orders for the current turn",
	Long: `This command...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			fmt.Printf("[turn] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		game, err := fh.GetGame(galaxyPath, verbose)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Printf("[turn] %-30s == %q\n", "CURRENT_TURN", game.CurrentTurn)
		}

		turnPath := filepath.Join(galaxyPath, game.TurnDir())
		fmt.Printf("[turn] %-30s == %q\n", "TURN_PATH", turnPath)

		// don't run if interspecies.json exists and is not empty
		file := filepath.Join(turnPath, "interspecies.json")
		if verbose {
			fmt.Printf("[turn] %-30s == %q\n", "INTERSPECIES_FILE", file)
		}
		if data, err := ioutil.ReadFile(file); err == nil {
			if len(data) != 0 {
				fmt.Printf("File %q present.\nHave you forgotten to run `fh turn clean`?\n", file)
				return fmt.Errorf("interspecies data file must be empty")
			}
		}

		if game.CurrentTurn == 0 {
			fmt.Printf("Turn 0 - running Locations...\n")
		} else {
			fmt.Printf("Running NoOrders...\n")
		}
		fmt.Printf("Running Combat...\n")
		fmt.Printf("Running PreDeparture...\n")
		fmt.Printf("Running Jump...\n")
		fmt.Printf("Running Production...\n")
		fmt.Printf("Running PostArrival...\n")
		fmt.Printf("Running Locations...\n")
		fmt.Printf("Running Strike...\n")
		fmt.Printf("Running Finish...\n")
		fmt.Printf("Running Report...\n")

		return nil
	},
}

func init() {
	turnCmd.AddCommand(turnRunCmd)
}

