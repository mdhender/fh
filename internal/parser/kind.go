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

type Node struct {
	Kind     Kind
	Children []*Node
}

type Kind int

const (
	ERROR Kind = iota
	ALLY
	AMBUSH
	ATTACK
	AUTO
	BASE
	BATTLE
	BUILD
	COMBAT
	CONTINUE
	DESTROY
	DEVELOP
	DISBAND
	END
	ENEMY
	ENGAGE
	ESTIMATE
	HAVEN
	HIDE
	HIJACK
	IBUILD
	ICONTINUE
	INSTALL
	INTERCEPT
	JUMP
	JUMPS
	LAND
	MESSAGE
	MOVE
	NAME
	NEUTRAL
	ORBIT
	PJUMP
	POSTARRIVAL
	PREDEPARTURE
	PRODUCTION
	RECYCLE
	REPAIR
	RESEARCH
	SCAN
	SEND
	SHIPYARD
	START
	STRIKES
	SUMMARY
	TARGET
	TEACH
	TELESCOPE
	TERRAFORM
	TRANSFER
	UNLOAD
	UPGRADE
	VISITED
	WITHDRAW
	WORMHOLE
	ZZZ
)
