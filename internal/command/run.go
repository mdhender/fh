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
	"io/ioutil"
	"path/filepath"
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.AddCommand(runCleanCmd)
	runCmd.AddCommand(runCombatCmd)
	runCmd.AddCommand(runDiscardCmd)
	runCmd.AddCommand(runFinishCmd)
	runCmd.AddCommand(runJumpCmd)
	runCmd.AddCommand(runLocationsCmd)
	runCmd.AddCommand(runNoOrdersCmd)
	runCmd.AddCommand(runPostArrivalCmd)
	runCmd.AddCommand(runPreDepartureCmd)
	runCmd.AddCommand(runProductionCmd)
	runCmd.AddCommand(runReportCmd)
	runCmd.AddCommand(runStrikeCmd)
	runCmd.AddCommand(runTurnCmd)
}

// runCmd implements the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the current turn",
	Long:  `Run the current turn.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %v\n", "VERBOSE_MODE", isVerbose)
		}
		return nil
	},
}

// runCleanCmd implements the run clean command
var runCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove temporary files",
	Long: `This command cleans up the galaxy directory by removing all
temporary files. It does not delete any files that can't be
rebuilt by running a command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		return fmt.Errorf("run clean is not implemented yet")
	},
}

// runCombatCmd implements the run combat command
var runCombatCmd = &cobra.Command{
	Use:   "combat",
	Short: "Carry out combat orders",
	Long: `This command scans the galaxy directory for order files.
For each file found, it extracts the combat orders, then
carries out the valid orders. If run with the test flag,
results will be reported but the galaxy will not be updated.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		return fmt.Errorf("run combat is not implemented yet")
	},
}

// runDiscardCmd implements the run discard command
var runDiscardCmd = &cobra.Command{
	Use:   "discard",
	Short: "Discard the current turn",
	Long: `This command discards the current turn by decrementing
the turn number in the game.json file. It does not delete
any data files. If the current turn is less than 1, then
it is set to 0, which is the setup turn.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		game, err := fh.GetGame(galaxyPath, isVerbose)
		if err != nil {
			return err
		}
		priorTurn := game.CurrentTurn
		game.CurrentTurn--
		if game.CurrentTurn < 0 {
			game.CurrentTurn = 0
		}
		if isVerbose {
			if priorTurn == game.CurrentTurn {
				fmt.Printf("[run] turn not changed from %d\n", game.CurrentTurn)
			} else {
				fmt.Printf("[run] turn decremented from %d to %d\n", priorTurn, game.CurrentTurn)
			}
		} else {
			fmt.Printf("game turn now %d\n", game.CurrentTurn)
		}
		return game.Write(galaxyPath, isVerbose)
	},
}

// runFinishCmd implements the run finish command
var runFinishCmd = &cobra.Command{
	Use:   "finish",
	Short: "Finish out a turn",
	Long: `The finish command creates the 'locations.dat' file, updates
populations, handle inter-species transactions, and does some
housekeeping chores.

This command should be run immediately before running the
Report command; i.e. immediately after the last run of AddSpecies
in the very first turn, or immediately after running PostArrival
on all subsequent turns.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		return fmt.Errorf("run finish is not implemented yet")
	},
}

// runJumpCmd implements the run jump command
var runJumpCmd = &cobra.Command{
	Use:   "jump",
	Short: "Carry out jump orders",
	Long: `This command scans the galaxy directory for order files.
For each file found, it extracts the jump orders, then
carries out the valid orders. If run with the test flag,
results will be reported but the galaxy will not be updated.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		return fmt.Errorf("run jump is not implemented yet")
	},
}

// runLocationsCmd implements the run locations command
var runLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Create the locations file",
	Long:  `This command creates a new locations file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		return fmt.Errorf("run locations is not implemented yet")
	},
}

// runNoOrdersCmd implements the run no-orders command
var runNoOrdersCmd = &cobra.Command{
	Use:   "no-orders",
	Short: "Report on players missing orders",
	Long: `
