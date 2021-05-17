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
)

type PlanetData struct {
	ID               string            `json:"id"`
	Coords           Coords            `json:"coords"`
	TemperatureClass int               `json:"temperature_class"` /* Temperature class, 1-30. */
	PressureClass    int               `json:"pressure_class"`    /* Pressure class, 0-29. */
	Special          PlanetSpecialType `json:"special"`
	Gases            []*GasData        `json:"atmosphere,omitempty"` /* Gas in atmosphere. Nil if none. */
	Diameter         int               `json:"diameter"`             /* Diameter in thousands of kilometers. */
	Density          int               `json:"density"`
	Gravity          int               `json:"gravity"`                              /* Surface gravity. Multiple of Earth gravity times 100. */
	MiningDifficulty int               `json:"mining_difficulty"`                    /* Mining difficulty times 100. */
	EconEfficiency   int               `json:"econ_efficiency"`                      /* Economic efficiency. Always 100 for a home planet. */
	MDIncrease       int               `json:"mining_difficulty_increase,omitempty"` /* Increase in mining difficulty. */
	Message          int               `json:"message_id,omitempty"`                 /* Message associated with this planet, if any. */
}

type GasData struct {
	Type       GasType `json:"type"`
	Percentage int     `json:"pct"`
}

// Values for the planets of Earth's solar system will be used as starting values.
// Diameters are in thousands of kilometers.
// The zeroth element of each array is a placeholder and is not used.
// The fifth element corresponds to the asteroid belt, and is pure fantasy on my part.
// I omitted Pluto because it is probably a captured planet, rather than an original member of our solar system.
var earth = []struct{ diameter, temperatureClass int }{
	{0, 0},   // unused
	{5, 29},  // Mercury
	{12, 27}, // Venus
	{13, 11}, // Earth
	{7, 9},   // Mars
	{20, 8},  // Asteroid Belt
	{143, 6}, // Jupiter
	{121, 5}, // Saturn
	{51, 5},  // Uranus
	{49, 3},  // Neptune
}

