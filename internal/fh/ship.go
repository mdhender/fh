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
	"fmt"
	"io"
)

/* Ship classes. */
type ShipClass int

const (
	PB ShipClass = iota /* Picketboat. */
	CT                  /* Corvette. */
	ES                  /* Escort. */
	DD                  /* Destroyer. */
	FG                  /* Frigate. */
	CL                  /* Light Cruiser. */
	CS                  /* Strike Cruiser. */
	CA                  /* Heavy Cruiser. */
	CC                  /* Command Cruiser. */
	BC                  /* Battlecruiser. */
	BS                  /* Battleship. */
	DN                  /* Dreadnought. */
	SD                  /* Super Dreadnought. */
	BM                  /* Battlemoon. */
	BW                  /* Battleworld. */
	BR                  /* Battlestar. */
	BA                  /* Starbase. */
	TR                  /* Transport. */
)

const NUM_SHIP_CLASSES = 18

/* Ship types. */
type ShipType int

const (
	FTL ShipType = iota
	SUB_LIGHT
	STARBASE
)

type JumpStatus int

const (
	DidNotJump        JumpStatus = iota
	JustJumped        JumpStatus = 1
	JustMovedHere     JumpStatus = 50
	JumpedViaWormhole JumpStatus = 99
)

type ShipData struct {
	Name               string         /* Name of ship. */
	Coords             Coords         `json:"coords"` /* Current coordinates. */
	Status             ShipStatus     `json:"status"` /* Current status of ship. */
	Type               ShipType       `json:"type"`   /* Ship type. */
	Dest               Coords         `json:"dest"`   // Destination if ship was forced to jump from combat. Also used by TELESCOPE command.
	JustJumped         JumpStatus     /* Set if ship jumped this turn. */
	ArrivedViaWormhole bool           /* Ship arrived via wormhole in the PREVIOUS turn. */
	Class              ShipClass      `json:"class"` /* Ship class. */
	Tonnage            int            /* Ship tonnage divided by 10,000. */
	ItemQuantity       [MAX_ITEMS]int /* Quantity of each item carried. */
	Age                int            /* Ship age. */
	RemainingCost      int            /* The cost needed to complete the ship if still under construction. */
	LoadingPoint       int            /* Nampla index for planet where ship was last loaded with CUs. Zero = none. Use 9999 for home planet. */
	UnloadingPoint     int            /* Nampla index for planet that ship should be given orders to jump to where it will unload. Zero = none. Use 9999 for home planet. */
	Special            int            /* Different for each application. */
	alreadyListed      bool           // used for reporting
}

/* Ship status codes. */
type ShipStatus struct {
	UnderConstruction bool `json:"under_construction,omitempty"`
	OnSurface         bool `json:"on_surface,omitempty"`
	InOrbit           bool `json:"in_orbit,omitempty"`
	InDeepSpace       bool `json:"in_deep_space,"`
	JumpedInCombat    bool `json:"jumped_in_combat,omitempty"`
	ForcedJump        bool `json:"forced_jump,omitempty"`
	Destroyed         bool `json:"destroyed,omitempty"`
}
type ship_data_struct struct {
	name                 [32]char         /* Name of ship. */
	x, y, z, pn          char             /* Current coordinates. */
	status               char             /* Current status of ship. */
	ship_type            char             /* Ship type. */ // was `type`
	dest_x, dest_y       char             /* Destination if ship was forced to jump from combat. */
	dest_z               char             /* Ditto. Also used by TELESCOPE command. */
	just_jumped          char             /* Set if ship jumped this turn. */
	arrived_via_wormhole char             /* Ship arrived via wormhole in the PREVIOUS turn. */
	reserved1            char             /* Unused. Zero for now. */
	reserved2            short            /* Unused. Zero for now. */
	reserved3            short            /* Unused. Zero for now. */
	class                short            /* Ship class. */
	tonnage              short            /* Ship tonnage divided by 10,000. */
	item_quantity        [MAX_ITEMS]short /* Quantity of each item carried. */
	age                  short            /* Ship age. */
	remaining_cost       short            /* The cost needed to complete the ship if still under construction. */
	reserved4            short            /* Unused. Zero for now. */
	loading_point        short            /* Nampla index for planet where ship was last loaded with CUs. Zero = none. Use 9999 for home planet. */
	unloading_point      short            /* Nampla index for planet that ship should be given orders to jump to where it will unload. Zero = none. Use 9999 for home planet. */
	special              long             /* Different for each application. */
	padding              [28]char         /* Use for expansion. Initialized to all zeroes. */
}

var shipData = []struct {
	abbr    string
	tonnage int
	cost    int
}{
	{"PB", 1, 100},
	{"CT", 2, 200},
	{"ES", 5, 500},
	{"FF", 10, 1_000},
	{"DD", 15, 1_500},
	{"CL", 20, 2_000},
	{"CS", 25, 2_500},
	{"CA", 30, 3_000},
	{"CC", 35, 3_500},
	{"BC", 40, 4_000},
	{"BS", 45, 4_500},
	{"DN", 50, 5_000},
	{"SD", 55, 5_500},
	{"BM", 60, 6_000},
	{"BW", 65, 6_500},
	{"BR", 70, 7_000},
	{"BA", 1, 100},
	{"TR", 1, 100},
}

var shipType = []string{"", "S", "S"}

func (s *ShipData) Report(w io.Writer, printHeader, printingAlien bool) {
	if printHeader {
		fmt.Fprintf(w, "  Name                          ")
		if printingAlien {
			fmt.Fprintf(w, "                     Species\n")
		} else {
			fmt.Fprintf(w, "                 Cap. Cargo\n")
		}
		fmt.Fprintf(w, " ----------------------------------------------------------------------------\n")
	}
	ignore_field_distorters := !printingAlien
	full_ship_id := ship_name(s)
	fmt.Fprintf(w, "  %s", s.Name)
	length := len(full_ship_id)
	if printingAlien {
		padding := 50 - length
		for i := 0; i < padding; i++ {
			fmt.Fprintf(w, " ")
		}
	} else {
		padding := 46 - length
		for i := 0; i < padding; i++ {
			fmt.Fprintf(w, " ")
		}
	}
	// TODO: capacity should be set when the ship is created
	capacity := s.Tonnage
	if s.Class == BA {
		capacity = s.Tonnage * 10
	} else if s.Class == TR {
		capacity = s.Tonnage*10 + (s.Tonnage*s.Tonnage)/2
	}
	if printingAlien {
		fmt.Fprintf(w, " ")
	} else {
		fmt.Fprintf(w, "%4d  ", capacity)
		if s.Status.UnderConstruction {
			fmt.Fprintf(w, "Left to pay = %d\n", s.RemainingCost)
			return
		}
	}
	if printingAlien {
		if s.Status.OnSurface || s.ItemQuantity[FD] != s.Tonnage {
			fmt.Fprintf(w, "SP %s", species.Name)
		} else {
			fmt.Fprintf(w, "SP %d", distorted(species.Number))
		}
	} else {
		need_comma := FALSE
		for i := 0; i < MAX_ITEMS; i++ {
			if s.ItemQuantity[i] > 0 {
				if need_comma {
					fmt.Fprintf(w, ",")
				}
				fmt.Fprintf(w, "%d %s", s.ItemQuantity[i], itemData[i].abbr)
				need_comma = TRUE
			}
		}
	}
	fmt.Fprintf(w, "\n")
}
