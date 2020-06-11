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

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.

import (
	"encoding/json"
	"fmt"
	"strconv"

	 "github.com/hyperledger/fabric-chaincode-go/shim"
        pb "github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type AdditionalInfo struct {
	Title string "json:\"title\""
	Text  string "json:\"text\""
	code  int64  "json:\"code\""
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Init")
	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Invoke")

	// GetArgs
	myargs := stub.GetArgs()
	fmt.Println("GetArgs()")
	for _, arg := range myargs {
		argStr := string(arg)
		fmt.Printf("%s", argStr)
	}
	fmt.Println()

	// GetStringArgs()
	stringArgs := stub.GetStringArgs()
	fmt.Println("GetStringArgs(): ", stringArgs)

	// GetArgsSlice()
	argsSlice, _ := stub.GetArgsSlice()
	fmt.Println("GetArgsSlice(): ", string(argsSlice))

	/*
		GetArgs()
		invokeab10
		GetStringArgs():  [invoke a b 10]
		GetArgsSlice():  invokeab10
	*/

	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Make payment of X units from A to B
		return t.invoke(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.query(stub, args)
	} else if function == "error" {
		msg := "{\"message\" : \"Error111\"}"
		return shim.Success([]byte(msg))
	} else if function == "more" {
		return t.more(stub, args)
	} else if function == "morequery" {
		return t.morequery(stub, args)
	} else {
		//return shi m.Error("can not find function")
		return pb.Response{Status: 404, Message: "Not Found"}
		/*
			ESCC invoke result: response:<status:404 message:"Not Found" >
			Error: endorsement failure during invoke. response: status:404 message:"Not Found"
		*/
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}
func (t *SimpleChaincode) more(stub shim.ChaincodeStubInterface, param []string) pb.Response {
	if len(param) != 3 {
		return shim.Error("# of parameter mismatching")
	}
	title, text, code := param[0], param[1], param[2]

	codeNum, err := strconv.ParseInt(code, 10, 32)
	if err != nil {
		return shim.Error("code error")
	}

	if len(title) == 0 || len(text) == 0 || len(code) == 0 {
		return shim.Error("value of paramter is not properly formatted")
	}

	// make my data of AdditionalInfo
	addInfo := &AdditionalInfo{Title: title, Text: text, code: codeNum}
	addInfoBytes, err := json.Marshal(addInfo)
	if err != nil {
		return shim.Error("failed to convert bytes " + err.Error())
	}

	err = stub.PutState(title, addInfoBytes) // key value에 공백 들어가면 조회시 값을 찾지 못함
	if err != nil {
		return shim.Error("PutState failure " + err.Error())
	}

	return shim.Success([]byte("invoke success"))
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) morequery(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	valbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}
	valueStr := &AdditionalInfo{}
	err = json.Unmarshal(valbytes, &valueStr)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to Unmarshal " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Title\":\"" + A + "\",\"Text\":\"" + string(valueStr.Text) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)

	// history
	iter, _ := stub.GetHistoryForKey(A)
	fmt.Println("Here is History for " + A)
	for iter.HasNext() {
		kv, _ := iter.Next()
		fmt.Println(string(kv.GetValue()) + " " + kv.GetTimestamp().String())
	}
	return shim.Success([]byte(string(valueStr.code)))
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
