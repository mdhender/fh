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

// reportCmd implements the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Run a report",
	Long:  `TODO: Command for reports.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		started := time.Now()
		prng.Seed(0x00C0FFEE) // seed random number generator

		currentTurn, _ := cmd.Flags().GetBool("current-turn")
		turn, _ := cmd.Flags().GetInt("turn")
		if verbose {
			fmt.Printf("[report] %-30s == %v\n", "CURRENT_TURN", currentTurn)
			fmt.Printf("[report] %-30s == %d\n", "TURN", turn)
			fmt.Printf("[report] %-30s == %v\n", "VERBOSE_MODE", verbose)
		}

		game, err := fh.GetGame(galaxyPath, verbose)
		if err != nil {
			return err
		}

		if !currentTurn {
			game.CurrentTurn--
		}

		turnPath := filepath.Join(galaxyPath, game.TurnDir())
		fmt.Printf("[report] %-30s == %q\n", "TURN_PATH", turnPath)
		outputPath := turnPath
		fmt.Printf("[report] %-30s == %q\n", "OUTPUT_PATH", outputPath)

		err = game.Report(os.Args, galaxyPath, testMode, verbose)
		if err != nil {
			return err
		}

		fmt.Printf("[report] Finished report in %v\n", time.Now().Sub(started))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().Bool("current-turn", false, "report on current turn (default is prior turn)")
	reportCmd.Flags().Int("turn", 0, "report on specified turn")
}
