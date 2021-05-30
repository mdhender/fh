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
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string
var galaxyPath string
var isTest bool
var isVerbose bool
var randomSeed string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fh",
	Short: "Command line manager for Far Horizons",
	Long:  `ofh is the original command line tool for managing game data.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// These persistent flags are global to commands in this package.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "~/.ofh.yaml", "config file")
	rootCmd.PersistentFlags().StringVarP(&galaxyPath, "galaxy-path", "g", ".", "path containing game.json file")
	rootCmd.PersistentFlags().StringVar(&randomSeed, "seed", "0x00C0FFEE", "seed for random number generator")
	rootCmd.PersistentFlags().BoolVarP(&isTest, "test", "t", false, "test command")
	rootCmd.PersistentFlags().BoolVarP(&isVerbose, "verbose", "v", false, "verbose output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".fh" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".fh")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		_, _ = fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

