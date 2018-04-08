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
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Pallinder/go-randomdata"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Salmon struct {
	Vessel   string `json:"vessel"`
	Datetime string `json:"datetime"`
	Location string `json:"location"`
	Holder   string `json:"holder"`
}

type SalmonChaincode struct {
}

func (t *SalmonChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {

	if len(args) != 1 && len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting 1 or 0")
	}

	salmonCount := 20

	if len(args) == 1 {
		var err error
		salmonCount, err = strconv.Atoi(args[0])
		if err != nil {
			return nil, fmt.Errorf("Parse salmon count err: %s", err)
		}
	}

	for i := 1; i <= spawnCount; i++ {
		_, err := recordSalmon(stub, []string{
			strconv.Itoa(i),
			randomdata.SillyName(),
			randomdata.FullDateInRange("2018-01-01", "2018-04-30"),
			randomdata.City(),
			"fredrick",
		})
		if err != nil {
			return nil, shim.Errorf(err)
		}
	}

	return shim.Success(nil)
}

func (t *SalmonChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	function, args := stub.GetFunctionAndParameters()

	var result []byte
	var err error
	switch function {
	case "initLedger":
		result, err = initLedger(stub, args)
	case "recordSalmon":
		result, err = recordSalmon(stub, args)
	case "changeSalmonHolder":
		result, err = changeSalmonHolder(stub, args)
	case "querySalmon":
		result, err = querySalmon(stub, args)
	case "queryAllSalmon":
		result, err = queryAllSalmon(stub, args)
	default:
		return fmt.Errorf("No function detected: %s", function)
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success(result)
}

func recordSalmon(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 5 {
		return nil, shim.Errorf("Incorrect number of arguments. Expecting 5")
	}

	id := args[0]
	vessel := args[1]
	datetime := args[2]
	location := args[3]
	holder := args[4]

	salmon := Salmon{Vessel: vessel, Datetime: datetime, Location: location, Holder: holder}
	data, err := json.Marshal(salmon)
	if err != nil {
		return nil, fmt.Errorf("Marshall fail: %s", err)
	}

	err = stub.PutState(id, data)
	if err != nil {
		return nil, fmt.Errorf("Failed to set asset: %s", args[0])
	}

	return nil, nil
}

func changeSalmonHolder(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 2 {
		return nil, shim.Errorf("Incorrect number of arguments. Expecting 2")
	}

	id := args[0]
	holder := args[1]

	data, err := stub.GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Get salmon failed: %s", err)
	}

	var salmon Salmon
	err = json.Unmarshal(data, &salmon)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal failed: %s", err)
	}

	salmon.Holder = holder

	data, err = json.Marshal(salmon)
	if err != nil {
		return nil, fmt.Errorf("Marshal failed: %s", err)
	}

	err = stub.PutState(id, data)
	if err != nil {
		return nil, fmt.Errorf("Failed to set asset: %s", args[0])
	}

	return nil, nil
}

func querySalmon(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 1")
	}

	id := args[0]
	data, err := stub.GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Get salmon failed: %s", err)
	}

	return marshalSalmonData(id, data), nil
}

func marshalSalmonData(id string, data []byte) []byte {
	var salmon Salmon

	err = json.Unmarshal(data, &salmon)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal failed: %s", err)
	}

	document := struct {
		ID     string `json:"id"`
		Salmon `json:",inline"`
	}{id, salmon}

	documentData, err := json.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("Marshal failed: %s", err)
	}

	return documentData
}

func queryAllSalmon(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) > 2 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 0 to 2")
	}

	var startId, endId string
	if len(args) > 0 {
		startId = args[0]
	}
	if len(args) > 1 {
		endId = args[1]
	}

	iter, err := stub.GetStateByRange(startId, endId)
	if err != nil {
		return nil, fmt.Errorf("Get salmons failed: %s", err)
	}
	defer iter.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for iter.HasNext() {
		queryResponse, err := iter.Next()
		if err != nil {
			return nil, fmt.Errorf("retrive next salmon fail: %s", err)
		}

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(marshalSalmonData(queryResponse.Key, queryResponse.Value)))
		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]")

	fmt.Printf("- queryAllSalmons:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// main function starts up the chaincode in the container during instantiate
func main() {
	// Start the chaincode and make it ready for futures requests
	err := shim.Start(new(SalmonChaincode))
	if err != nil {
		fmt.Printf("Error starting SalmonChaincode: %s", err)
	}
}
