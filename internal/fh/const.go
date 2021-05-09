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

import (
	"bytes"
	"encoding/json"
	"fmt"
)

var StarSizeChar = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

// Star colors
type StarColor int

const (
	BLUE         = 1
	BLUE_WHITE   = 2
	WHITE        = 3
	YELLOW_WHITE = 4
	YELLOW       = 5
	ORANGE       = 6
	RED          = 7
)

func (t StarColor) Char() string {
	switch t {
	case BLUE:
		return "O"
	case BLUE_WHITE:
		return "B"
	case WHITE:
		return "A"
	case YELLOW_WHITE:
		return "F"
	case YELLOW:
		return "G"
	case ORANGE:
		return "K"
	case RED:
		return "M"
	}
	return " "
}

func (t StarColor) String() string {
	switch t {
	case BLUE:
		return "blue"
	case BLUE_WHITE:
		return "blue-white"
	case WHITE:
		return "white"
	case YELLOW_WHITE:
		return "yellow-white"
	case YELLOW:
		return "yellow"
	case ORANGE:
		return "orange"
	case RED:
		return "red"
	}
	return "unknown"
}

// MarshalJSON marshals the enum as a quoted json string
func (t StarColor) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(t.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshals a quoted json string to the enum value
func (t *StarColor) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	switch string(b) {
	case `"blue"`:
		*t = BLUE
	case `"blue-white"`:
		*t = BLUE_WHITE
	case `"white"`:
		*t = WHITE
	case `"yellow-white"`:
		*t = YELLOW_WHITE
	case `"yellow"`:
		*t = YELLOW
	case `"orange"`:
		*t = ORANGE
	case `"red"`:
		*t = RED
	default:
		return fmt.Errorf("invalid StarColor %q", string(b))
	}
	return nil
}

// Gas types
type GasType int

const (
	/* Gases in planetary atmospheres. */
	H2  = 1  /* Hydrogen */
	CH4 = 2  /* Methane */
	HE  = 3  /* Helium */
	NH3 = 4  /* Ammonia */
	N2  = 5  /* Nitrogen */
	CO2 = 6  /* Carbon Dioxide */
	O2  = 7  /* Oxygen */
	HCL = 8  /* Hydrogen Chloride */
	CL2 = 9  /* Chlorine */
	F2  = 10 /* Fluorine */
	H2O = 11 /* Steam */
	SO2 = 12 /* Sulfur Dioxide */
	H2S = 13 /* Hydrogen Sulfide */
)

func (t GasType) Char() string {
	switch t {
	case H2:
		return "H2"
	case CH4:
		return "CH4"
	case HE:
		return "He"
	case NH3:
		return "NH3"
	case N2:
		return "N2"
	case CO2:
		return "CO2"
	case O2:
		return "O2"
	case HCL:
		return "HCl"
	case CL2:
		return "Cl2"
	case F2:
		return "F2"
	case H2O:
		return "H2O"
	case SO2:
		return "SO2"
	case H2S:
		return "H2S"
	}
	return "   "
}

func (t GasType) String() string {
	switch t {
	case H2:
		return "Hydrogen"
	case CH4:
		return "Methane"
	case HE:
		return "Helium"
	case NH3:
		return "Ammonia"
	case N2:
		return "Nitrogen"
	case CO2:
		return "Carbon Dioxide"
	case O2:
		return "Oxygen"
	case HCL:
		return "Hydrogen Chloride"
	case CL2:
		return "Chlorine"
	case F2:
		return "Fluorine"
	case H2O:
		return "Steam"
	case SO2:
		return "Sulfur Dioxide"
	case H2S:
		return "Hydrogen Sulfide"
	}
	return "unknown"
}

// MarshalJSON marshals the enum as a quoted json string
func (t GasType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(t.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshals a quoted json string to the enum value
func (t *GasType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	switch string(b) {
	case `"Hydrogen"`:
		*t = H2
	case `"Methane"`:
		*t = CH4
	case `"Helium"`:
		*t = HE
	case `"Ammonia"`:
		*t = NH3
	case `"Nitrogen"`:
		*t = N2
	case `"Carbon Dioxide"`:
		*t = CO2
	case `"Oxygen"`:
		*t = O2
	case `"Hydrogen Chloride"`:
		*t = HCL
	case `"Chlorine"`:
		*t = CL2
	case `"Fluorine"`:
		*t = F2
	case `"Steam"`:
		*t = H2O
	case `"Sulfur Dioxide"`:
		*t = SO2
	case `"Hydrogen Sulfide"`:
		*t = H2S
	default:
		return fmt.Errorf("invalid GasType %q", string(b))
	}
	return nil
}

type PlanetSpecialType int

const (
	NOT_SPECIAL          = 0
	IDEAL_HOME_PLANET    = 1
	IDEAL_COLONY_PLANET  = 2
	RADIOACTIVE_HELLHOLE = 3
)

func (t PlanetSpecialType) String() string {
	switch t {
	case NOT_SPECIAL:
		return "not-special"
	case IDEAL_HOME_PLANET:
		return "ideal-home-planet"
	case IDEAL_COLONY_PLANET:
		return "ideal-colony-planet"
	case RADIOACTIVE_HELLHOLE:
		return "radioactive-hellhole"
	}
	return "unknown"
}

// MarshalJSON marshals the enum as a quoted json string
func (t PlanetSpecialType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(t.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshals a quoted json string to the enum value
func (t *PlanetSpecialType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	switch string(b) {
	case `"not-special"`:
		*t = NOT_SPECIAL
	case `"ideal-home-planet"`:
		*t = IDEAL_HOME_PLANET
	case `"ideal-colony-planet"`:
		*t = IDEAL_COLONY_PLANET
	case `"radioactive-hellhole"`:
		*t = RADIOACTIVE_HELLHOLE
	default:
		return fmt.Errorf("invalid StarType %q", string(b))
	}
	return nil
}

/* Star types. */
type StarType int

const (
	DWARF         = 1
	DEGENERATE    = 2
	MAIN_SEQUENCE = 3
	GIANT         = 4
)

func (t StarType) Char() string {
	switch t {
	case DWARF:
		return "d"
	case DEGENERATE:
		return "D"
	case MAIN_SEQUENCE:
		return " "
	case GIANT:
		return "g"
	}
	return " "
}

func (t StarType) String() string {
	switch t {
	case DWARF:
		return "dwarf"
	case DEGENERATE:
		return "degenerate"
	case MAIN_SEQUENCE:
		return "main-sequence"
	case GIANT:
		return "giant"
	}
	return "unknown"
}

// MarshalJSON marshals the enum as a quoted json string
func (t StarType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(t.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshals a quoted json string to the enum value
func (t *StarType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	switch string(b) {
	case `"dwarf"`:
		*t = DWARF
	case `"degenerate"`:
		*t = DEGENERATE
	case `"main-sequence"`:
		*t = MAIN_SEQUENCE
	case `"giant"`:
		*t = GIANT
	default:
		return fmt.Errorf("invalid StarType %q", string(b))
	}
	return nil
}
