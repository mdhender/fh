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

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createGalaxyCmd)
	createGalaxyCmd.Flags().StringP("setup-file", "s", "", "configuration file name")
	_ = createGalaxyCmd.MarkFlagRequired("setup-file")
	createGalaxyCmd.Flags().Int("initial-turn-number", 0, "initial turn number (for development use only)")
}

// createCmd implements the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new item",
	Long:  `Create a completely new item and write the results to files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("create with no arguments not implemented yet")
	},
}

// createGalaxyCmd implements the create galaxy command
var createGalaxyCmd = &cobra.Command{
	Use:   "galaxy",
	Short: "Create a new galaxy",
	Long: `This commands loads configuration data from the setup.json
files, then creates a new galaxy file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		started := time.Now()
		prng.Seed(randomSeed) // seed random number generator
		isVerbose = true

		startupPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("unable to determine startup path: %w", err)
		}
		startupPath = filepath.Clean(startupPath)
		if isVerbose {
			fmt.Printf("[create] %-30s == %q\n", "STARTUP_PATH", startupPath)
		}

		setupFile, err := cmd.Flags().GetString("setup-file")
		if err != nil {
			return err
		} else if setupFile == "" {
			return fmt.Errorf("you must specify a valid setup file name")
		} else if galaxyPath != "." {
			return fmt.Errorf("you must not specify setup file and galaxy path")
		}
		if isVerbose {
			fmt.Printf("[create] %-30s == %q\n", "SETUP_FILE", setupFile)
		}

		setupPath, file := filepath.Split(setupFile)
		if err = os.Chdir(setupPath); err != nil {
			return fmt.Errorf("unable to set def to setup path: %w", err)
		} else if setupPath, err = os.Getwd(); err != nil {
			return fmt.Errorf("unable to determine setup path: %w", err)
		}
		setupPath = filepath.Clean(setupPath)
		setupFile = filepath.Join(setupPath, file)
		if isVerbose {
			fmt.Printf("[create] %-30s == %q\n", "SETUP_PATH", setupPath)
			fmt.Printf("[create] %-30s == %q\n", "SETUP_FILE", setupFile)
		}

		// return to the startup directory because???
		if err = os.Chdir(startupPath); err != nil {
			return fmt.Errorf("unable to set def to startup path: %w", err)
		}

		setupData, err := fh.GetSetup(setupPath, setupFile, isVerbose)
		if err != nil {
			return err
		}

		galaxyPath = setupData.Galaxy.Path
		if isVerbose {
			fmt.Printf("[create] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}

		players, err := fh.GetPlayers(filepath.Join(galaxyPath, "players.json"), isVerbose)
		if err != nil {
			return err
		}

		game := &fh.GameData{}
		outputPath := galaxyPath
		if isVerbose {
			fmt.Printf("[create] %-30s == %q\n", "OUTPUT_PATH", outputPath)
		}

		logFile, err := os.Create(filepath.Join(outputPath, "create.log"))
		if err != nil {
			return err
		}

		// NewGalaxy step in setup_game.py
		g, err := fh.GenerateGalaxy(logFile, setupData, players)
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
		for i, player := range players {
			spec := &fh.SpeciesData{ID: fmt.Sprintf("%02d", i+1)}
			spec.Number = i + 1
			spec.Name = player.SpeciesName
			spec.Government.Name = player.GovName
			spec.Government.Type = player.GovType
			g.AddSpecies(spec)
			err = g.AddHomePlanets(logFile, galaxyPath, outputPath, setupData, player, spec)
			if err != nil {
				return err
			}
		}

		err = game.Write(galaxyPath, isVerbose)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(logFile, "Created game file, turn number %d\n", game.CurrentTurn)

		err = g.Write(galaxyPath, isVerbose)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(logFile, "Created galaxy file\n")

		systems := &fh.Systems{}
		for _, star := range g.Stars {
			systems.Data = append(systems.Data, star)
		}
		err = systems.Write(galaxyPath, isVerbose)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(logFile, "Created systems file\n")

		_, _ = fmt.Fprintf(logFile, "Created galaxy in %v\n", time.Now().Sub(started))

		return nil
	},
}
