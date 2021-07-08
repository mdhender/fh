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
	"errors"
)

type Store struct {
	Version    string
	TurnNumber int

	Systems map[string]*System // indexed by systemId
	Planets []*Planet          // indexed by planetId
	Species []*Species         // indexed by spId

	Colonies map[string]*Colony // colonies are "named planets"
	Ships    map[string]*Ship   // key for ship is spId / shipId
}

var ErrInternalError = errors.New("internal error")
var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

func (ds *Store) GetVersion() (string, error) {
	if ds == nil {
		return "?", ErrInternalError
	}
	return ds.Version, nil
}
