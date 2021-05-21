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

		currentTurn, _ := cmd.Flags().GetBool("current-turn")
		turn, _ := cmd.Flags().GetInt("turn")
		if isVerbose {
			fmt.Printf("[report] %-30s == %v\n", "CURRENT_TURN", currentTurn)
			fmt.Printf("[report] %-30s == %d\n", "TURN", turn)
			fmt.Printf("[report] %-30s == %v\n", "VERBOSE_MODE", isVerbose)
		}
		if currentTurn && turn != 0 {
			return fmt.Errorf("you must not specify both --current-turn and --turn")
		} else if turn != 0 {
			if turn < 1 || turn > 999999 {
				return fmt.Errorf("turn must be between 1 and 999999")
			}
		}

		game, err := fh.GetGame(galaxyPath, isVerbose)
		if err != nil {
			return err
		}

		if !currentTurn {
			game.CurrentTurn--
		} else if turn != 0 {
			game.CurrentTurn = turn
		}
		if game.CurrentTurn < 0 {
			game.CurrentTurn = 0
		}
		if isVerbose {
			fmt.Printf("[report] %-30s == %q\n", "REPORT_TURN", game.CurrentTurn)
		}

		turnPath := filepath.Join(galaxyPath, game.TurnDir())
		if isVerbose {
			fmt.Printf("[report] %-30s == %q\n", "TURN_PATH", turnPath)
		}
		g, err := fh.GetGalaxy(turnPath)
		if err != nil {
			return err
		}

		outputPath := filepath.Join(galaxyPath, game.TurnDir())
		if isVerbose {
			fmt.Printf("[report] %-30s == %q\n", "OUTPUT_PATH", outputPath)
		}
		err = game.Report(g, os.Args, galaxyPath, outputPath, isTest, isVerbose)
		if err != nil {
			return err
		}

		fmt.Printf("Finished report in %v\n", time.Now().Sub(started))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().Bool("current-turn", false, "report on current turn (default is prior turn)")
	reportCmd.Flags().Int("turn", 0, "report on specified turn")
}