// GenerateEarthLikePlanet will try to random generate a set of planets
// that contains one Earth-like planet. If it can't, it will return nil.
func GenerateEarthLikePlanet(starId string, num_planets int) []*PlanetData {
	// set flag to indicate this star system requires an earth-like planet.
	// We will reset it after we have created one.
	make_earth := true

	var planets []*PlanetData
	var homePlanet *PlanetData

	/* Main loop. Generate one planet at a time. */
	for planet_number := 1; planet_number <= num_planets; planet_number++ {
		planet := &PlanetData{ID: fmt.Sprintf("%s-%02d", starId, planet_number)}
		planets = append(planets, planet)

		/* Start with diameters, temperature classes and pressure classes based on the planets in Earth's solar system. */
		var startOffset int
		if num_planets <= 3 {
			startOffset = 2*planet_number + 1
		} else {
			startOffset = (9 * planet_number) / num_planets
		}
		planet.Diameter = planet.GenerateDiameter(earth[startOffset].diameter)
		planet.TemperatureClass = earth[startOffset].temperatureClass

		/* If diameter is greater than 40,000 km, assume the planet is a gas giant. */
		gas_giant := (planet.Diameter > 40)

		planet.Density = planet.GenerateDensity(gas_giant)

		// Gravitational acceleration is proportional to the mass divided by the radius-squared.
		// The radius is proportional to the diameter, and the mass is proportional to the density times the radius-cubed.
		// The net result is that "g" is proportional to the density times the diameter.
		// Our value for "g" will be a multiple of Earth gravity, and will be further multiplied by 100 to allow us to use integer arithmetic.
		// The factor 72 ensures that "g" will be 100 for Earth (density=550 and diameter=13).
		planet.Gravity = (planet.Density * planet.Diameter) / 72

		planet.TemperatureClass = planet.GenerateTemperatureClass(num_planets, planet_number, gas_giant, earth[startOffset].temperatureClass)
		/* Make sure that planets farther from the sun are not warmer than planets closer to the sun. */
		if planet_number > 1 && planets[planet_number-1].TemperatureClass < planet.TemperatureClass {
			planet.TemperatureClass = planets[planet_number-1].TemperatureClass - (rnd(3) - 1)
			if planet.TemperatureClass < 1 {
				planet.TemperatureClass = 1
			}
		}

		// Check if this planet should be earth-like.
		// If so, discard all of the above and replace with earth-like characteristics.
		if make_earth && homePlanet == nil && planet.TemperatureClass <= 11 {
			homePlanet, make_earth = planet, false /* Once only. */

			planet.Diameter = 11 + rnd(3)
			planet.Gravity = 93 + rnd(11) + rnd(11) + rnd(5)
			planet.TemperatureClass = 9 + rnd(3)
			planet.PressureClass = 8 + rnd(3)
			planet.MiningDifficulty = 208 + rnd(11) + rnd(11)
			planet.Special = IDEAL_HOME_PLANET /* Maybe ideal home planet. */

			pctRemaining := 100
			if rnd(3) == 1 {
				/* Give it a shot of ammonia. */
				gas := &GasData{NH3, rnd(30)}
				planet.Gases = append(planet.Gases, gas)
				pctRemaining -= gas.Percentage
			}

			if rnd(3) == 1 {
				/* Give it a shot of carbon dioxide. */
				gas := &GasData{CO2, rnd(30)}
				planet.Gases = append(planet.Gases, gas)
				pctRemaining -= gas.Percentage
			}

			/* Now do oxygen. */
			gas := &GasData{O2, rnd(20) + 10}
			planet.Gases = append(planet.Gases, gas)
			pctRemaining -= gas.Percentage

			/* Give the rest to nitrogen. */
			gas = &GasData{N2, pctRemaining}
			planet.Gases = append(planet.Gases, gas)

			continue
		}

		/* Pressure class depends primarily on gravity. Calculate an approximate value and randomize it. */
		planet.PressureClass = planet.GeneratePressureClass(planet.Gravity, planet.TemperatureClass, gas_giant)

		/* Generate gases, if any, in the atmosphere. */
		for _, gas := range planet.GenerateGases(planet.PressureClass, planet.TemperatureClass) {
			planet.Gases = append(planet.Gases, gas)
		}

		// Get mining difficulty.
		planet.MiningDifficulty = planet.GenerateMiningDifficulty(planet.Diameter, true)
	}

	if homePlanet == nil {
		//fmt.Printf("found no home planet\n")
		return nil
	}

	// If this is a potential home system, make sure it passes certain tests.
	// What this test is, I do not know.
	potential := 0
	for _, planet := range planets {
		potential += 20000 / ((LSN(planet, homePlanet) + 3) * (50 + planet.MiningDifficulty))
	}
	if potential < 54 || potential > 56 {
		//fmt.Printf("home planet potential %d did not pass certain tests\n", potential)
		return nil
	}

	return planets
}

