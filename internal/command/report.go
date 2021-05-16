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
	"path"
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
		prng.Seed(0x00C0FFEE) // seed random number generator

		galaxyPath, err := cmd.Flags().GetString("galaxy-path")
		if err != nil {
			return err
		} else if galaxyPath == "" {
			return fmt.Errorf("you must specify a valid path to read and create galaxy data in")
		}
		testMode, _ := cmd.Flags().GetBool("test")
		verboseMode, _ := cmd.Flags().GetBool("verbose")

		g, err := fh.GetGalaxy(path.Join(galaxyPath, "galaxy.json"))
		if err != nil {
			return err
		}
		err = g.Report(os.Args, galaxyPath, testMode, verboseMode)
		if err != nil {
			return err
		}

		fmt.Printf("Finished report in %v\n", time.Now().Sub(started))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().BoolP("test", "t", false, "enable test mode")
	reportCmd.Flags().BoolP("verbose", "v", false, "enable verbose mode")
	reportCmd.Flags().StringP("galaxy-path", "g", "", "path to galaxy data")
	_ = reportCmd.MarkFlagRequired("galaxy-path")
}
