package main

import (
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
* The Init method *
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method *
 called when an application requests to run the Smart Contract "tuna-chaincode"
 The app also specifies the specific smart contract function to call with args
*/
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger
	if function == "deliver" {
		return s.deliver(APIstub, args)
	} else if function == "changeDeliveryTime" {
		return s.changeDeliveryTime(APIstub, args)
	} else if function == "queryAsset" {
		return s.queryAsset(APIstub, args)
	} else if function == "transfer" {
		return s.transfer(APIstub, args)
	} else if function == "addAsset" {
		return s.addAsset(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

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
Client provides timestamps as strings of RFC3339 - args[2]
Client provides estimated time as string of Duration in go. e.g. "10h2m20s" - args[1]
*/
func (s *SmartContract) deliver(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	assetAsBytes, _ := APIstub.GetState(args[0])
	if assetAsBytes == nil {
		return shim.Error("Could not locate asset")
	}
	asset := Asset{}

	json.Unmarshal(assetAsBytes, &asset)
	if asset.State != "READY_FOR_DISTRIBUTION" {
		return shim.Error("delivering without asset state=READY_FOR_DISTRIBUTION ?")
	}
	asset.State = "ON_WAY"
	dur, err := time.ParseDuration(args[1])
	if err != nil {
		return shim.Error("Duration format is wrong!")
	}
	asset.Timestamp = args[2]
	currTime, err := time.Parse(time.RFC3339, asset.Timestamp)
	if err != nil {
		return shim.Error("Timestamp format is wrong.RFC3339 Required!")
	}
	asset.EstTime = currTime.Add(dur).String()

	assetAsBytes, _ = json.Marshal(asset)
	err = APIstub.PutState(args[0], assetAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to change asset holder: %s", args[0]))
	}

	return shim.Success(nil)
}

/*
Owner -> args[1]
Timestamp -> args[2]

*/
func (s *SmartContract) transfer(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	assetAsBytes, _ := APIstub.GetState(args[0])
	if assetAsBytes == nil {
		return shim.Error("Could not locate asset")
	}
	asset := Asset{}

	json.Unmarshal(assetAsBytes, &asset)
	// Normally check that the specified argument is a valid holder of asset
	// we are skipping this check for this example
	asset.Owner = args[1]
	asset.Timestamp = args[2]
	currTime, err := time.Parse(time.RFC3339, asset.Timestamp)
	if err != nil {
		return shim.Error("Timestamp format is wrong.RFC3339 Required!")
	}

	estTime, _ := time.Parse(time.RFC3339, asset.EstTime)
	dur := currTime.Sub(estTime)
	asset.Delay = strconv.FormatFloat(dur.Seconds(), 'f', 6, 64) //negative delay means that asset was delivered before estimated time.
	if asset.State != "ON_WAY" {
		return shim.Error("Trying to deliver an asset that isn't ON_WAY!")
	}
	asset.State = "DELIVERED"

	assetAsBytes, _ = json.Marshal(asset)
	err = APIstub.PutState(args[0], assetAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to change asset holder: %s", args[0]))
	}

	return shim.Success(nil)
}

/*
 * The changeTunaHolder method *
The data in the world state can be updated with who has possession.
This function takes in 2 arguments, tuna id and new holder name.
*/
func (s *SmartContract) changeDeliveryTime(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	assetAsBytes, _ := APIstub.GetState(args[0])
	if assetAsBytes == nil {
		return shim.Error("Could not locate asset")
	}
	asset := Asset{}

	json.Unmarshal(assetAsBytes, &asset)
	dur, err := time.ParseDuration(args[1])
	if err != nil {
		return shim.Error("Duration format is wrong!")
	}
	if asset.State != "ON_WAY" {
		return shim.Error("changeDeliveryTime on asset that isn't ON_WAY")
	}

	estTime, _ := time.Parse(time.RFC3339, asset.EstTime)
	asset.EstTime = estTime.Add(dur).String()
	asset.Delay += strconv.FormatFloat(dur.Seconds(), 'f', 6, 64)

	assetAsBytes, _ = json.Marshal(asset)
	err = APIstub.PutState(args[0], assetAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to change asset holder: %s", args[0]))
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
