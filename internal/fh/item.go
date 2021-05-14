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

package fh

var itemData = []struct {
	abbr            string
	cost            int
	carryCapacity   int
	criticalTech    int
	techRequirement int
}{
	{"RM", 1, 1, MI, 1},
	{"PD", 1, 3, ML, 1},
	{"SU", 110, 20, MA, 20},
	{"DR", 50, 1, MA, 30},
	{"CU", 1, 1, LS, 1},
	{"IU", 1, 1, MI, 1},
	{"AU", 1, 1, MA, 1},
	{"FS", 25, 1, GV, 20},
	{"JP", 100, 10, GV, 25},
	{"FM", 100, 5, GV, 30},
	{"FJ", 125, 5, GV, 40},
	{"GT", 500, 20, GV, 50},
	{"FD", 50, 1, LS, 20},
	{"TP", 50000, 100, BI, 40},
	{"GW", 1000, 100, BI, 50},
	{"SG1", 250, 5, LS, 10},
	{"SG2", 500, 10, LS, 20},
	{"SG3", 750, 15, LS, 30},
	{"SG4", 1000, 20, LS, 40},
	{"SG5", 1250, 25, LS, 50},
	{"SG6", 1500, 30, LS, 60},
	{"SG7", 1750, 35, LS, 70},
	{"SG8", 2000, 40, LS, 80},
	{"SG9", 2250, 45, LS, 90},
	{"GU1", 250, 5, ML, 10},
	{"GU2", 500, 10, ML, 20},
	{"GU3", 750, 15, ML, 30},
	{"GU4", 1000, 20, ML, 40},
	{"GU5", 1250, 25, ML, 50},
	{"GU6", 1500, 30, ML, 60},
	{"GU7", 1750, 35, ML, 70},
	{"GU8", 2000, 40, ML, 80},
	{"GU9", 2250, 45, ML, 90},
	{"X1", 9999, 9999, 99, 999},
	{"X2", 9999, 9999, 99, 999},
	{"X3", 9999, 9999, 99, 999},
	{"X4", 9999, 9999, 99, 999},
	{"X5", 9999, 9999, 99, 999},
}