func GeneratePlanet(starId string, coords Coords, num_planets int) ([]*PlanetData, error) {
	var planets []*PlanetData

	/* Main loop. Generate one planet at a time. */
	for planet_number := 1; planet_number <= num_planets; planet_number++ {
		planet := &PlanetData{
			ID:     fmt.Sprintf("%s-%02d", starId, planet_number),
			Coords: Coords{X: coords.X, Y: coords.Y, Z: coords.Z, Orbit: planet_number},
		}
		planets = append(planets, planet)

		/* Start with diameters, temperature classes and pressure classes based on the planets in Earth's solar system. */
		var startOffset int
		if num_planets <= 3 {
			startOffset = 2*planet_number + 1
		} else {
			startOffset = (9 * planet_number) / num_planets
		}
		planet.Diameter = planet.GenerateDiameter(earth[startOffset].diameter)
		planet.TemperatureClass = earth[startOffset].temperatureClass

		/* If diameter is greater than 40,000 km, assume the planet is a gas giant. */
		gas_giant := (planet.Diameter > 40)

		planet.Density = planet.GenerateDensity(gas_giant)

		// Gravitational acceleration is proportional to the mass divided by the radius-squared.
		// The radius is proportional to the diameter, and the mass is proportional to the density times the radius-cubed.
		// The net result is that "g" is proportional to the density times the diameter.
		// Our value for "g" will be a multiple of Earth gravity, and will be further multiplied by 100 to allow us to use integer arithmetic.
		// The factor 72 ensures that "g" will be 100 for Earth (density=550 and diameter=13).
		planet.Gravity = (planet.Density * planet.Diameter) / 72

		planet.TemperatureClass = planet.GenerateTemperatureClass(num_planets, planet_number, gas_giant, earth[startOffset].temperatureClass)
		/* Make sure that planets farther from the sun are not warmer than planets closer to the sun. */
		if planet_number > 1 && planets[planet_number-1].TemperatureClass < planet.TemperatureClass {
			planet.TemperatureClass = planets[planet_number-1].TemperatureClass - (rnd(3) - 1)
			if planet.TemperatureClass < 1 {
				planet.TemperatureClass = 1
			}
		}

		/* Pressure class depends primarily on gravity. Calculate an approximate value and randomize it. */
		planet.PressureClass = planet.GeneratePressureClass(planet.Gravity, planet.TemperatureClass, gas_giant)

		/* Generate gases, if any, in the atmosphere. */
		for _, gas := range planet.GenerateGases(planet.PressureClass, planet.TemperatureClass) {
			planet.Gases = append(planet.Gases, gas)
		}

		// Get mining difficulty.
		planet.MiningDifficulty = planet.GenerateMiningDifficulty(planet.Diameter, false)
	}

	return planets, nil
}

func (p *PlanetData) Clone() *PlanetData {
	var gases []*GasData
	for _, gas := range p.Gases {
		gases = append(gases, &GasData{
			Type:       gas.Type,
			Percentage: gas.Percentage,
		})
	}
	return &PlanetData{
		Coords:           p.Coords,
		TemperatureClass: p.TemperatureClass,
		PressureClass:    p.PressureClass,
		Special:          p.Special,
		Gases:            gases,
		Diameter:         p.Diameter,
		Density:          p.Density,
		Gravity:          p.Gravity,
		MiningDifficulty: p.MiningDifficulty,
		EconEfficiency:   p.EconEfficiency,
		MDIncrease:       p.MDIncrease,
		Message:          p.Message,
	}
}

// GenerateDensity
// Density will depend on whether or not the planet is a gas giant.
// Again ignoring Pluto, densities range from 0.7 to 1.6 times the density of water for the gas giants, and from 3.9 to 5.5 for the others.
// We will expand this range slightly and use 100 times the actual density so that we can use integer arithmetic.
func (p *PlanetData) GenerateDensity(gasGiant bool) int {
	var base, sigma int
	if gasGiant {
		/* Final values from 60 thru 170. */
		base, sigma = 58, 56
	} else {
		/* Final values from 370 thru 570. */
		base, sigma = 368, 101
	}
	return base + rnd(sigma) + rnd(sigma)
}

// GenerateDiameter
func (p *PlanetData) GenerateDiameter(baseDiameter int) int {
	diameter, die_size := baseDiameter, baseDiameter/4
	if die_size < 2 {
		die_size = 2
	}
	for i := 1; i <= 4; i++ {
		if rnd(100) > 50 {
			diameter += rnd(die_size)
		} else {
			diameter -= rnd(die_size)
		}
	}

	// Minimum allowable diameter is 3,000 km.
	// Note that the maximum diameter we can generate is 283,000 km.
	for diameter < 3 {
		diameter += rnd(4)
	}

	return diameter
}

