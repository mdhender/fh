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
	"path/filepath"
)

// turnDiscardCmd implements the turn discard command
var turnDiscardCmd = &cobra.Command{
	Use:   "discard",
	Short: "Discard the current turn",
	Long: `This command discards the current turn by decrementing
the turn number in the game.json file. It does not delete
any data files. If the current turn is less than 1, then
it is set to 0, which is the setup turn.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			fmt.Printf("[turn] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		game, err := fh.GetGame(galaxyPath, verbose)
		if err != nil {
			return err
		}
		priorTurn := game.CurrentTurn
		game.CurrentTurn--
		if game.CurrentTurn < 0 {
			game.CurrentTurn = 0
		}
		if verbose {
			if priorTurn == game.CurrentTurn {
				fmt.Printf("[turn] turn not changed from %d\n", game.CurrentTurn)
			} else {
				fmt.Printf("[turn] turn decremented from %d to %d\n", priorTurn, game.CurrentTurn)
			}
		} else {
			fmt.Printf("game turn now %d\n", game.CurrentTurn)
		}
		return game.Write(filepath.Join(galaxyPath, "game.json"))
	},
}

func init() {
	turnCmd.AddCommand(turnDiscardCmd)
}
