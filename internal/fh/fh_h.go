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

const TRUE = 1
const FALSE = 0

/* A standard game has 15 species. */
const STANDARD_NUMBER_OF_SPECIES = 15

/* A standard game has 90 star systems. */
const STANDARD_NUMBER_OF_STAR_SYSTEMS = 90

/* A standard game has a galaxy with a radius of 20 parsecs. */
const STANDARD_GALACTIC_RADIUS = 20

/* Minimum and maximum values for a galaxy. */
const MIN_SPECIES = 1
const MAX_SPECIES = 100
const MIN_STARS = 12
const MAX_STARS = 1000
const MIN_RADIUS = 6
const MAX_RADIUS = 50
const MAX_DIAMETER = 2 * MAX_RADIUS
const MAX_PLANETS = 9 * MAX_STARS

const HP_AVAILABLE_POP = 1500

const NUM_EXTRA_NAMPLAS = 50
const NUM_EXTRA_SHIPS = 100

const MAX_LOCATIONS = 10000

type char = byte
type sp_loc_data struct {
	s, x, y, z char /* Species number, x, y, and z. */
}

type galaxy_data struct {
	d_num_species int /* Design number of species in galaxy. */
	num_species   int /* Actual number of species allocated. */
	radius        int /* Galactic radius in parsecs. */
	turn_number   int /* Current turn number. */
}

/* Assume at least 32 bits per long word. */
const NUM_CONTACT_WORDS = ((MAX_SPECIES - 1) / 32) + 1

type short = int32

type long = int
type planet_data struct {
	temperature_class char    /* Temperature class, 1-30. */
	pressure_class    char    /* Pressure class, 0-29. */
	special           char    /* 0 = not special, 1 = ideal home planet, 2 = ideal colony planet, 3 = radioactive hellhole. */
	reserved1         char    /* Reserved for future use. Zero for now. */
	gas               [4]char /* Gas in atmosphere. Zero if none. */
	gas_percent       [4]char /* Percentage of gas in atmosphere. */
	reserved2         short   /* Reserved for future use. Zero for now. */
	diameter          short   /* Diameter in thousands of kilometers. */
	gravity           short   /* Surface gravity. Multiple of Earth gravity times 100. */
	mining_difficulty short   /* Mining difficulty times 100. */
	econ_efficiency   short   /* Economic efficiency. Always 100 for a home planet. */
	md_increase       short   /* Increase in mining difficulty. */
	message           long    /* Message associated with this planet, if any. */
	reserved3         long    /* Reserved for future use. Zero for now. */
	reserved4         long    /* Reserved for future use. Zero for now. */
	reserved5         long    /* Reserved for future use. Zero for now. */
}

/* Tech level ids. */
const MI = 0 /* Mining tech level. */
const MA = 1 /* Manufacturing tech level. */
const ML = 2 /* Military tech level. */
const GV = 3 /* Gravitics tech level. */
const LS = 4 /* Life Support tech level. */
const BI = 5 /* Biology tech level. */

type species_data struct {
	name               [32]char                /* Name of species. */
	govt_name          [32]char                /* Name of government. */
	govt_type          [32]char                /* Type of government. */
	x, y, z, pn        char                    /* Coordinates of home planet. */
	required_gas       char                    /* Gas required by species. */
	required_gas_min   char                    /* Minimum needed percentage. */
	required_gas_max   char                    /* Maximum allowed percentage. */
	reserved5          char                    /* Zero for now. */
	neutral_gas        [6]char                 /* Gases neutral to species. */
	poison_gas         [6]char                 /* Gases poisonous to species. */
	auto_orders        char                    /* AUTO command was issued. */
	reserved3          char                    /* Zero for now. */
	reserved4          short                   /* Zero for now. */
	tech_level         [6]short                /* Actual tech levels. */
	init_tech_level    [6]short                /* Tech levels at start of turn. */
	tech_knowledge     [6]short                /* Unapplied tech level knowledge. */
	num_namplas        int                     /* Number of named planets, including home planet and colonies. */
	num_ships          int                     /* Number of ships. */
	tech_eps           [6]long                 /* Experience points for tech levels. */
	hp_original_base   long                    /* If non-zero, home planet was bombed either by bombardment or germ warfare and has not yet fully recovered. Value is total economic base before bombing. */
	econ_units         long                    /* Number of economic units. */
	fleet_cost         long                    /* Total fleet maintenance cost. */
	fleet_percent_cost long                    /* Fleet maintenance cost as a percentage times one hundred. */
	contact            [NUM_CONTACT_WORDS]long /* A bit is set if corresponding species has been met. */
	ally               [NUM_CONTACT_WORDS]long /* A bit is set if corresponding species is considered an ally. */
	enemy              [NUM_CONTACT_WORDS]long /* A bit is set if corresponding species is considered an enemy. */
	padding            [12]char                /* Use for expansion. Initialized to all zeroes. */
}