// GenerateGases
/* Generate gases, if any, in the atmosphere. */
func (p *PlanetData) GenerateGases(pressureClass, temperatureClass int) []*GasData {
	if pressureClass == 0 {
		// no atmosphere, no gases
		return nil
	}
	var gases []*GasData

	// Convert planet's temperature class to a value between 1 and 9.
	// We will use it as the start index into the list of 13 potential gases.
	var firstGas GasType
	switch 100 * temperatureClass / 225 {
	case 0:
		firstGas = H2 /* Hydrogen */
	case 1:
		firstGas = H2 /* Hydrogen */
	case 2:
		firstGas = CH4 /* Methane */
	case 3:
		firstGas = HE /* Helium */
	case 4:
		firstGas = NH3 /* Ammonia */
	case 5:
		firstGas = N2 /* Nitrogen */
	case 6:
		firstGas = CO2 /* Carbon Dioxide */
	case 7:
		firstGas = O2 /* Oxygen */
	case 8:
		firstGas = HCL /* Hydrogen Chloride */
	default:
		firstGas = CL2 /* Chlorine */
	}

	/* The following algorithm is something I tweaked until it worked well. */
	num_gases_wanted := (rnd(4) + rnd(4)) / 2
	for len(gases) == 0 {
		for i := firstGas; i <= firstGas+4 && len(gases) < num_gases_wanted; i++ {
			if i == HE && temperatureClass > 5 {
				// too hot for Helium
				continue
			}
			// skip to the next gas about one-third of the time
			// (unless we're on Helium, then it's two-thirds of the time)
			switch rnd(3) {
			case 2:
				if i == HE {
					/* Don't want too many Helium planets. */
					continue
				}
			case 3:
				continue
			}

			gas := &GasData{Type: i}
			switch i {
			case HE:
				// Helium is self-limiting
				gas.Percentage = rnd(20)
			case O2:
				// Oxygen is self-limiting
				gas.Percentage = rnd(50)
			default:
				gas.Percentage = rnd(100)
			}
			gases = append(gases, gas)
		}
	}

	// determine total quantity of gases in the atmosphere
	var total_quantity int
	for _, gas := range gases {
		total_quantity += gas.Percentage
	}
	// convert gas quantities to percentages
	var total_percent int
	for _, gas := range gases {
		gas.Percentage = 100 * gas.Percentage / total_quantity
		total_percent += gas.Percentage
	}

	// give leftover to first gas
	gases[0].Percentage += 100 - total_percent

	return gases
}

// GenerateMiningDifficulty
// Basically, mining difficulty is proportional to planetary diameter with randomization and an occasional big surprise.
// Earth-like values will range between 0.30 and 10.00.
// Non earth-like values will range between 0.80 and 10.00.
// Again, the actual value will be multiplied by 100 to allow use of integer arithmetic.
func (p *PlanetData) GenerateMiningDifficulty(diameter int, earthLike bool) int {
	mining_dif := 0
	for mining_dif < 40 || mining_dif > 500 {
		mining_dif = (rnd(3)+rnd(3)+rnd(3)-rnd(4))*rnd(diameter) + rnd(30) + rnd(30)
	}

	if earthLike {
		for mining_dif < 30 || mining_dif > 1000 {
			mining_dif = (rnd(3)+rnd(3)+rnd(3)-rnd(4))*rnd(diameter) + rnd(20) + rnd(20)
		}
	} else {
		for mining_dif < 40 || mining_dif > 500 {
			mining_dif = (rnd(3)+rnd(3)+rnd(3)-rnd(4))*rnd(diameter) + rnd(30) + rnd(30)
		}
		mining_dif = (mining_dif * 11) / 5 /* Fudge factor. */
	}

	return mining_dif
}

