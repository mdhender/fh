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

package order

type Order interface {
	Process(args ...string)
}

// ALLY declare species "sp" to be an ally
type ALLY struct {
	SP int // species ID
}

// AMBUSH spend "n" in preparation for ambush
type AMBUSH struct {
	N int
}

// ATTACK attack opponent "sp"
// or field-distorted species number "n"
// or attack all declared enemies if SP and N are both 0
type ATTACK struct {
	SP int // species ID
	N  int // field-distorted species number
}

// AUTO automatically generate sensible orders for next turn
type AUTO struct{}

// BASE build or increase size of starbase "base" using "n" starbase units from "s"
type BASE struct {
	N    int // optional, may be zero
	S    int
	Base int
}

// BATTLE set the location for a battle
type BATTLE struct {
	X, Y, Z int
}

//|BUILD           |n ab            |Build "n" items of class "ab"
//|BUILD           |ship            |Build "ship"
//|BUILD           |ship,n          |Start building "ship", spend only "n"
//|BUILD           |base,n          |Start building starbase "base", spend "n"
//|CONTINUE        |ship            |Finish construction of "ship"
//|CONTINUE        |ship,n          |Continue construction on "ship", spend only "n"
//|CONTINUE        |base,n          |Increase size of starbase "base", spend "n"
//|DESTROY         |ship            |Destroy "ship"
//|DESTROY         |base            |Destroy starbase "base"
//|DEVELOP         |[n]             |Build CUs, IUs, and AUs for producing planet but do not spend more than "n"
//|DEVELOP         |[n] pl          |Build CUs, IUs, and AUs for colony planet "pl" in same sector but do not spend more than "n"
//|DEVELOP         |[n] pl, ship    |Build CUs, IUs, and AUs for colony planet "pl" and load units onto "ship" but do not spend more than "n"
//|DISBAND         |pl              |Disband colony "pl"
//|END             |                |End current section of the order form
//|ENEMY           |sp              |Declare species "sp" to be an enemy
//|ENEMY           |n               |Declare all species to be enemies
//|ENGAGE          |n [p]           |Specify combat engagement option "n" and optional planet number "p"
//|ESTIMATE        |sp              |Estimate tech levels of species "sp"
//|HAVEN           |x y x           |Set rendezvous point for ships that withdraw from combat
//|HIDE            |                |Actively hide this planet from alien observation
//|HIDE            |ship            |Keep "ship" out of combat unless you start to lose the battle
//|HIJACK          |sp              |Hijack opponent "sp"
//|HIJACK          |SP n            |Hijack field-distorted species number "n"
//|HIJACK          |0               |Hijack all declared enemies
//|IBUILD          |sp,n ab         |Build "n" items of class "ab" for species "sp"
//|IBUILD          |sp,ship         |Build "ship" for species "sp"
//|IBUILD          |sp,base,n       |Build starbase "base" for species "sp", spend "n"
//|ICONTINUE       |sp ship         |Finish construction of "ship" for species "sp"
//|ICONTINUE       |sp base,n       |Increase size of starbase "base" for species "sp", spend "n"
//|INSTALL         |n ab pl         |Install "n" IUs or AUs on planet "pl"
//|INSTALL         |pl              |Install all available IUs and AUs on planet "pl"
//|INTERCEPT       |n               |Spend "n" in preparation for interception
//|JUMP            |ship,loc        |Have "ship" jump to destination "loc"
//|LAND            |ship,pl         |Have "ship" land on planet in same star system
//|MESSAGE         |sp              |Send a message to species "sp"
//|MOVE            |ship, x y z     |Move "ship" up to one parsec
//|MOVE            |base, x y z     |Tow starbase "base" up to one parsec
//|NAME            |x y z p PL name |Give "name" to planet "p" at location "x y z"
//|NEUTRAL         |sp              |Declare neutrality towards species "sp"
//|NEUTRAL         |n               |Declare neutrality towards all species
//|ORBIT           |ship,pl         |Have "ship" orbit planet in same star system
//|PJUMP           |ship,loc,bas    |Have "ship" jump to destination "loc" via jump portals on starbase "bas"
//|PRODUCTION      |PL name         |Start production on planet "name"
//|RECYCLE         |n ab            |Recycle "n" items of class "ab"
//|RECYCLE         |ship            |Recycle "ship"
//|RECYCLE         |base            |Recycle starbase "base"
//|REPAIR          |ship,n          |Repair "ship" using "n" onboard damage repair units
//|REPAIR          |base,n          |Repair "base" using "n" onboard damage repair units
//|REPAIR          |x y z [age]     |Repair as many ships/starbases as possible in sector x y z, pooling damage repair units but do not reduce age below "age"
//|RESEARCH        |n tech          |Spend "n" on research in technology "tech"
//|SCAN            |ship            |Have "ship" do a scan of its current location
//|SEND            |n sp            |Send "n" economic units to species "sp"
//|SHIPYARD        |                |Increase shipyard capacity by one.
//|START           |section         |Start processing "section" of the order form
//|SUMMARY         |                |Provide only a brief summary of combat results, instead of listing every single hit and miss
//|TARGET          |n               |Concentrate fire on target type "n" during combat
//|TEACH           |tech [n] sp     |Transfer knowledge of technology "tech" to species "sp" to maximum tech level "n"
//|TELESCOPE       |base            |Operate gravitic telescope on starbase "base"
//|TERRAFORM       |[n] pl          |Terraform planet "pl" using "n" TPs
//|TRANSFER        |n ab s,d        |Transfer "n" items of class "ab" from "s" to "d"
//|UNLOAD          |ship            |Transfer all CUs, IUs, and AUs from "ship" or
//|UNLOAD          |base            |starbase "base" to the planet it is at and install as many IUs and AUs as possible
//|UPGRADE         |ship            |Upgrade "ship" to age zero
//|UPGRADE         |base            |Upgrade starbase "base" to age zero
//|UPGRADE         |ship,n          |Upgrade "ship", spend "n"
//|UPGRADE         |base,n          |Upgrade starbase "base", spend "n"
//|VISITED         |x y z           |Mark a star system as having been visited, even if you have not actually been there.
//|WITHDRAW        |n1 n2 n3        |Set conditions for withdrawing from combat
//|WORMHOLE        |ship [,pl]      |Have "ship" jump to opposite end of wormhole and orbit planet "pl" on arrival
//|WORMHOLE        |base [,pl]      |Have starbase "base" jump to opposite end of wormhole and orbit planet "pl" on arrival
//|ZZZ             |                |Terminate a MESSAGE
//|===
//
//where:
//
//|===
//|ab|class abbreviation
//|base|name of a starbase, including "BAS" abbreviation
//|d|name of a ship, starbase, or planet, including class abbreviation
//|loc|jump destination, either "x y z" or "PL name"
//|n|a whole number, 0 or more
//|[n]|an optional whole number, 1 or more
//|name|name string, including any embedded spaces. May not start with a digit!
//|p|planet number
//|pl|planet name, including abbreviation "PL"
//|s|name of a ship, starbase, or planet, including class abbreviation
//|section|COMBAT, PRE-DEPARTURE, JUMPS, PRODUCTION or POST-ARRIVAL
//|ship|name of a ship, including class abbreviation
//|sp|species name, including "SP" abbreviation
//|tech|technology abbreviation: MI, MA, ML, GV, LS, or BI
//|x y z|galactic coordinates of a sector
//|===
