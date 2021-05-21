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

// Package prng implements a PRNG to return a random int between 1 and max, inclusive.
// It uses the so-called "Algorithm M" method, which is a combination of the congruential and shift-register methods.
package prng

import "time"

type PRNG struct {
	seed uint64
}

var defaultPRNG = &PRNG{seed: 0xBADC0FFEE}

func New(seed uint64) *PRNG {
	return &PRNG{seed: seed}
}

func (p *PRNG) Roll(max int) int {
	// for congruential method, multiply previous value by the prime number 16417.
	a, b, c := p.seed, p.seed<<5, p.seed<<14
	congResult := a + b + c // effectively multiply by 16417

	// for shift-register method, use shift-right 15 and shift-left 17 with no-carry addition (i.e., exclusive-or)
	a = p.seed >> 15
	shiftResult := a ^ p.seed
	a = shiftResult << 17
	shiftResult ^= a

	p.seed = congResult ^ shiftResult

	// avoid returning the low-order bits
	return int(((p.seed&0x0000FFFF)*uint64(max))>>16) + 1
}

// Seed random number generator
func (p *PRNG) Seed(seed uint64) {
	p.seed = seed
	n := p.Roll(100) + p.Roll(200) + p.Roll(300)
	for i := 0; i < n; i++ {
		p.Roll(10)
	}
}

// SeedFromTime seeds the generator from the system clock.
func (p *PRNG) SeedFromTime() {
	p.Seed(uint64(time.Now().UnixNano()))
}

// Roll returns a random number in the range 1..n.
func Roll(max int) int {
	return defaultPRNG.Roll(max)
}

// Seed random number generator
func Seed(seed string) {
	var hash uint64
	if seed == "0xBADC0FFEE" || seed == "0" || len(seed) == 0 {
		hash = 0xBADC0FFEE
	} else {
		hash = 5381
		for _, ch := range []byte(seed) {
			hash = (hash * 33) ^ uint64(ch)
		}
	}
	defaultPRNG.Seed(hash)
}

// SeedFromTime seeds the generator from the system clock.
func SeedFromTime() {
	defaultPRNG.SeedFromTime()
}
