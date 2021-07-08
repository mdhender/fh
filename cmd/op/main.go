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

package main

import (
	"fmt"
	"github.com/mdhender/fh/internal/parser"
	"github.com/mdhender/fh/internal/scanner"
	"io/ioutil"
)

func main() {
	for _, err := range run("D:/GoLand/fh/testdata/sp18.ord.txt") {
		fmt.Printf("%+v\n", err)
	}
	fmt.Println("parsed!")
}

func run(name string) []error {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return []error{nil}
	}
	tokenizer := scanner.NewTokenizer(b)
	for tk := tokenizer.Next(); tk != nil; tk = tokenizer.Next() {
		fmt.Printf("[token] %3d %2d %q\n", tk.Line, tk.Col, tk.Text)
	}
	//l, err := lexer.Lex(name)
	//if err != nil {
	//	return []error{err}
	//}
	//for lexeme := l.Next(); lexeme.Kind != lexer.EOF; lexeme = l.Next() {
	//	if lexeme.ArgNo == 1 {
	//		fmt.Printf("line %4d/%2d: %-12s %q\n", lexeme.Line, lexeme.ArgNo, lexeme.Kind, lexeme.Text)
	//	} else {
	//		fmt.Printf("         /%2d: %-12s %q\n", lexeme.ArgNo, lexeme.Kind, lexeme.Text)
	//	}
	//}
	//
	_, errors := parser.Parse(name)
	if errors != nil {
		return errors
	}
	return nil
}
