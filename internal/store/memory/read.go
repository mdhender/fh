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
	"github.com/mdhender/fh/internal/store/jsondb"
	"log"
)

func (ds *Store) Read(jdb *jsondb.Store) error {
	log.Printf("reading json store\n")
	ds.Version = jdb.Version

	var maxPlanetId int
	for _, planet := range jdb.Planets {
		if maxPlanetId < planet.Id {
			maxPlanetId = planet.Id
		}
	}
	ds.Planets = make([]*Planet, maxPlanetId+1, maxPlanetId+1)

	for _, planet := range jdb.Planets {
		p := &Planet{
			Id:               planet.Id,
			Diameter:         planet.Diameter,
			EconEfficiency:   float64(planet.EconEfficiency) / 100,
			Gases:            make(map[string]int),
			Gravity:          float64(planet.Gravity) / 100,
			Message:          planet.Message,
			MiningDifficulty: float64(planet.MiningDifficulty) / 100,
			PressureClass:    planet.PressureClass,
			TemperatureClass: planet.TemperatureClass,
		}
		var totalGases int
		for gas, percentage := range planet.Gases {
			switch gas {
			case "NH3": // Ammonia
				p.Gases[gas] = percentage
			case "CO2": // Carbon Dioxide
				p.Gases[gas] = percentage
			case "Cl2": // Chlorine
				p.Gases[gas] = percentage
			case "F2": // Fluorine
				p.Gases[gas] = percentage
			case "He": // Helium
				p.Gases[gas] = percentage
			case "H2": // Hydrogen
				p.Gases[gas] = percentage
			case "HCl": // Hydrogen Chloride
				p.Gases[gas] = percentage
			case "H2S": // Hydrogen Sulfide
				p.Gases[gas] = percentage
			case "CH4": // Methane
				p.Gases[gas] = percentage
			case "N2": // Nitrogen
				p.Gases[gas] = percentage
			case "O2": // Oxygen
				p.Gases[gas] = percentage
			case "SO2": // Sulfur Dioxide
				p.Gases[gas] = percentage
			case "H2O": // Water or Steam
				p.Gases[gas] = percentage
			default:
				return fmt.Errorf("unknown gas %q on planet %d", gas, planet.Id)
			}
			if percentage < 1 || percentage > 100 {
				return fmt.Errorf("invalid percentage %d for gas %q on planet %d", percentage, gas, planet.Id)
			}
			totalGases += percentage
		}
		if totalGases > 0 && totalGases != 100 {
			return fmt.Errorf("invalid percentage %d for gases on planet %d", totalGases, planet.Id)
		}
		ds.Planets[p.Id] = p
	}

	ds.Systems = make(map[string]*System)
	for _, system := range jdb.Systems {
		s := System{
			Id:      fmt.Sprintf("%d %d %d", system.Coords.X, system.Coords.Y, system.Coords.Z),
			Coords:  Coords{X: system.Coords.X, Y: system.Coords.Y, Z: system.Coords.Z},
			Planets: make([]*Planet, len(system.Planets)+1, len(system.Planets)+1),
		}
		for o, pIndex := range system.Planets {
			if pIndex < 0 || !(pIndex < len(jdb.Planets)) {
				return fmt.Errorf("invalid planet index %d in system %q", pIndex, system.Id)
			}
			planet := ds.Planets[jdb.Planets[pIndex].Id]
			planet.System = &s
			planet.Coords = Coords{X: s.Coords.X, Y: s.Coords.Y, Z: s.Coords.Z, Orbit: o + 1}
			s.Planets[planet.Coords.Orbit] = planet
		}
		ds.Systems[s.Id] = &s
	}

	var maxSpeciesId int
	for _, species := range jdb.Species {
		if species.Id < 1 {
			return fmt.Errorf("invalid species id %d", species.Id)
		} else if maxSpeciesId < species.Id {
			maxSpeciesId = species.Id
		}
	}
	ds.Species = make([]*Species, maxSpeciesId+1, maxSpeciesId+1)

	for _, species := range jdb.Species {
		sp := Species{
			Id:                  species.Id,
			Name:                species.Name,
			AutoOrders:          species.AutoOrders,
			BankedEconomicUnits: species.BankedEconUnits,
			Relationships:       make(map[int]Relationship),
			Tech:                make(map[string]*Tech),
		}
		for id := 1; id <= maxSpeciesId; id++ {
			sp.Relationships[id] = None
		}
		for id, r := range species.Aliens {
			switch r {
			case "ally":
				sp.Relationships[id] = Ally
			case "enemy":
				sp.Relationships[id] = Enemy
			case "neutral":
				sp.Relationships[id] = Neutral
			default:
				return fmt.Errorf("unknown relationship %q for species %q %d", r, sp.Name, id)
			}
		}
		sp.FleetCost = species.FleetCost
		sp.FleetPercentCost = float64(species.FleetPercentCost) / 100
		sp.Government.Name = species.Government.Name
		sp.Government.Type = species.Government.Type
		sp.Homeworld.OriginalBase = species.HpOriginalBase
		sp.Tech["BI"] = &Tech{
			Level:     species.Tech.Biology.Level,
			Init:      species.Tech.Biology.Init,
			Knowledge: species.Tech.Biology.Knowledge,
			BankedXp:  species.Tech.Biology.BankedXp,
		}
		sp.Tech["GV"] = &Tech{
			Level:     species.Tech.Gravitics.Level,
			Init:      species.Tech.Gravitics.Init,
			Knowledge: species.Tech.Gravitics.Knowledge,
			BankedXp:  species.Tech.Gravitics.BankedXp,
		}
		sp.Tech["LS"] = &Tech{
			Level:     species.Tech.LifeSupport.Level,
			Init:      species.Tech.LifeSupport.Init,
			Knowledge: species.Tech.LifeSupport.Knowledge,
			BankedXp:  species.Tech.LifeSupport.BankedXp,
		}
		sp.Tech["MA"] = &Tech{
			Level:     species.Tech.Manufacturing.Level,
			Init:      species.Tech.Manufacturing.Init,
			Knowledge: species.Tech.Manufacturing.Knowledge,
			BankedXp:  species.Tech.Manufacturing.BankedXp,
		}
		sp.Tech["MI"] = &Tech{
			Level:     species.Tech.Mining.Level,
			Init:      species.Tech.Mining.Init,
			Knowledge: species.Tech.Mining.Knowledge,
			BankedXp:  species.Tech.Mining.BankedXp,
		}
		sp.Tech["ML"] = &Tech{
			Level:     species.Tech.Military.Level,
			Init:      species.Tech.Military.Init,
			Knowledge: species.Tech.Military.Knowledge,
			BankedXp:  species.Tech.Military.BankedXp,
		}
		ds.Species[sp.Id] = &sp
	}

	return nil
}

