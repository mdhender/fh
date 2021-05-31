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
	"github.com/mdhender/fh/config"
	"github.com/mdhender/fh/internal/prng"
	"github.com/mdhender/fh/logger"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("setup-file", "s", "setup.json", "setup file name")
}

// createCmd implements the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new galaxy",
	Long: `This commands loads configuration data from the setup.json
files, then creates a new galaxy file. The results are saved
to the game directory specified in the setup file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		started := time.Now()
		prng.Seed(randomSeed) // seed random number generator
		isVerbose = true

		startupPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("unable to determine startup path: %w", err)
		}
		startupPath = filepath.Clean(startupPath)
		fmt.Printf("%-30s == %q\n", "STARTUP_PATH", startupPath)

		setupFile, err := cmd.Flags().GetString("setup-file")
		if err != nil {
			return err
		} else if setupFile == "" {
			return fmt.Errorf("you must specify a valid setup file name")
		} else if galaxyPath != "." {
			return fmt.Errorf("you must not specify setup file and galaxy path")
		}
		fmt.Printf("%-30s == %q\n", "SETUP_FILE", setupFile)

		setupPath, file := filepath.Split(setupFile)
		setupPath = filepath.Clean(setupPath)
		setupFile = filepath.Join(setupPath, file)
		fmt.Printf("%-30s == %q\n", "SETUP_PATH", setupPath)
		if setupPath == "" {
			// expect to find everything in the current directory
		} else if err = os.Chdir(setupPath); err != nil {
			return fmt.Errorf("unable to set def to setup path: %w", err)
		} else if setupPath, err = os.Getwd(); err != nil {
			return fmt.Errorf("unable to determine setup path: %w", err)
		} else if err = os.Chdir(startupPath); err != nil { // return to the startup directory because???
			return fmt.Errorf("unable to set def to startup path: %w", err)
		}

		defaultGalaxyPath := setupPath
		cfg, err := config.Get(setupFile, defaultGalaxyPath, isVerbose)
		if err != nil {
			return err
		}

		galaxyPath = cfg.Galaxy.Path
		outputPath := galaxyPath
		fmt.Printf("%-30s == %q\n", "GALAXY_PATH", galaxyPath)
		fmt.Printf("%-30s == %q\n", "OUTPUT_PATH", outputPath)

		logFileName := filepath.Join(outputPath, "create.log")
		logFile, err := os.Create(logFileName)
		if err != nil {
			return err
		}
		fmt.Printf("%-30s == %q\n", "LOG_FILE", logFileName)
		w := &logger.Logger{File: logFile}
		defer w.Close()
		w.Printf("%-30s == %q\n", "SETUP_FILE", setupFile)

		w.Printf("%-30s == %q\n", "location", cfg.Galaxy.Location)
		var density float64
		switch cfg.Galaxy.Location {
		case "outer rim":
			density = 0.001658
		case "rim":
			density = 0.002488
		case "inner rim":
			density = 0.003732
		case "outer core":
			density = 0.005898
		default:
			panic("invalid cluster location")
		}
		w.Printf("%-30s == %8.6f\n", "density", density)

		w.Printf("%-30s == %d\n", "resourceLevel", cfg.Galaxy.ResourceLevel)
		var systemsPerSpecies int
		switch cfg.Galaxy.ResourceLevel {
		case 0:
			systemsPerSpecies = 6
		case 1:
			systemsPerSpecies = 7
		case 2:
			systemsPerSpecies = 8
		case 3:
			systemsPerSpecies = 9
		case 4:
			systemsPerSpecies = 13
		case 5:
			systemsPerSpecies = 15
		case 6:
			systemsPerSpecies = 20
		case 7:
			systemsPerSpecies = 30
		case 8:
			systemsPerSpecies = 40
		case 9:
			systemsPerSpecies = 50
		default:
			panic("error: resource level must be between 0 and 9")
		}
		w.Printf("%-30s == %d\n", "systemsPerSpecies", systemsPerSpecies)

		w.Printf("%-30s == %d\n", "numberOfSpecies", cfg.Species.Number)
		numberOfSystems := systemsPerSpecies * cfg.Species.Number
		w.Printf("%-30s == %d\n", "numberOfSystems", numberOfSystems)

		w.Printf("Created galaxy in %v\n", time.Now().Sub(started))

		return nil
	},
}
