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

		fmt.Printf("Created galaxy %q in %v\n", setupData.Galaxy.Name, time.Now().Sub(started))

		return nil
	},
}

