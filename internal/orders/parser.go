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

package orders

import (
	"io/ioutil"
)

func Parse(name string) *Orders {
	orders := &Orders{}
	b, err := ioutil.ReadFile(name)
	if err != nil {
		orders.Errors = append(orders.Errors, err)
		return orders
	}

	var commands []*Command
	var command *Command
	scanner := NewScanner(b)
	for tk := scanner.Next(); tk != nil; tk = scanner.Next() {
		switch tk.Text {
		case "\n":
			if command != nil {
				commands = append(commands, command)
			}
			command = nil
		default:
			if command == nil {
				command = &Command{Line: tk.Line, Name: tk.Text}
			} else {
				command.Args = append(command.Args, tk.Text)
			}
		}
	}
	if command != nil {
		commands = append(commands, command)
	}
	//for _, command := range commands {
	//	fmt.Println(*command)
	//}
	var section *Section
	for _, command := range commands {
		switch command.Name {
		case "START":
			var name string
			if len(command.Args) != 0 {
				name = command.Args[0]
			}
			switch name {
			case "COMBAT":
				if orders.Combat == nil {
					orders.Combat = &Section{Line: command.Line, Name: name}
				}
				section = orders.Combat
			case "JUMPS":
				if orders.Jumps == nil {
					orders.Jumps = &Section{Line: command.Line, Name: name}
				}
				section = orders.Jumps
			case "POST-ARRIVAL":
				if orders.PostArrival == nil {
					orders.PostArrival = &Section{Line: command.Line, Name: name}
				}
				section = orders.PostArrival
			case "PRE-DEPARTURE":
				if orders.PreDeparture == nil {
					orders.PreDeparture = &Section{Line: command.Line, Name: name}
				}
				section = orders.PreDeparture
			case "PRODUCTION":
				if orders.Production == nil {
					orders.Production = &Section{Line: command.Line, Name: name}
				}
				section = orders.Production
			case "STRIKES":
				if orders.Strikes == nil {
					orders.Strikes = &Section{Line: command.Line, Name: name}
				}
				section = orders.Strikes
			}
		case "END":
		default:
			if section != nil {
				section.Commands = append(section.Commands, command)
			}
		}
	}

	return orders
}
