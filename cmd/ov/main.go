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

package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run("D:\\FarHorizons\\TestGame\\galaxy.list.txt"); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(2)
	}
}

func run(glist string) error {
	systems, err := getSystems(glist)
	if err != nil {
		return err
	}
	fmt.Println(*systems[0])
	return nil
}
