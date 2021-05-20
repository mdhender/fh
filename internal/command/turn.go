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
)

// turnCmd implements the turn command
var turnCmd = &cobra.Command{
	Use:   "turn",
	Short: "Display the current turn",
	Long:  `Displays the current turn.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			fmt.Printf("[finish] %-30s == %v\n", "VERBOSE_MODE", verbose)
		}

		game, err := fh.GetGame(galaxyPath, verbose)
		if err != nil {
			return err
		}
		fmt.Printf("%d\n", game.CurrentTurn)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(turnCmd)
}