//func (ds *Store) ORead(root string) error {
//	log.Printf("loading %q\n", root)
//	ds.Planets = make(map[string]*Planet)
//	ds.Species = make(map[string]*Species)
//	ds.Systems = make(map[string]*System)
//
//	filename := filepath.Join(root, "store.json")
//	var data struct {
//		Species map[string]*rSpecies
//		Systems []*rSystem
//		Turn    int
//	}
//	b, err := ioutil.ReadFile(filename)
//	if err != nil {
//		return err
//	} else if err = json.Unmarshal(b, &data); err != nil {
//		return err
//	}
//
//	// inject system coordinates into planets and ships
//	for _, s := range data.Systems {
//		fields := strings.Split(s.Id, " ")
//		if len(fields) != 3 {
//			return fmt.Errorf("system %s: invalid system id", s.Id)
//		}
//		var err error
//		if s.X, err = strconv.Atoi(fields[0]); err != nil {
//			return fmt.Errorf("system %s: invalid x: %+v", s.Id, err)
//		} else if s.Y, err = strconv.Atoi(fields[1]); err != nil {
//			return fmt.Errorf("system %s: invalid y: %+v", s.Id, err)
//		} else if s.Z, err = strconv.Atoi(fields[2]); err != nil {
//			return fmt.Errorf("system %s: invalid z: %+v", s.Id, err)
//		}
//		log.Println(fields, s.X, s.Y, s.Z)
//		for _, p := range s.Planets {
//			p.X, p.Y, p.Z = s.X, s.Y, s.Z
//		}
//	}
//
//	// reject input if we find duplicate systems, planets, or names
//	systems := make(map[string]*rSystem)
//	planets := make(map[string]*rPlanet)
//	species := make(map[int]*rSpecies)
//	names := make(map[string]bool)
//	for name, sp := range data.Species {
//		if _, ok := species[sp.Id]; ok {
//			return fmt.Errorf("species %q: id %d: found species with same id", name, sp.Id)
//		}
//		for k := range sp.TechLevels {
//			switch k {
//			case "biology":
//			case "gravitics":
//			case "life_support":
//			case "manufacturing":
//			case "military":
//			case "mining":
//			default:
//				return fmt.Errorf("species %q: invalid tech %q", name, k)
//			}
//		}
//	}
//	for _, s := range data.Systems {
//		s.Id = fmt.Sprintf("%d %d %d", s.X, s.Y, s.Z)
//		if _, ok := systems[s.Id]; ok {
//			return fmt.Errorf("system %q: found system with same coordinates", s.Id)
//		}
//		systems[s.Id] = s
//
//		for _, p := range s.Planets {
//			p.Id = fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, p.Orbit)
//			if _, ok := planets[p.Id]; ok {
//				return fmt.Errorf("planet %q: found planet with same orbit", p.Id)
//			}
//			planets[p.Id] = p
//			if p.Name != "" {
//				if names[p.Name] {
//					return fmt.Errorf("planet %q: name %q: found another thing with same name", p.Id, p.Name)
//				}
//				planets[p.Name] = p
//			}
//		}
//	}
//
//	// move the data into our in-memory data structures
//	ds.Turn = data.Turn
//	for name, sp := range data.Species {
//		species := &Species{
//			Id:            sp.Id,
//			Name:          name,
//			EconomicUnits: sp.EconomicUnits,
//			TechLevels:    make(map[string]*TechLevel),
//		}
//		for name, tl := range sp.TechLevels {
//			v := &TechLevel{Value: tl.Value}
//			switch name {
//			case "biology":
//				species.TechLevels["BI"] = v
//			case "gravitics":
//				species.TechLevels["GV"] = v
//			case "life_support":
//				species.TechLevels["LS"] = v
//			case "manufacturing":
//				species.TechLevels["MA"] = v
//			case "military":
//				species.TechLevels["ML"] = v
//			case "mining":
//				species.TechLevels["MI"] = v
//			}
//		}
//		ds.Species[species.Name] = species
//		ds.Sorted.Species = append(ds.Sorted.Species, species)
//	}
//	for _, v := range data.Systems {
//		log.Printf("loading system %s\n", v.Id)
//		sys := &System{
//			Id:      v.Id,
//			Empty:   v.Empty,
//			Scanned: v.Scanned,
//			Visited: v.Visited,
//		}
//		sys.X, sys.Y, sys.Z = v.X, v.Y, v.Z
//		if _, ok := ds.Systems[sys.Id]; ok {
//			return fmt.Errorf("system %s: duplicate system", sys.Id)
//		}
//		ds.Systems[sys.Id] = sys
//		ds.Sorted.Systems = append(ds.Sorted.Systems, sys)
//
//		for _, vp := range v.Planets {
//			log.Printf("loading system %q planet %d\n", v.Id, vp.Orbit)
//			pla := &Planet{
//				Id:                       vp.Id,
//				System:                   sys,
//				Orbit:                    vp.Orbit,
//				Name:                     vp.Name,
//				HomeWorld:                vp.HomeWorld,
//				AvailablePopulationUnits: vp.AvailablePopulationUnits,
//				EconomicEfficiency:       vp.EconomicEfficiency,
//				Inventory:                make(map[string]*Item),
//				LSN:                      vp.LSN,
//				MiningDifficulty:         vp.MiningDifficulty,
//				ProductionPenalty:        vp.ProductionPenalty,
//				Shipyards:                vp.Shipyards,
//				Ships:                    make(map[string]*Ship),
//			}
//			if pla.Id == pla.Name {
//				return fmt.Errorf("system %s: planet: %s: id == name", sys.Id, pla.Id)
//			} else if _, ok := ds.Planets[pla.Id]; ok {
//				return fmt.Errorf("system %s: planet: %s: duplicate location", sys.Id, pla.Id)
//			} else if _, ok := ds.Planets[pla.Name]; ok {
//				return fmt.Errorf("system %s: planet: %s: duplicate name %q", sys.Id, pla.Id, pla.Name)
//			}
//			for _, np := range sys.Planets {
//				if np.Orbit == pla.Orbit {
//					return fmt.Errorf("system %s: planet %s: duplicate orbit %d", sys.Id, pla.Id, pla.Orbit)
//				}
//			}
//			for code, item := range vp.Inventory {
//				pla.Inventory[code] = &Item{
//					Code:     code,
//					Location: item.Location,
//					Quantity: item.Quantity,
//				}
//			}
//			for name, ship := range vp.Ships {
//				sh := &Ship{
//					Id:              name,
//					Age:             ship.Age,
//					Capacity:        ship.Capacity,
//					FTL:             !ship.SubLight,
//					Inventory:       make(map[string]*Item),
//					Location:        ship.Location,
//					MaintenanceCost: ship.MaintenanceCost,
//				}
//				if ship.Location == "(C)" {
//					sh.Landed = true
//				} else if ship.Location != "" {
//					loc, orb := ship.Location[:len(ship.Location)-1], ship.Location[len(ship.Location)-1:]
//					if orb != fmt.Sprintf("%d", pla.Orbit) {
//						return fmt.Errorf("system %s: planet %s: ship %q: invalid orbit %q", sys.Id, pla.Id, name, orb)
//					}
//					switch loc {
//					case "D":
//						sh.DeepSpace = true
//					case "FJ":
//						sh.DeepSpace = true
//						sh.ForcedJump = true
//					case "L":
//						sh.Landed = true
//					case "O":
//						sh.Orbiting = true
//					case "WD":
//						sh.DeepSpace = true
//						sh.WithdrewFromCombat = true
//					}
//				}
//				pla.Ships[name] = sh
//			}
//
//			ds.Sorted.Planets = append(ds.Sorted.Planets, pla)
//			sys.Planets = append(sys.Planets, pla)
//			ds.Planets[pla.Id] = pla
//			if pla.Name != "" {
//				ds.Planets[pla.Name] = pla
//			}
//		}
//	}
//
//	for i := 0; i < len(ds.Sorted.Planets); i++ {
//		for k := i + 1; k < len(ds.Sorted.Planets); k++ {
//			if ds.Sorted.Planets[k].Less(ds.Sorted.Planets[i]) {
//				ds.Sorted.Planets[i], ds.Sorted.Planets[k] = ds.Sorted.Planets[k], ds.Sorted.Planets[i]
//			}
//		}
//	}
//	for i := 0; i < len(ds.Sorted.Species); i++ {
//		for k := i + 1; k < len(ds.Sorted.Species); k++ {
//			if ds.Sorted.Species[k].Id < ds.Sorted.Species[i].Id {
//				ds.Sorted.Species[i], ds.Sorted.Species[k] = ds.Sorted.Species[k], ds.Sorted.Species[i]
//			}
//		}
//	}
//	for i := 0; i < len(ds.Sorted.Systems); i++ {
//		for k := i + 1; k < len(ds.Sorted.Systems); k++ {
//			if ds.Sorted.Systems[k].Less(ds.Sorted.Systems[i]) {
//				ds.Sorted.Systems[i], ds.Sorted.Systems[k] = ds.Sorted.Systems[k], ds.Sorted.Systems[i]
//			}
//		}
//	}
//
//	log.Printf("loaded %6d species\n", len(ds.Species))
//	log.Printf("loaded %6d systems\n", len(ds.Systems))
//	log.Printf("loaded %6d planets\n", len(ds.Planets))
//	return nil
//}
//
//type rSystem struct {
//	Id        string     `json:"id"`
//	X         int        `json:"x"`
//	Y         int        `json:"y"`
//	Z         int        `json:"z"`
//	Empty     bool       `json:"empty,omitempty"`
//	Inventory []*rItem   `json:"inventory,omitempty"`
//	Planets   []*rPlanet `json:"planets,omitempty"`
//	Scanned   int        `json:"scanned,omitempty"`
//	Ships     []*rShip   `json:"ships,omitempty"`
//	Visited   bool       `json:"visited,omitempty"`
//}
//
//type rPlanet struct {
//	Id                       string            `json:"id"`
//	X                        int               `json:"-"`
//	Y                        int               `json:"-"`
//	Z                        int               `json:"-"`
//	Orbit                    int               `json:"orbit"`
//	Name                     string            `json:"name,omitempty"`
//	HomeWorld                bool              `json:"home_world,omitempty"`
//	AvailablePopulationUnits int               `json:"available_population_units,omitempty"`
//	EconomicEfficiency       int               `json:"economic_efficiency,omitempty"`
//	Inventory                map[string]*rItem `json:"inventory,omitempty"`
//	LSN                      int               `json:"lsn,omitempty"`
//	MiningDifficulty         float64           `json:"mining_difficulty,omitempty"`
//	ProductionPenalty        int               `json:"production_penalty,omitempty"`
//	Ships                    map[string]*rShip `json:"ships,omitempty"`
//	Shipyards                int               `json:"shipyards,omitempty"`
//}
//
//type rShip struct {
//	Id              string            `json:"id"`
//	Age             int               `json:"age"`
//	Location        string            `json:"location,omitempty"`
//	Capacity        int               `json:"capacity,omitempty"`
//	MaintenanceCost int               `json:"maintenance_cost,omitempty"`
//	SubLight        bool              `json:"sub_light,omitempty"`
//	Inventory       map[string]*rItem `json:"inventory,omitempty"`
//}
//
//type rItem struct {
//	Quantity int    `json:"qty"`
//	Location string `json:"location,omitempty"`
//}
//
//type rSpecies struct {
//	Id            int                    `json:"id"`
//	EconomicUnits int                    `json:"economic_units,omitempty"`
//	TechLevels    map[string]*rTechLevel `json:"tech_levels"`
//}
//
//type rTechLevel struct {
//	Value int `json:"value"`
//}
