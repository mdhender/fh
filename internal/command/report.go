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

	"github.com/spf13/cobra"
)

// reportCmd implements the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Run a report",
	Long:  `TODO: Command for reports.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("report called")
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().BoolP("test", "t", false, "enable test mode")
	reportCmd.Flags().BoolP("verbose", "v", false, "enable verbose mode")
}