/* Item IDs. */
const RM = 0   /* Raw Material Units. */
const PD = 1   /* Planetary Defense Units. */
const SU = 2   /* Starbase Units. */
const DR = 3   /* Damage Repair Units. */
const CU = 4   /* Colonist Units. */
const IU = 5   /* Colonial Mining Units. */
const AU = 6   /* Colonial Manufacturing Units. */
const FS = 7   /* Fail-Safe Jump Units. */
const JP = 8   /* Jump Portal Units. */
const FM = 9   /* Forced Misjump Units. */
const FJ = 10  /* Forced Jump Units. */
const GT = 11  /* Gravitic Telescope Units. */
const FD = 12  /* Field Distortion Units. */
const TP = 13  /* Terraforming Plants. */
const GW = 14  /* Germ Warfare Bombs. */
const SG1 = 15 /* Mark-1 Auxiliary Shield Generators. */
const SG2 = 16 /* Mark-2. */
const SG3 = 17 /* Mark-3. */
const SG4 = 18 /* Mark-4. */
const SG5 = 19 /* Mark-5. */
const SG6 = 20 /* Mark-6. */
const SG7 = 21 /* Mark-7. */
const SG8 = 22 /* Mark-8. */
const SG9 = 23 /* Mark-9. */
const GU1 = 24 /* Mark-1 Auxiliary Gun Units. */
const GU2 = 25 /* Mark-2. */
const GU3 = 26 /* Mark-3. */
const GU4 = 27 /* Mark-4. */
const GU5 = 28 /* Mark-5. */
const GU6 = 29 /* Mark-6. */
const GU7 = 30 /* Mark-7. */
const GU8 = 31 /* Mark-8. */
const GU9 = 32 /* Mark-9. */
const X1 = 33  /* Unassigned. */
const X2 = 34  /* Unassigned. */
const X3 = 35  /* Unassigned. */
const X4 = 36  /* Unassigned. */
const X5 = 37  /* Unassigned. */

const MAX_ITEMS = 38 /* Always bump this up to a multiple of two. Don't forget to make room for zeroth element! */

/* Status codes for named planets. These are logically ORed together. */
const HOME_PLANET = 1
const COLONY = 2
const POPULATED = 8
const MINING_COLONY = 16
const RESORT_COLONY = 32
const DISBANDED_COLONY = 64

