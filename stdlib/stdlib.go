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

package stdlib

import (
	"fmt"
	"strings"
	"unicode"
)

func IsValidName(name string) error {
	if name != strings.TrimSpace(name) {
		return fmt.Errorf("name can't have leading or trailing spaces")
	} else if name == "" {
		return fmt.Errorf("name can't be blank")
	}
	for _, ch := range name {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || unicode.IsSpace(ch) || ch == '.' || ch == '\'' || ch == '-') {
			return fmt.Errorf("invalid character %q in name", string(ch))
		}
	}
	return nil
}
