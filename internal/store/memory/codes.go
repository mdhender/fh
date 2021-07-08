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

package memory

import (
	"fmt"
	"strconv"
	"strings"
)

func (ds *Store) Codes(code string) (d struct {
	Code            string
	Name            string
	Descr           string
	StorageRequired int
}) {
	switch code {
	case "AU":
		d.Name = "Colonial Manufacturing Unit"
	case "BAS":
		d.Name = "Starbase"
	case "BC":
		d.Name = "Battlecruiser"
	case "BI":
		d.Name = "Biology tech level"
	case "BM":
		d.Name = "Battlemoon"
	case "BR":
		d.Name = "Battlestar"
	case "BS":
		d.Name = "Battleship"
	case "BW":
		d.Name = "Battleworld"
	case "CA":
		d.Name = "Heavy Cruiser"
	case "CC":
		d.Name = "Command Cruiser"
	case "CL":
		d.Name = "Light Cruiser"
	case "CS":
		d.Name = "Strike Cruiser"
	case "CT":
		d.Name = "Corvette"
	case "CU":
		d.Name = "Colonist Unit"
	case "DD":
		d.Name = "Destroyer"
	case "DN":
		d.Name = "Dreadnought"
	case "DR":
		d.Name = "Damage Repair Unit"
	case "ES":
		d.Name = "Escort"
	case "FD":
		d.Name = "Field Distortion Unit"
	case "FF":
		d.Name = "Frigate"
	case "FJ":
		d.Name = "Forced Jump Unit"
	case "FM":
		d.Name = "Forced Mis-jump Unit"
	case "FS":
		d.Name = "Fail-Safe Jump Unit"
	case "GT":
		d.Name = "Gravitic Telescope Unit"
	case "GUn":
		d.Name = "Auxiliary Gun Unit, Mark-n"
	case "GV":
		d.Name = "Gravitics tech level"
	case "GW":
		d.Name = "Germ Warfare Bomb"
	case "IU":
		d.Name = "Colonial Mining Unit"
	case "JP":
		d.Name = "Jump Portal Units"
	case "LS":
		d.Name = "Life Support tech level"
	case "MA":
		d.Name = "Manufacturing tech level"
	case "MI":
		d.Name = "Mining tech level"
	case "ML":
		d.Name = "Military tech level"
	case "PB":
		d.Name = "Picketboat"
	case "PD":
		d.Name = "Planetary Defense Unit"
	case "PL":
		d.Name = "Planet"
	case "RM":
		d.Name = "Raw Material Unit"
	case "SD":
		d.Name = "Super Dreadnought"
	case "SGn":
		d.Name = "Auxiliary Shield Generator, Mark-n"
	case "SP":
		d.Name = "Species"
	case "SU":
		d.Name = "Starbase Unit"
	case "TP":
		d.Name = "Terraforming Plant"
	case "TRn":
		d.Name = "Transport, eg. TR7 for 70,000 tons, TR14 for 140,000 tons, etc."
	}
	return d
}

func (i Item) Descr() string {
	switch i.Code {
	case "AU":
		return "Colonial Manufacturing Unit"
	case "BAS":
		return "Starbase"
	case "BC":
		return "Battlecruiser"
	case "BI":
		return "Biology tech level"
	case "BM":
		return "Battlemoon"
	case "BR":
		return "Battlestar"
	case "BS":
		return "Battleship"
	case "BW":
		return "Battleworld"
	case "CA":
		return "Heavy Cruiser"
	case "CC":
		return "Command Cruiser"
	case "CL":
		return "Light Cruiser"
	case "CS":
		return "Strike Cruiser"
	case "CT":
		return "Corvette"
	case "CU":
		return "Colonist Unit"
	case "DD":
		return "Destroyer"
	case "DN":
		return "Dreadnought"
	case "DR":
		return "Damage Repair Unit"
	case "ES":
		return "Escort"
	case "FD":
		return "Field Distortion Unit"
	case "FF":
		return "Frigate"
	case "FJ":
		return "Forced Jump Unit"
	case "FM":
		return "Forced Mis-jump Unit"
	case "FS":
		return "Fail-Safe Jump Unit"
	case "GT":
		return "Gravitic Telescope Unit"
	case "GV":
		return "Gravitics tech level"
	case "GW":
		return "Germ Warfare Bomb"
	case "IU":
		return "Colonial Mining Unit"
	case "JP":
		return "Jump Portal Units"
	case "LS":
		return "Life Support tech level"
	case "MA":
		return "Manufacturing tech level"
	case "MI":
		return "Mining tech level"
	case "ML":
		return "Military tech level"
	case "PB":
		return "Picketboat"
	case "PD":
		return "Planetary Defense Unit"
	case "PL":
		return "Planet"
	case "RM":
		return "Raw Material Unit"
	case "SD":
		return "Super Dreadnought"
	case "SP":
		return "Species"
	case "SU":
		return "Starbase Unit"
	case "TP":
		return "Terraforming Plant"
	}
	if strings.HasPrefix(i.Code, "GU") {
		return "Auxiliary Gun Unit, Mark-n" + i.Code[2:]
	}
	if strings.HasPrefix(i.Code, "TR") {
		return "Transport, eg. TR7 for 70,000 tons, TR14 for 140,000 tons, etc."
	}
	if strings.HasPrefix(i.Code, "SG") {
		return "Auxiliary Shield Generator, Mark-" + i.Code[2:]
	}

	return fmt.Sprintf("?%s?", i.Code)
}