type nampla_data struct {
	name           [32]char        /* Name of planet. */
	x, y, z, pn    char            /* Coordinates. */
	status         char            /* Status of planet. */
	reserved1      char            /* Zero for now. */
	hiding         char            /* HIDE order given. */
	hidden         char            /* Colony is hidden. */
	reserved2      short           /* Zero for now. */
	planet_index   short           /* Index (starting at zero) into the file "planets.dat" of this planet. */
	siege_eff      short           /* Siege effectiveness - a percentage between 0 and 99. */
	shipyards      short           /* Number of shipyards on planet. */
	reserved4      int             /* Zero for now. */
	IUs_needed     int             /* Incoming ship with only CUs on board. */
	AUs_needed     int             /* Incoming ship with only CUs on board. */
	auto_IUs       int             /* Number of IUs to be automatically installed. */
	auto_AUs       int             /* Number of AUs to be automatically installed. */
	reserved5      int             /* Zero for now. */
	IUs_to_install int             /* Colonial mining units to be installed. */
	AUs_to_install int             /* Colonial manufacturing units to be installed. */
	mi_base        long            /* Mining base times 10. */
	ma_base        long            /* Manufacturing base times 10. */
	pop_units      long            /* Number of available population units. */
	item_quantity  [MAX_ITEMS]long /* Quantity of each item available. */
	reserved6      long            /* Zero for now. */
	use_on_ambush  long            /* Amount to use on ambush. */
	message        long            /* Message associated with this planet, if any. */
	special        long            /* Different for each application. */
	padding        [28]char        /* Use for expansion. Initialized to all zeroes. */
}

/* Ship classes. */
const PB = 0  /* Picketboat. */
const CT = 1  /* Corvette. */
const ES = 2  /* Escort. */
const DD = 3  /* Destroyer. */
const FG = 4  /* Frigate. */
const CL = 5  /* Light Cruiser. */
const CS = 6  /* Strike Cruiser. */
const CA = 7  /* Heavy Cruiser. */
const CC = 8  /* Command Cruiser. */
const BC = 9  /* Battlecruiser. */
const BS = 10 /* Battleship. */
const DN = 11 /* Dreadnought. */
const SD = 12 /* Super Dreadnought. */
const BM = 13 /* Battlemoon. */
const BW = 14 /* Battleworld. */
const BR = 15 /* Battlestar. */
const BA = 16 /* Starbase. */
const TR = 17 /* Transport. */

const NUM_SHIP_CLASSES = 18

/* Ship types. */
const FTL = 0
const SUB_LIGHT = 1
const STARBASE = 2

/* Ship status codes. */
const UNDER_CONSTRUCTION = 0
const ON_SURFACE = 1
const IN_ORBIT = 2
const IN_DEEP_SPACE = 3
const JUMPED_IN_COMBAT = 4
const FORCED_JUMP = 5

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

/* Command codes. */
const UNDEFINED = 0
const ALLY = 1
const AMBUSH = 2
const ATTACK = 3
const AUTO = 4
const BASE = 5
const BATTLE = 6
const BUILD = 7
const CONTINUE = 8
const DEEP = 9
const DESTROY = 10
const DEVELOP = 11
const DISBAND = 12
const END = 13
const ENEMY = 14
const ENGAGE = 15
const ESTIMATE = 16
const HAVEN = 17
const HIDE = 18
const HIJACK = 19
const IBUILD = 20
const ICONTINUE = 21
const INSTALL = 22
const INTERCEPT = 23
const JUMP = 24
const LAND = 25
const MESSAGE = 26
const MOVE = 27
const NAME = 28
const NEUTRAL = 29
const ORBIT = 30
const PJUMP = 31
const PRODUCTION = 32
const RECYCLE = 33
const REPAIR = 34
const RESEARCH = 35
const SCAN = 36
const SEND = 37
const SHIPYARD = 38
const START = 39
const SUMMARY = 40
const SURRENDER = 41
const TARGET = 42
const TEACH = 43
const TECH = 44
const TELESCOPE = 45
const TERRAFORM = 46
const TRANSFER = 47
const UNLOAD = 48
const UPGRADE = 49
const VISITED = 50
const WITHDRAW = 51
const WORMHOLE = 52
const ZZZ = 53

const NUM_COMMANDS = ZZZ + 1

/* Constants needed for parsing. */
const UNKNOWN = 0
const TECH_ID = 1
const ITEM_CLASS = 2
const SHIP_CLASS = 3
const PLANET_ID = 4
const SPECIES_ID = 5

var TechAbbr = []string{"MI", "MA", "ML", "GV", "LS", "BI"}
var TechName = []string{"Mining", "Manufacturing", "Military", "Gravitics", "Life Support", "Biology"}
