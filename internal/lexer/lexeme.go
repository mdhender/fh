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

package lexer

import "fmt"

type Lexeme struct {
	Line    int
	ArgNo   int
	Kind    Kind
	Integer int
	Text    string
}

type Kind int

const (
	Unknown Kind = iota
	EOF
	EOL
	ERROR
	MISSING
	Ally
	Ambush
	Attack
	Auto
	Base
	Battle
	Build
	Colony
	Combat
	Continue
	Destroy
	Develop
	Disband
	End
	Enemy
	Engage
	Estimate
	Haven
	Hide
	Hijack
	IBuild
	IContinue
	Install
	Integer
	Intercept
	Jump
	Jumps
	Land
	Message
	Move
	Name
	Neutral
	Orbit
	PJump
	PostArrival
	PreDeparture
	Production
	Recycle
	Repair
	Research
	Scan
	Send
	Ship
	Shipyard
	Species
	Start
	Strikes
	SublightShip
	Summary
	Target
	Teach
	Telescope
	Terraform
	Text
	Transfer
	Unload
	Upgrade
	Visited
	Withdraw
	Wormhole
	ZZZ
)

func (k Kind) String() string {
	switch k {
	case Unknown:
		return "Unknown"
	case EOF:
		return "EOF"
	case EOL:
		return "EOL"
	case ERROR:
		return "ERROR"
	case MISSING:
		return "MISSING"
	case Ally:
		return "Ally"
	case Ambush:
		return "Ambush"
	case Attack:
		return "Attack"
	case Auto:
		return "Auto"
	case Base:
		return "Base"
	case Battle:
		return "Battle"
	case Build:
		return "Build"
	case Colony:
		return "Colony"
	case Combat:
		return "Combat"
	case Continue:
		return "Continue"
	case Destroy:
		return "Destroy"
	case Develop:
		return "Develop"
	case Disband:
		return "Disband"
	case End:
		return "End"
	case Enemy:
		return "Enemy"
	case Engage:
		return "Engage"
	case Estimate:
		return "Estimate"
	case Haven:
		return "Haven"
	case Hide:
		return "Hide"
	case Hijack:
		return "Hijack"
	case IBuild:
		return "IBuild"
	case IContinue:
		return "IContinue"
	case Install:
		return "Install"
	case Integer:
		return "Integer"
	case Intercept:
		return "Intercept"
	case Jump:
		return "Jump"
	case Jumps:
		return "Jumps"
	case Land:
		return "Land"
	case Message:
		return "Message"
	case Move:
		return "Move"
	case Name:
		return "Name"
	case Neutral:
		return "Neutral"
	case Orbit:
		return "Orbit"
	case PJump:
		return "PJump"
	case PostArrival:
		return "PostArrival"
	case PreDeparture:
		return "PreDeparture"
	case Production:
		return "Production"
	case Recycle:
		return "Recycle"
	case Repair:
		return "Repair"
	case Research:
		return "Research"
	case Scan:
		return "Scan"
	case Send:
		return "Send"
	case Ship:
		return "Ship"
	case Shipyard:
		return "Shipyard"
	case Species:
		return "Species"
	case Start:
		return "Start"
	case Strikes:
		return "Strikes"
	case SublightShip:
		return "SublightShip"
	case Summary:
		return "Summary"
	case Target:
		return "Target"
	case Teach:
		return "Teach"
	case Telescope:
		return "Telescope"
	case Terraform:
		return "Terraform"
	case Text:
		return "Text"
	case Transfer:
		return "Transfer"
	case Unload:
		return "Unload"
	case Upgrade:
		return "Upgrade"
	case Visited:
		return "Visited"
	case Withdraw:
		return "Withdraw"
	case Wormhole:
		return "Wormhole"
	case ZZZ:
		return "ZZZ"
	}
	return fmt.Sprintf("kind(%d)", k)
}