func (s Ship) CarryingCapacity() int {
	switch s.Code {
	case "PB":
		return 1
	case "CT":
		return 2
	case "ES":
		return 5
	case "FF":
		return 10
	case "DD":
		return 15
	case "CL":
		return 20
	case "CS":
		return 25
	case "CA":
		return 30
	case "CC":
		return 35
	case "BC":
		return 40
	case "BS":
		return 45
	case "DN":
		return 50
	case "SD":
		return 55
	case "BM":
		return 60
	case "BW":
		return 65
	case "BR":
		return 70
	case "BAS":
		return 0
	case "SU":
		return 0
	}
	if strings.HasPrefix(s.Code, "TR") {
		if c, err := strconv.Atoi(s.Code[2:]); err == nil {
			return c * 10_000
		}
	}

	return 0
}

func (s Ship) Class() string {
	switch s.Code {
	case "PB":
		return "Picketboat"
	case "CT":
		return "Corvette"
	case "ES":
		return "Escort"
	case "FF":
		return "Frigate"
	case "DD":
		return "Destroyer"
	case "CL":
		return "Light Cruiser"
	case "CS":
		return "Strike Cruiser"
	case "CA":
		return "Heavy Cruiser"
	case "CC":
		return "Command Cruiser"
	case "BC":
		return "Battlecruiser"
	case "BS":
		return "Battleship"
	case "DN":
		return "Dreadnought"
	case "SD":
		return "Super Dreadnought"
	case "BM":
		return "Battlemoon"
	case "BW":
		return "Battleworld"
	case "BR":
		return "Battlestar"
	case "BAS":
		return "Starbase"
	case "SU":
		return "Starbase Unit"
	}
	if strings.HasPrefix(s.Code, "TR") {
		return "Transport, eg. TR7 for 70,000 tons, TR14 for 140,000 tons, etc."
	}

	return fmt.Sprintf("?%s?", s.Code)
}

func (s Ship) FTLCost() int {
	switch s.Code {
	case "PB":
		return 100
	case "CT":
		return 200
	case "ES":
		return 500
	case "FF":
		return 1_000
	case "DD":
		return 1_500
	case "CL":
		return 2_000
	case "CS":
		return 2_500
	case "CA":
		return 3_000
	case "CC":
		return 3_500
	case "BC":
		return 4_000
	case "BS":
		return 4_500
	case "DN":
		return 5_000
	case "SD":
		return 5_500
	case "BM":
		return 6_000
	case "BW":
		return 6_500
	case "BR":
		return 7_000
	case "BAS":
		return 0
	case "SU":
		return 0
	}
	if strings.HasPrefix(s.Code, "TR") {
		if c, err := strconv.Atoi(s.Code[2:]); err == nil {
			return c * 50
		}
	}

	return 0
}

func MaxTonnage(maLevel int) int {
	return maLevel * 5_000
}

func (s Ship) MinMALevel() int {
	switch s.Code {
	case "PB":
		return 2
	case "CT":
		return 4
	case "ES":
		return 10
	case "FF":
		return 20
	case "DD":
		return 30
	case "CL":
		return 40
	case "CS":
		return 50
	case "CA":
		return 60
	case "CC":
		return 70
	case "BC":
		return 80
	case "BS":
		return 90
	case "DN":
		return 100
	case "SD":
		return 110
	case "BM":
		return 120
	case "BW":
		return 130
	case "BR":
		return 140
	}
	if strings.HasPrefix(s.Code, "TR") {
		if c, err := strconv.Atoi(s.Code[2:]); err == nil {
			return c * 2
		}
	}

	return 0
}

func (s Ship) SublightCost() int {
	return 75 * s.FTLCost() / 100
}
func (s Ship) Tonnage() int {
	switch s.Code {
	case "PB":
		return 10_000
	case "CT":
		return 20_000
	case "ES":
		return 50_000
	case "FF":
		return 100_000
	case "DD":
		return 150_000
	case "CL":
		return 200_000
	case "CS":
		return 250_000
	case "CA":
		return 300_000
	case "CC":
		return 350_000
	case "BC":
		return 400_000
	case "BS":
		return 450_000
	case "DN":
		return 500_000
	case "SD":
		return 550_000
	case "BM":
		return 600_000
	case "BW":
		return 650_000
	case "BR":
		return 700_000
	case "BAS":
		return 0
	case "SU":
		return 0
	}
	if strings.HasPrefix(s.Code, "TR") {
		if c, err := strconv.Atoi(s.Code[2:]); err == nil {
			return c * 10_000
		}
	}

	return 0
}
