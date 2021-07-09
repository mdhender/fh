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
	"math"
	"os"
)

// power returns 100 * (tonnage ^ 1.2).
// it assumes that tonnage has already been scaled down by 10,000.
func power(tonnage int) int {
	if tonnage < len(ship_power) {
		return ship_power[tonnage]
	} else if tonnage > 4068 {
		fmt.Fprintf(os.Stderr, "\n\n\tLong integer overflow will occur in call to 'power(tonnage)'!\n")
		fmt.Fprintf(os.Stderr, "\t\tActual call is power(%d).\n\n", tonnage)
		os.Exit(2)
	}
	return int(math.Round(100 * math.Pow(float64(tonnage), 1.2)))
	//// Tonnage is not in table. Break it up into two halves and get
	//// approximate result = 1.149 * (x1 + x2), using recursion if
	//// necessary.
	// x1 := tonnage / 2
	// x2 := tonnage - x1
	// return 1149 * (power(x1) + power(x2)) / 1000
}

// Look-up table for ship defensive/offensive power uses ship->tonnage
// as an index. Each value is equal to 100 * (ship->tonnage)^1.2. The
// 'power' subroutine uses recursion to calculate values for tonnages
// over 100.
var ship_power = [101]int{0, /* Zeroth element not used. */
	100, 230, 374, 528, 690, 859, 1033, 1213, 1397, 1585,
	1777, 1973, 2171, 2373, 2578, 2786, 2996, 3209, 3424, 3641,
	3861, 4082, 4306, 4532, 4759, 4988, 5220, 5452, 5687, 5923,
	6161, 6400, 6641, 6883, 7127, 7372, 7618, 7866, 8115, 8365,
	8617, 8870, 9124, 9379, 9635, 9893, 10151, 10411, 10672, 10934,
	11197, 11461, 11725, 11991, 12258, 12526, 12795, 13065, 13336, 13608,
	13880, 14154, 14428, 14703, 14979, 15256, 15534, 15813, 16092, 16373,
	16654, 16936, 17218, 17502, 17786, 18071, 18356, 18643, 18930, 19218,
	19507, 19796, 20086, 20377, 20668, 20960, 21253, 21547, 21841, 22136,
	22431, 22727, 23024, 23321, 23619, 23918, 24217, 24517, 24818, 25119,
}