// GeneratePressureClass
func (p *PlanetData) GeneratePressureClass(gravity, temperatureClass int, gasGiant bool) int {
	if gravity < 10 {
		// gravity is too low to retain an atmosphere
		return 0
	} else if temperatureClass < 2 || temperatureClass > 27 {
		// Planets outside this temperature range have no atmosphere
		return 0
	}

	pressureClass := gravity / 10
	die_size := pressureClass / 4
	if die_size < 2 {
		die_size = 2
	}
	for i, nRolls := 1, rnd(3)+rnd(3)+rnd(3); i <= nRolls; i++ {
		if rnd(100) > 50 {
			pressureClass += rnd(die_size)
		} else {
			pressureClass -= rnd(die_size)
		}
	}

	if gasGiant {
		for pressureClass < 11 {
			pressureClass += rnd(3)
		}
		for pressureClass > 29 {
			pressureClass -= rnd(3)
		}
	} else {
		for pressureClass < 0 {
			pressureClass += rnd(3)
		}
		for pressureClass > 12 {
			pressureClass -= rnd(3)
		}
	}

	return pressureClass
}

// GenerateTemperatureClass
func (p *PlanetData) GenerateTemperatureClass(numPlanets, orbit int, gasGiant bool, baseTemperatureClass int) int {
	/* Randomize the temperature class obtained earlier. */
	temperatureClass, die_size := baseTemperatureClass, baseTemperatureClass/4
	if die_size < 2 {
		die_size = 2
	}
	n_rolls := rnd(3) + rnd(3) + rnd(3)
	for i := 1; i <= n_rolls; i++ {
		if rnd(100) > 50 {
			temperatureClass += rnd(die_size)
		} else {
			temperatureClass -= rnd(die_size)
		}
	}

	if gasGiant {
		for temperatureClass < 3 {
			temperatureClass += rnd(2)
		}
		for temperatureClass > 7 {
			temperatureClass -= rnd(2)
		}
	} else {
		for temperatureClass < 1 {
			temperatureClass += rnd(3)
		}
		for temperatureClass > 30 {
			temperatureClass -= rnd(3)
		}
	}

	// Sometimes, an inner planet in star systems with less than four planets are too cold.
	// Warm them up a little.
	if numPlanets < 4 && orbit < 3 {
		for temperatureClass < 12 {
			temperatureClass += rnd(4)
		}
	}

	return temperatureClass
}

// LSN provides an approximate LSN (Life Support Needed) for a planet.
// It assumes that oxygen is required and any gas that does not appear on the home planet is poisonous.
func LSN(current_planet, home_planet *PlanetData) int {
	ls_needed := 0
	// need 2 points of life support for every point difference in Temperature class.
	if current_planet.TemperatureClass < home_planet.TemperatureClass {
		ls_needed += 2*home_planet.TemperatureClass - current_planet.TemperatureClass
	} else if current_planet.TemperatureClass > home_planet.TemperatureClass {
		ls_needed += 2*current_planet.TemperatureClass - home_planet.TemperatureClass
	}

	// need 2 points of life support for every point difference in Pressure class.
	if current_planet.PressureClass < home_planet.PressureClass {
		ls_needed += 2*home_planet.PressureClass - current_planet.PressureClass
	} else if current_planet.PressureClass > home_planet.PressureClass {
		ls_needed += 2*current_planet.PressureClass - home_planet.PressureClass
	}

	// check for oxygen and any poisonous gases
	var hasOxygen bool
	for _, gas := range current_planet.Gases {
		if gas.Type == O2 {
			hasOxygen = true
		}
	}
	if !hasOxygen {
		ls_needed += 2
	}

	// check for poisonous gases
	for _, gas := range current_planet.Gases {
		// not poisonous if found on home planet
		isPoison := true
		for _, hg := range home_planet.Gases {
			if hg.Type == gas.Type {
				isPoison = false
			}
		}
		if isPoison {
			ls_needed += 2
		}
	}

	return ls_needed
}
