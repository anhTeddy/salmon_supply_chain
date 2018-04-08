/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Agreement struct {
	Price float64 `json:"price"`
}

type AgreementChaincode struct {
}

func (t *AgreementChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t *AgreementChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	function, args := stub.GetFunctionAndParameters()

	var result []byte
	var err error
	switch function {
	case "recordAgreement":
		result, err = recordAgreement(stub, args)
	case "queryAgreement":
		result, err = queryAgreement(stub, args)
	default:
		return fmt.Errorf("No function detected: %s", function)
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(result)
}

func recordAgreement(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 2 {
		return nil, fmt.Error("Incorrect number of arguments. Expecting 2")
	}

	id := args[0]
	price, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse price: %s", err)
	}

	agreement := Agreement{price}
	data, err := json.Marshal(agreement)
	if err != nil {
		return nil, fmt.Errorf("Marshal failed: %s", err)
	}

	err = stub.PutState(id, data)
	if err != nil {
		return nil, fmt.Errorf("Failed to put state: %s", args[0])
	}

	return nil, nil
}

func queryAgreement(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 1")
	}
	id := args[0]

	data, err := stub.GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Get agreement failed: %s", err)
	}

	return data, nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	// Start the chaincode and make it ready for futures requests
	err := shim.Start(new(AgreementChaincode))
	if err != nil {
		fmt.Printf("Error starting AgreementChaincode: %s", err)
	}
}
