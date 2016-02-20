/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"github.com/openblockchain/obc-peer/openchain/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// Run callback representing the invocation of a chaincode
// This chaincode will manage two accounts A and B and will transfer X units from A to B upon invoke
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	switch function {
	case "createAccount":
		if len(args) <= 0 {
			return nil, errors.New("createAccount operation must include an accound ID")
		}
		accountID := args[0]

		var columns []*shim.Column

		accountIDCol := shim.Column{Value: &shim.Column_String_{String_: accountID}}
		balanceCol := shim.Column{Value: &shim.Column_Int32{Int32: 0}}

		columns = append(columns, &accountIDCol)
		columns = append(columns, &balanceCol)

		row := shim.Row{columns}
		ok, err := stub.InsertRow("accounts", row)
		if err != nil {
			return nil, errors.New("createAccount operation failed while accessing state")
		}
		if !ok {
			return nil, errors.New("createAccount operation failed. Account already exists")
		}

		return nil, nil

	case "deposit":
		if len(args) < 2 {
			return nil, errors.New("deposit operation must include an accound ID and amount")
		}
		accountID := args[0]

		var key []shim.Column
		accountIDCol := shim.Column{Value: &shim.Column_String_{String_: accountID}}
		key = append(key, accountIDCol)
		row, err := stub.GetRow("accounts", key)
		if err != nil {
			return nil, errors.New("deposit operation fail. Error fetching account ID")
		}
		if &row == nil {
			return nil, errors.New("deposit operation fail. Account ID does not exist")
		}

		balanceCol := row.Columns[1]
		currentBal := balanceCol.GetInt32()

		depositBal, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			return nil, errors.New("deposit operation fail. Deposit amount is invalid")
		}
		newBal := currentBal + int32(depositBal)

		balanceCol.Value = &shim.Column_Int32{Int32: newBal}
		row.Columns[1] = balanceCol

		ok, err := stub.ReplaceRow("accounts", row)
		if err != nil {
			return nil, errors.New("deposit operation fail. Error updating balance")
		}
		if !ok {
			return nil, errors.New("deposit operation fail. Account not found")
		}

		return nil, nil

	case "init":

		// Create the accounts table
		var columnDefs []*shim.ColumnDefinition

		accountIDColumnDef := shim.ColumnDefinition{Name: "accountID",
			Type: shim.ColumnDefinition_STRING, Key: true}
		balanceColumnDef := shim.ColumnDefinition{Name: "balance",
			Type: shim.ColumnDefinition_INT32, Key: false}

		columnDefs = append(columnDefs, &accountIDColumnDef)
		columnDefs = append(columnDefs, &balanceColumnDef)

		stub.CreateTable("accounts", columnDefs)

	default:
		return nil, errors.New("Unsupported operation")
	}
	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	switch function {
	case "getBalance":
		if len(args) <= 0 {
			return nil, errors.New("getBalance operation must include an accound ID")
		}
		accountID := args[0]

		var key []shim.Column
		accountIDCol := shim.Column{Value: &shim.Column_String_{String_: accountID}}
		key = append(key, accountIDCol)
		row, err := stub.GetRow("accounts", key)
		if err != nil {
			return nil, errors.New("getBalance operation fail. Error fetching account ID")
		}
		if &row == nil {
			return nil, errors.New("getBalance operation fail. Account ID does not exist")
		}

		balanceCol := row.Columns[1]
		balance := balanceCol.GetInt32()

		bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytes, uint32(balance))
		return bytes, nil

	default:
		return nil, errors.New("Unsupported operation")
	}
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
