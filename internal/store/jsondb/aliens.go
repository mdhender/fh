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

package jsondb

import (
	"bytes"
	"fmt"
)

type AlienRelationship int

const (
	ARNone AlienRelationship = iota
	ARAlly
	AREnemy
	ARNeutral
)

func (ar AlienRelationship) MarshalJSON() ([]byte, error) {
	switch ar {
	case ARNone:
		return []byte("none"), nil
	case ARAlly:
		return []byte("ally"), nil
	case AREnemy:
		return []byte("enemy"), nil
	case ARNeutral:
		return []byte("neutral"), nil
	}
	panic(fmt.Sprintf("assert(alienRelationship != %d)", ar))
}

func (ar *AlienRelationship) UnmarshalJSON(b []byte) error {
	*ar = ARNone
	if bytes.Equal(b, []byte(`"ally"`)) {
		*ar = ARAlly
	} else if bytes.Equal(b, []byte(`"enemy"`)) {
		*ar = AREnemy
	} else if bytes.Equal(b, []byte(`"neutral"`)) {
		*ar = ARNeutral
	}
	return nil
}
