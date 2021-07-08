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

import "log"

func (ds *Store) Write(root string) error {
	log.Printf("not saving %q\n", root)
	//filename := filepath.Join(root, "wstore.json")
	//
	//// convert in-memory structures to json-file structures
	//var data struct {
	//	Turn    int                  `json:"turn"`
	//	Species map[string]*wSpecies `json:"species"`
	//	Systems []*wSystem           `json:"systems"`
	//}
	//data.Turn = ds.Turn
	//data.Species = make(map[string]*wSpecies)
	//for _, s := range ds.Sorted.Species {
	//	sp := &wSpecies{
	//		Id:            s.Id,
	//		EconomicUnits: s.EconomicUnits,
	//		TechLevels:    make(map[string]*wTechLevel),
	//	}
	//	for k, v := range s.TechLevels {
	//		val := &wTechLevel{Value: v.Value}
	//		switch k {
	//		case "BI":
	//			sp.TechLevels["biology"] = val
	//		case "GV":
	//			sp.TechLevels["gravitics"] = val
	//		case "LS":
	//			sp.TechLevels["life_support"] = val
	//		case "MA":
	//			sp.TechLevels["manufacturing"] = val
	//		case "MI":
	//			sp.TechLevels["mining"] = val
	//		case "ML":
	//			sp.TechLevels["military"] = val
	//		}
	//	}
	//	data.Species[s.Name] = sp
	//}
	//for _, s := range ds.Sorted.Systems {
	//	//if s.Empty && (len(s.Planets) == 0 && len(s.Ships) == 0) {
	//	//	continue
	//	//}
	//	sys := &wSystem{
	//		Id:      fmt.Sprintf("%d %d %d", s.X, s.Y, s.Z),
	//		Empty:   s.Empty,
	//		Scanned: s.Scanned,
	//		Ships:   nil,
	//		Visited: s.Visited,
	//	}
	//	for _, p := range s.Planets {
	//		pla := &wPlanet{
	//			Orbit:                    p.Orbit,
	//			Name:                     p.Name,
	//			HomeWorld:                p.HomeWorld,
	//			AvailablePopulationUnits: p.AvailablePopulationUnits,
	//			EconomicEfficiency:       p.EconomicEfficiency,
	//			Inventory:                nil,
	//			LSN:                      p.LSN,
	//			MiningDifficulty:         p.MiningDifficulty,
	//			ProductionPenalty:        p.ProductionPenalty,
	//			Ships:                    nil,
	//			Shipyards:                p.Shipyards,
	//		}
	//		sys.Planets = append(sys.Planets, pla)
	//	}
	//	data.Systems = append(data.Systems, sys)
	//}
	//b, err := json.MarshalIndent(&data, "", "  ")
	//if err != nil {
	//	return err
	//} else if err = ioutil.WriteFile(filename, b, 0644); err != nil {
	//	return err
	//}
	//
	//log.Printf("saved %6d systems\n", len(data.Systems))
	return nil
}

type wSpecies struct {
	Id            int                    `json:"id"`
	EconomicUnits int                    `json:"economic_units"`
	TechLevels    map[string]*wTechLevel `json:"tech_levels"`
}
type wTechLevel struct {
	Value int `json:"value"`
}
type wSystem struct {
	Id        string     `json:"id"`
	Empty     bool       `json:"empty,omitempty"`
	Inventory []*wItem   `json:"inventory,omitempty"`
	Planets   []*wPlanet `json:"planets,omitempty"`
	Scanned   int        `json:"scanned,omitempty"`
	Ships     []*wShip   `json:"ships,omitempty"`
	Visited   bool       `json:"visited,omitempty"`
}
type wPlanet struct {
	Orbit                    int      `json:"orbit"`
	Name                     string   `json:"name,omitempty"`
	HomeWorld                bool     `json:"home_world,omitempty"`
	AvailablePopulationUnits int      `json:"available_population_units,omitempty"`
	EconomicEfficiency       int      `json:"economic_efficiency,omitempty"`
	Inventory                []*wItem `json:"inventory,omitempty"`
	LSN                      int      `json:"lsn,omitempty"`
	MiningDifficulty         float64  `json:"mining_difficulty,omitempty"`
	ProductionPenalty        int      `json:"production_penalty,omitempty"`
	Ships                    []*wShip `json:"ships,omitempty"`
	Shipyards                int      `json:"shipyards,omitempty"`
}
type wShip struct {
	Id        string   `json:"id"`
	Landed    bool     `json:"landed"`
	Orbiting  bool     `json:"orbiting"`
	DeepSpace bool     `json:"deep_space"`
	Hiding    bool     `json:"hiding"`
	Inventory []*wItem `json:"inventory,omitempty"`
}
type wItem struct {
	Code     string `json:"code"`
	Quantity int    `json:"qty"`
	Location string `json:"location,omitempty"`
}
