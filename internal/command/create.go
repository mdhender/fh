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
	"math"
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
		outputPath := galaxyPath
		if isVerbose {
			fmt.Printf("[create] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
			fmt.Printf("[create] %-30s == %q\n", "OUTPUT_PATH", outputPath)
		}

		logFileName := filepath.Join(outputPath, "create.log")
		if isVerbose {
			fmt.Printf("[create] %-30s == %q\n", "LOG_FILE", logFileName)
		}
		logFile, err := os.Create(logFileName)
		if err != nil {
			return err
		}
		w := &fh.Writer{File: logFile}
		defer w.Close()

		playersFileName := filepath.Join(galaxyPath, "players.json")
		if isVerbose {
			fmt.Printf("[create] %-30s == %q\n", "PLAYERS_FILE", playersFileName)
		}
		players, err := fh.GetPlayers(playersFileName, isVerbose)
		if err != nil {
			return err
		}

		game := &fh.GameData{}
		if err = fh.ValidateNumberOfPlayers(len(players)); err != nil {
			return err
		}

		ga, err := fh.NewGalaxy(w, setupData)
		if err != nil {
			fmt.Printf("%+v\n", err)
			os.Exit(2)
		}
		fmt.Printf("[create] galaxy skeleton created in %v\n", time.Now().Sub(started))

		for _, p := range players {
			if err := ga.AddPlayer(w, p); err != nil {
				fmt.Printf("%+v\n", err)
				os.Exit(2)
			}
		}
		fmt.Printf("[create] species skeleton created in %v\n", time.Now().Sub(started))

		r, err := ga.CalculateRadius(w, ga.NumberOfSpecies(), setupData.Galaxy.LargeCluster)
		if err != nil {
			fmt.Printf("%+v\n", err)
			os.Exit(2)
		}
		if r < setupData.Galaxy.Radius.Minimum {
			fmt.Printf("[create] forcing radius to minimum %d parsecs\n", setupData.Galaxy.Radius.Minimum)
			r = setupData.Galaxy.Radius.Minimum
		}
		if r > setupData.Galaxy.Radius.Maximum {
			fmt.Printf("[create] forcing radius to maximum %d parsecs\n", setupData.Galaxy.Radius.Maximum)
			r = setupData.Galaxy.Radius.Maximum
		}
		ga.Radius = r
		fmt.Printf("[create] setting radius to %d parsecs\n", ga.Radius)

		points, pointsIn := 0, 0
		for x := -1 * r; x <= r; x++ {
			for y := -1 * r; y <= r; y++ {
				for z := -1 * r; z <= r; z++ {
					points++
					if x*x+y*y+z*z < r*r {
						pointsIn++
					}
				}
			}
		}
		fmt.Printf("[create] cluster contains %s possible systems\n", fh.Commas(pointsIn))
		density := 0.002488
		switch setupData.Galaxy.Density {
		case "sparse":
			density = 0.001658
		case "high":
			density = 0.003732
		}
		numberOfSystems := int(math.Round(float64(pointsIn) * density))
		fmt.Printf("[create] cluster density %7.4f%% will generate %s systems\n", 100*density, fh.Commas(numberOfSystems))

		fmt.Printf("[create] generating home systems no closer together than %d parsecs\n", setupData.Galaxy.MinimumDistance)
		if err = ga.CreateHomeSystems(w, setupData.Galaxy.MinimumDistance); err != nil {
			fmt.Printf("%+v\n", err)
			os.Exit(2)
		}
		fmt.Printf("[create] created %d home systems in %v\n", ga.NumberOfSpecies(), time.Now().Sub(started))

		ga.CheckSpacing(w, true)

		l := &fh.Logger{Stdout: logFile}
		defer l.Close()

		// NewGalaxy step in setup_game.py
		g, err := fh.GenerateGalaxy(l, setupData, galaxyPath, players)
		if err != nil {
			return err
		}

		err = game.Write(galaxyPath, isVerbose)
		if err != nil {
			return err
		}
		fmt.Printf("Created game file, turn number %d\n", game.CurrentTurn)

		err = g.Write(galaxyPath, isVerbose)
		if err != nil {
			return err
		}
		fmt.Printf("Created galaxy file\n")

		fmt.Printf("Created galaxy in %v\n", time.Now().Sub(started))

		return nil
	},
}
