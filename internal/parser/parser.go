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

package parser

import (
	"fmt"
	"github.com/mdhender/fh/internal/scanner"
	"io/ioutil"
)

type Section struct {
	Line int
	Name string
	Commands []*Command
}
type Command struct {
	Line int
	Name string
	Args []string
}
func Parse(name string) (root []*Section, errors []error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, []error{nil}
	}
	var commands []*Command
	var command *Command
	tokenizer := scanner.NewTokenizer(b)
	for tk := tokenizer.Next(); tk != nil; tk = tokenizer.Next() {
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
	for _, command := range commands {
		fmt.Println(*command)
	}
	var sections []*Section
	for _, command := range commands {
		switch command.Name {
		case "START":
			var name string
			if len(command.Args) != 0 {
				name = command.Args[0]
			}
			sections = append(sections, &Section{Line: command.Line, Name: name})
		case "END":
		default:
			if len(sections) != 0 {
				sections[len(sections)-1].Commands = append(sections[len(sections)-1].Commands, command)
			}
		}
	}
	for _, section := range sections {
		fmt.Printf("START %q\n", section.Name)
		for _, command := range section.Commands {
			fmt.Printf("  %-12s", command.Name)
			for _, arg := range command.Args {
				fmt.Printf(" %q", arg)
			}
			fmt.Printf("\n")
		}
	}
	return sections, errors
}

	//section := p.scanner.Next()
	//switch section.Kind {
	//case lexer.Combat:
	//case lexer.Jumps:
	//case lexer.PostArrival:
	//case lexer.PreDeparture:
	//case lexer.Production:
	//case lexer.Strikes:
	//
	//switch command.Kind {
	//case lexer.Ally:
	//case lexer.Ambush:
	//case lexer.Attack:
	//case lexer.Auto:
	//case lexer.Base:
	//case lexer.Battle:
	//case lexer.Build:
	//case lexer.Continue:
	//case lexer.Destroy:
	//case lexer.Develop:
	//case lexer.Disband:
	//case lexer.Enemy:
	//case lexer.Engage:
	//case lexer.Estimate:
	//case lexer.Haven:
	//case lexer.Hide:
	//case lexer.Hijack:
	//case lexer.IBuild:
	//case lexer.IContinue:
	//case lexer.Install:
	//case lexer.Intercept:
	//case lexer.Jump:
	//case lexer.Land:
	//case lexer.Message:
	//case lexer.Move:
	//case lexer.Name:
	//case lexer.Neutral:
	//case lexer.Orbit:
	//case lexer.PJump:
	//case lexer.Production:
	//case lexer.Recycle:
	//case lexer.Repair:
	//case lexer.Research:
	//case lexer.Scan:
	//case lexer.Send:
	//case lexer.Shipyard:
	//case lexer.Summary:
	//case lexer.Target:
	//case lexer.Teach:
	//case lexer.Telescope:
	//case lexer.Terraform:
	//case lexer.Transfer:
	//case lexer.Unload:
	//	fmt.Println(command)
	//case lexer.Upgrade:
	//case lexer.Visited:
	//case lexer.Withdraw:
	//case lexer.Wormhole:
	//case lexer.ZZZ:
