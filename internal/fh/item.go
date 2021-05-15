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

func check_high_tech_items(tech, old_tech_level, new_tech_level int, l *Logger) {
	for _, item := range itemData {
		if !(item.criticalTech == tech && new_tech_level < item.techRequirement && old_tech_level >= item.techRequirement) {
			continue
		}

		l.String("  You now have the technology to build ")
		l.String(item.name)
		l.String("s.\n")
	}

	/* Check for high tech abilities that are not associated with specific items. */
	if tech == MA && old_tech_level < 25 && new_tech_level >= 25 {
		l.String("  You now have the technology to do interspecies construction.\n")
	}
}

var itemData = []struct {
	abbr            string
	name            string
	cost            int
	carryCapacity   int
	criticalTech    int
	techRequirement int
}{
	{"RM", "Raw Material Unit", 1, 1, MI, 1},
	{"PD", "Planetary Defense Unit", 1, 3, ML, 1},
	{"SU", "Starbase Unit", 110, 20, MA, 20},
	{"DR", "Damage Repair Unit", 50, 1, MA, 30},
	{"CU", "Colonist Unit", 1, 1, LS, 1},
	{"IU", "Colonial Mining Unit", 1, 1, MI, 1},
	{"AU", "Colonial Manufacturing Unit", 1, 1, MA, 1},
	{"FS", "Fail-Safe Jump Unit", 25, 1, GV, 20},
	{"JP", "Jump Portal Unit", 100, 10, GV, 25},
	{"FM", "Forced Misjump Unit", 100, 5, GV, 30},
	{"FJ", "Forced Jump Unit", 125, 5, GV, 40},
	{"GT", "Gravitic Telescope Unit", 500, 20, GV, 50},
	{"FD", "Field Distortion Unit", 50, 1, LS, 20},
	{"TP", "Terraforming Plant", 50000, 100, BI, 40},
	{"GW", "Germ Warfare Bomb", 1000, 100, BI, 50},
	{"SG1", "Mark-1 Shield Generator", 250, 5, LS, 10},
	{"SG2", "Mark-2 Shield Generator", 500, 10, LS, 20},
	{"SG3", "Mark-3 Shield Generator", 750, 15, LS, 30},
	{"SG4", "Mark-4 Shield Generator", 1000, 20, LS, 40},
	{"SG5", "Mark-5 Shield Generator", 1250, 25, LS, 50},
	{"SG6", "Mark-6 Shield Generator", 1500, 30, LS, 60},
	{"SG7", "Mark-7 Shield Generator", 1750, 35, LS, 70},
	{"SG8", "Mark-8 Shield Generator", 2000, 40, LS, 80},
	{"SG9", "Mark-9 Shield Generator", 2250, 45, LS, 90},
	{"GU1", "Mark-1 Gun Unit", 250, 5, ML, 10},
	{"GU2", "Mark-2 Gun Unit", 500, 10, ML, 20},
	{"GU3", "Mark-3 Gun Unit", 750, 15, ML, 30},
	{"GU4", "Mark-4 Gun Unit", 1000, 20, ML, 40},
	{"GU5", "Mark-5 Gun Unit", 1250, 25, ML, 50},
	{"GU6", "Mark-6 Gun Unit", 1500, 30, ML, 60},
	{"GU7", "Mark-7 Gun Unit", 1750, 35, ML, 70},
	{"GU8", "Mark-8 Gun Unit", 2000, 40, ML, 80},
	{"GU9", "Mark-9 Gun Unit", 2250, 45, ML, 90},
	{"X1", "X1 Unit", 9999, 9999, 99, 999},
	{"X2", "X2 Unit", 9999, 9999, 99, 999},
	{"X3", "X3 Unit", 9999, 9999, 99, 999},
	{"X4", "X4 Unit", 9999, 9999, 99, 999},
	{"X5", "X5 Unit", 9999, 9999, 99, 999},
}
