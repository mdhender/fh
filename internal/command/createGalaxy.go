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
	"path/filepath"
	"strings"
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

		startupPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("unable to determine startup path: %w", err)
		}
		startupPath = filepath.Clean(startupPath)
		fmt.Printf("[create] %-30s == %q\n", "STARTUP_PATH", startupPath)

		setupFile, err := cmd.Flags().GetString("setup-file")
		if err != nil {
			return err
		} else if setupFile == "" {
			return fmt.Errorf("you must specify a valid setup file name")
		}
		fmt.Printf("[create] %-30s == %q\n", "SETUP_FILE", setupFile)
		// try to use the location of the setup file to get a default path for the galaxy
		setupPath, file := filepath.Split(setupFile)
		if err = os.Chdir(setupPath); err != nil {
			return fmt.Errorf("unable to set def to setup path: %w", err)
		} else if setupPath, err = os.Getwd(); err != nil {
			return fmt.Errorf("unable to determine setup path: %w", err)
		}
		setupPath = filepath.Clean(setupPath)
		fmt.Printf("[create] %-30s == %q\n", "SETUP_PATH", setupPath)
		setupFile = filepath.Join(setupPath, file)
		fmt.Printf("[create] %-30s == %q\n", "SETUP_FILE", setupFile)

		// return to the startup directory because???
		if err = os.Chdir(startupPath); err != nil {
			return fmt.Errorf("unable to set def to startup path: %w", err)
		}

		setupData, err := fh.GetSetup(setupPath, setupFile)
		if err != nil {
			return err
		}

		galaxyPath, err := cmd.Flags().GetString("galaxy-path")
		if err != nil {
			return err
		} else if galaxyPath != strings.TrimSpace(galaxyPath) {
			return fmt.Errorf("galaxy-path can't have leading or trailing spaces")
		} else if galaxyPath == "" {
			galaxyPath = setupData.Galaxy.Path
		} else {
			fmt.Printf("[create] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}

		game := &fh.GameData{}
		outputPath := filepath.Join(galaxyPath, game.TurnDir())

		logFile, err := os.Create(filepath.Join(outputPath, "create.log"))
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
		fmt.Printf("[create] skipping ListGalaxy step from setup_game.py\n")

		// create home planets from the templates
		for i, player := range setupData.Players {
			spec := &fh.SpeciesData{ID: fmt.Sprintf("%02d", i+1)}
			spec.Number = i + 1
			spec.Name = player.SpeciesName
			g.AddSpecies(spec)
			err = g.AddHomePlanets(logFile, galaxyPath, outputPath, setupData, &player, spec)
			if err != nil {
				return err
			}
		}

		gameFile := filepath.Join(galaxyPath, "game.json")
		err = game.Write(gameFile)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(logFile, "Created file %q, turn number %d\n", gameFile, game.CurrentTurn)

		galaxyFile := filepath.Join(outputPath, "galaxy.json")
		err = g.Write(galaxyFile)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(logFile, "Created file %q in %v\n", galaxyFile, time.Now().Sub(started))

		return nil
	},
}

func init() {
	createCmd.AddCommand(createGalaxyCmd)
	createGalaxyCmd.Flags().StringP("setup-file", "s", "", "configuration file name")
	_ = createGalaxyCmd.MarkFlagRequired("setup-file")
	createGalaxyCmd.Flags().StringP("galaxy-path", "g", "", "path to create galaxy in")
	createGalaxyCmd.Flags().Int("initial-turn-number", 0, "initial turn number (for development use only)")
}
