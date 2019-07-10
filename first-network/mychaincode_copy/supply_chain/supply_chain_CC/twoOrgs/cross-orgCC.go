/*
Notes: I can use GetTxTimestamp to manage the currTime.

*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"time"
)

// Define the Smart Contract structure
type SmartContract struct {
}

type Asset struct {
	//ID               string
	Value            string `private field - value agreed with retailer`
	Quantity         string `private field - value agreed with retailer`
	Owner            string `who currently owns the asset - retailer or factory`
	State            string `READY_FOR_DISTRIBUTION,ON_WAY,DELIVERED,DISFUNCTIONAL`
	Type             string `Car,Machinery,...`
	Timestamp        string
	EstTime          string `estimated arrival time of the asset when it is ON_WAY`
	Delay            string `how much is delayed`
	StartingLocation string `where the asset is located before transport`
	Destination      string
	ProofURL         string `URL that cointains proof that this transaction is actualy real`
	ProofHash        string `Hash of the above docuemnt that proves the existence of the transaction`
}

/*
* The Init method *
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method *
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger
	if function == "addAsset" {
		return s.addAsset(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "queryAllAssets" {
		return s.queryAllAssets(APIstub)
	} else if function == "queryAsset" {
		return s.queryAsset(APIstub, args)
	} else if function == "makeDisFunctional" {
		return s.makeDisFunctional(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

/*
 */
func (s *SmartContract) queryAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	assetAsBytes, _ := APIstub.GetState(args[0])
	if assetAsBytes == nil {
		return shim.Error("Could not locate asset")
	}
	return shim.Success(assetAsBytes)
}

/*
 * The initLedger method *
 */
func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	asset := []Asset{
		Asset{Value: "92", Quantity: "200053", Owner: "1504054225", State: "READY_FOR_DISTRIBUTION", Type: "Car", Timestamp: "2019-03-27T14:12:47.921Z", EstTime: "", Delay: "", StartingLocation: "org2", Destination: "org1", ProofURL: "www.google.com", ProofHash: "dummy_hash"},
		Asset{Value: "34", Quantity: "204", Owner: "1504054225", State: "READY_FOR_DISTRIBUTION", Type: "Car", Timestamp: "2019-03-27T13:12:47.921Z", EstTime: "", Delay: "", StartingLocation: "org1", Destination: "org2", ProofURL: "www.google.com", ProofHash: "dummy_hash"},
		Asset{Value: "57", Quantity: "5", Owner: "1504054225", State: "READY_FOR_DISTRIBUTION", Type: "Car", Timestamp: "2019-03-27T15:12:47.921Z", EstTime: "", Delay: "", StartingLocation: "org1", Destination: "org2", ProofURL: "www.google.com", ProofHash: "dummy_hash"},
	}

	i := 0
	for i < len(asset) {
		fmt.Println("i is ", i)
		assetAsBytes, _ := json.Marshal(asset[i])
		APIstub.PutState(strconv.Itoa(i+1), assetAsBytes)
		fmt.Println("Added", asset[i])
		i = i + 1
	}

	return shim.Success(nil)
}

/*
Time stamp is in RFC3339 format passed as a string
*/
func (s *SmartContract) addAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	//check if creator is factory??
	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. Expecting 13")
	}
	_, err := time.Parse(time.RFC3339, args[5])
	if err != nil {
		return shim.Error("Reuqired RFC3339 in addAsset")
	}
	var asset = Asset{Value: args[1], Quantity: args[2], Owner: args[3], State: "READY_FOR_DISTRIBUTION", Type: args[4], Timestamp: args[5], EstTime: "", Delay: "", StartingLocation: args[6], Destination: args[7], ProofURL: args[8], ProofHash: args[9]}
	assetAsBytes, _ := json.Marshal(asset)
	err = APIstub.PutState(args[0], assetAsBytes) //if an asset with the same ID already exists, this will override it.
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to add asset: %s", args[0]))
	}

	return shim.Success(nil)

}

/*
 */
func (s *SmartContract) queryAllAssets(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "0"
	endKey := "999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add comma before array members,suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllTuna:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

/*
arg[0]=ID and arg[1]=Timestamp
*/
func (s *SmartContract) makeDisFunctional(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	//check if creator belongs to retailer??
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	assetAsBytes, _ := APIstub.GetState(args[0])
	if assetAsBytes == nil {
		return shim.Error("Could not locate asset")
	}
	asset := Asset{}

	json.Unmarshal(assetAsBytes, &asset)
	if asset.State != "DELIVERED" {
		return shim.Error(fmt.Sprintf("Attempted to make an asset DISFUNCTIONAL without first being DELIVRED! Asset was: %s", args[0]))
	}

	asset.State = "DISFUNCTIONAL"
	//set value zero because asset is garbage now.
	asset.Value = "0"
	asset.Timestamp = args[1]
	_, err := time.Parse(time.RFC3339, asset.Timestamp)
	if err != nil {
		return shim.Error("RFC3339 required in makeDisFunctional")
	}
	assetAsBytes, _ = json.Marshal(asset)
	err = APIstub.PutState(args[0], assetAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to change asset state to DISFUNCTIONAL: %s", args[0]))
	}

	return shim.Success(nil)
}

/*
 * main function *
calls the Start function
The main function starts the chaincode in the container during instantiation.
*/
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
