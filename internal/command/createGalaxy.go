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
	"github.com/spf13/cobra"
	"os"
	"path"
	"time"
)

// createGalaxyCmd implements the create galaxy command
var createGalaxyCmd = &cobra.Command{
	Use:   "galaxy",
	Short: "Create a new galaxy",
	Long: `This commands loads setup data from a
configuration file, then creates a new galaxy file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		started := time.Now()
		prng.Seed(0x00C0FFEE) // seed random number generator

		galaxyPath, err := cmd.Flags().GetString("galaxy-path")
		if err != nil {
			return err
		} else if galaxyPath == "" {
			return fmt.Errorf("you must specify a valid path to read and create galaxy data in")
		}

		logFile, err := os.Create(path.Join(galaxyPath, "create-galaxy.log"))
		if err != nil {
			return err
		}

		setupData, err := fh.GetSetup(path.Join(galaxyPath, "setup.json"))
		if err != nil {
			return err
		}

		// NewGalaxy step in setup_game.py
		g, err := fh.GenerateGalaxy(logFile, setupData)
		if err != nil {
			return err
		}

		if setupData.Galaxy.MinimumDistance < 1 || setupData.Galaxy.MinimumDistance > g.Radius*2 {
			return fmt.Errorf("minimum-distance must be between 1 and %d", g.Radius*2)
		}

		// MakeHomes step in setup_game.py
		err = g.MakeHomeTemplates(logFile)
		if err != nil {
			return err
		}

		// skip ListGalaxy step in setup_game.py

		// create home planets from the templates
		for i, player := range setupData.Players {
			spec := &fh.SpeciesData{ID: fmt.Sprintf("%02d", i+1)}
			spec.Number = i + 1
			spec.Name = player.SpeciesName
			g.AddSpecies(spec)
			err = g.AddHomePlanets(logFile, galaxyPath, setupData, &player, spec)
			if err != nil {
				return err
			}
		}

		err = g.Write(path.Join(galaxyPath, "galaxy.json"))
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(logFile, "Created file %q in %v\n", path.Join(galaxyPath, "galaxy.json"), time.Now().Sub(started))
		return nil
	},
}

func init() {
	createCmd.AddCommand(createGalaxyCmd)
	createGalaxyCmd.Flags().StringP("galaxy-path", "g", "", "path to galaxy data")
	_ = createGalaxyCmd.MarkFlagRequired("galaxy-path")
}