This command scans the galaxy directory for order files
and creates a report of all players that have not sent
an order file for the current turn.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		return fmt.Errorf("run no-orders is not implemented yet")
	},
}

// runPostArrivalCmd implements the run post-arrival command
var runPostArrivalCmd = &cobra.Command{
	Use:   "post-arrival",
	Short: "Carry out post-arrival orders",
	Long: `This command scans the galaxy directory for order files.
For each file found, it extracts the post-arrival orders,
then carries out the valid orders. If run with the test flag,
results will be reported but the galaxy will not be updated.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		return fmt.Errorf("run post-arrival is not implemented yet")
	},
}

// runPreDepartureCmd implements the run pre-departure command
var runPreDepartureCmd = &cobra.Command{
	Use:   "pre-departure",
	Short: "Carry out pre-departure orders",
	Long: `This command scans the galaxy directory for order files.
For each file found, it extracts the pre-departure orders,
then carries out the valid orders. If run with the test flag,
results will be reported but the galaxy will not be updated.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		return fmt.Errorf("run pre-departure is not implemented yet")
	},
}

// runProductionCmd implements the run production command
var runProductionCmd = &cobra.Command{
	Use:   "production",
	Short: "Carry out production orders",
	Long: `This command scans the galaxy directory for order files.
For each file found, it extracts the production orders,
then carries out the valid orders. If run with the test flag,
results will be reported but the galaxy will not be updated.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		return fmt.Errorf("run production is not implemented yet")
	},
}

// runReportCmd implements the run report command
var runReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Run end of turn reports",
	Long:  `TODO: Command for reports.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		return fmt.Errorf("run report is not implemented yet")
	},
}

// runStrikeCmd implements the run strike command
var runStrikeCmd = &cobra.Command{
	Use:   "strike",
	Short: "Carry out strike orders",
	Long: `This command scans the galaxy directory for order files.
For each file found, it extracts the strike orders, then
carries out the valid orders. If run with the test flag,
results will be reported but the galaxy will not be updated.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		return fmt.Errorf("run strike is not implemented yet")
	},
}

// runTurnCmd implements the run turn command
var runTurnCmd = &cobra.Command{
	Use:   "turn",
	Short: "Run all command for the current turn",
	Long: `
This command runs the following commands in order:
    fh run no-orders
    fh run combat
    fh run pre-departure
    fh run jump
    fh run production
    fh run post-arrival
    fh run locations
    fh run strike
    fh run finish
    fh run report`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "GALAXY_PATH", galaxyPath)
		}
		game, err := fh.GetGame(galaxyPath, isVerbose)
		if err != nil {
			return err
		}
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "CURRENT_TURN", game.CurrentTurn)
		}

		turnPath := filepath.Join(galaxyPath, game.TurnDir())
		fmt.Printf("[run] %-30s == %q\n", "TURN_PATH", turnPath)

		// don't run if interspecies.json exists and is not empty
		file := filepath.Join(turnPath, "interspecies.json")
		if isVerbose {
			fmt.Printf("[run] %-30s == %q\n", "INTERSPECIES_FILE", file)
		}
		if data, err := ioutil.ReadFile(file); err == nil {
			if len(data) != 0 {
				fmt.Printf("File %q present.\nHave you forgotten to run `fh run clean`?\n", file)
				return fmt.Errorf("interspecies data file must be empty")
			}
		}

		if game.CurrentTurn == 0 {
			fmt.Printf("Turn 0 - running Locations...\n")
		} else {
			fmt.Printf("Running NoOrders...\n")
		}
		fmt.Printf("Running Combat...\n")
		fmt.Printf("Running PreDeparture...\n")
		fmt.Printf("Running Jump...\n")
		fmt.Printf("Running Production...\n")
		fmt.Printf("Running PostArrival...\n")
		fmt.Printf("Running Locations...\n")
		fmt.Printf("Running Strike...\n")
		fmt.Printf("Running Finish...\n")
		fmt.Printf("Running Report...\n")

		return nil
	},
}
