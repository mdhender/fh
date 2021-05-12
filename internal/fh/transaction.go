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

package fh

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
)

/* Interspecies transactions. */

const MAX_TRANSACTIONS = 1000

type TransType int

const (
	EU_TRANSFER               TransType = 1
	MESSAGE_TO_SPECIES                  = 2
	BESIEGE_PLANET                      = 3
	SIEGE_EU_TRANSFER                   = 4
	TECH_TRANSFER                       = 5
	DETECTION_DURING_SIEGE              = 6
	SHIP_MISHAP                         = 7
	ASSIMILATION                        = 8
	INTERSPECIES_CONSTRUCTION           = 9
	TELESCOPE_DETECTION                 = 10
	ALIEN_JUMP_PORTAL_USAGE             = 11
	KNOWLEDGE_TRANSFER                  = 12
	LANDING_REQUEST                     = 13
	LOOTING_EU_TRANSFER                 = 14
	ALLIES_ORDER                        = 15
)

type TransactionData struct {
	Type      int    `json:"type"` /* Transaction type. */
	Donor     int    `json:"donor"`
	Recipient int    `json:"recipient"`
	Value     int    `json:"value"` /* Value of transaction. */
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Z         int    `json:"z"`
	PN        int    `json:"pn"`       /* Location associated with transaction. */
	Number1   int    `json:"number_1"` /* Other items associated with transaction.*/
	Name1     string `json:"name_1"`
	Number2   int    `json:"number_2"`
	Name2     string `json:"name_2"`
	Number3   int    `json:"number_3"`
	Name3     string `json:"name_3"`
}

func GetTransactionData(galaxyPath string) ([]*TransactionData, error) {
	// read transactions from file
	data, err := ioutil.ReadFile(path.Join(galaxyPath, "interspecies.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var td []*TransactionData
	err = json.Unmarshal(data, &td)
	if err != nil {
		return nil, err
	}
	return td, nil
}
