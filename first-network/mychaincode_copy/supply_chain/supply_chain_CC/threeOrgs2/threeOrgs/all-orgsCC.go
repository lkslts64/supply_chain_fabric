//convert all times in time.Parse(rfc3999,arg_given)
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"strings"
	"time"
)

/*
API:

addCrude
refine
addFuelForDelivery - coupled with a retailer.
deliverFuel - make a plan for distributing to different retailers. accumulate addFuelDelivery tx's.


*/
// Define the Smart Contract structure
type SmartContract struct {
}

/*
Crude Oil ID should be like this: CrudeXXXX where XXXX is an ever increasing number.

*/
type Vehicle struct {
	Type string
	ID   string
}
type DeliveryDetails struct {
	//State            string
	EstTime          time.Time
	Delay            float64
	StartingLocation string
	Destination      string
}
type TxProof struct {
	URL  string
	Hash string
}
type AssetDetails struct {
	Value    float64
	Quantity int
	Owner    string
	State    string
}

/*
Put in db wih key in form 'CrudeXXX'
*/
type Crude struct {
	AD        AssetDetails
	DD        DeliveryDetails
	Proof     TxProof
	Veh       Vehicle
	Timestamp time.Time
}

/*
Put in db wih key in form 'FuelXXX'
*/
type Fuel struct {
	AD        AssetDetails
	Density   float64 //quality
	Type      string
	CrudeID   string //like parent ID
	Timestamp time.Time
}

/*
Put in db wih key in form 'FuelOrderXXX'
*/
type FuelOrder struct {
	AD        AssetDetails
	Dest      string
	Proof     TxProof
	FuelID    string //like parent ID
	Timestamp time.Time
}

/*type FuelDelivery struct {
	DD    DeliveryDetails
	Proof TxProof
}
*/
type FuelOrderID = string

/*
A delivery plan from refinary towards the gas stations.
Contains the vehicle that will deliver the fuels at many fueling stations
A map for easy access to delivery details with key the orders that org2 has added.
*/
type FuelDeliveryPlan struct {
	Veh  Vehicle
	Plan map[FuelOrderID]DeliveryDetails
}

/*
CrudeOil becomes Fuel at refinary. A tx should be submited in order to make this transformation.
Produced by org2 - at refinary and distributed by org3
*/

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
	if function == "deliverCrude" {
		return s.deliverCrude(APIstub, args)
	} else if function == "refine" {
		return s.refine(APIstub, args)
	} else if function == "addFuelOrder" {
		return s.addFuelOrder(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}
func HasPrefixOrg(s string) bool {
	return strings.HasPrefix(s, "org")
}

func NewAssetDetails(val, quant, own, st string) (AssetDetails, error) {
	//value can be zero if shipper doesn't want to make it public.
	value, err := strconv.ParseFloat(val, 64)
	if err != nil || value < 0 {
		return AssetDetails{}, errors.New("Value is not a float number")
	}
	quantity, err := strconv.ParseInt(quant, 10, 64)
	if err != nil || quantity < 0 {
		return AssetDetails{}, errors.New("Quantity is not an int number")
	}
	if HasPrefixOrg(own) == false {
		return AssetDetails{}, errors.New("Owner value is not prefixed with string 'org'")
	}
	return AssetDetails{value, int(quantity), own, st}, nil
}
func NewDeliveryDetails(est, sloc, dest string) (DeliveryDetails, error) {

	estTime, err := time.Parse(time.RFC3339, est)
	if err != nil {
		return DeliveryDetails{}, errors.New("Time is not in RFC3339 format")
	}
	if HasPrefixOrg(sloc) == false {
		return DeliveryDetails{}, errors.New("Starting Location value is not prefixed with 'org'")
	}
	if HasPrefixOrg(dest) == false {
		return DeliveryDetails{}, errors.New("Destination value is not prefixed with 'org'")
	}
	return DeliveryDetails{estTime, 0, sloc, dest}, nil
}

/*
A dummy proof constructor.
Hash is the SHA256("ait")
*/
func NewProof() TxProof {
	return TxProof{"www.ait.gr", "7cb0d761a60f4968299cda86c333dafe318fbf87b0979f60befd0499e39e21d6"}
}
func NewVehicle(typ, id string) Vehicle {
	return Vehicle{typ, id}
}
func RFCtoTime(rfc string) (time.Time, error) {
	currtime, err := time.Parse(time.RFC3339, rfc)
	if err != nil {
		return time.Time{}, errors.New("Time not provided in RFC3339 format.")
	}
	return currtime, nil
}

/*
args[0] = crudeID like 'CrudeXXXX'
arg1 = value,arg2 = quantity, arg3 = owner
arg4 = estTime, arg5 = startLoc, arg6 = dest
*/
func (s *SmartContract) deliverCrude(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	//check if creator is org1-shipper??
	if len(args) != 9 {
		return shim.Error("Incorrect number of arguments. Expecting 9")
	}
	AD, err := NewAssetDetails(args[1], args[2], args[3], "ON_WAY")
	if err != nil {
		return shim.Error(err.Error())
	}
	DD, err := NewDeliveryDetails(args[4], args[5], args[6])
	if err != nil {
		return shim.Error(err.Error())
	}
	Proof := NewProof()
	//hardcoded vehID.TODO: construct base on the Hash(args[1]+args[2]...+)
	Veh := NewVehicle("Vessel", args[7])
	Timestamp, err := RFCtoTime(args[8])
	if err != nil {
		return shim.Error(err.Error())
	}
	crude := Crude{AD, DD, Proof, Veh, Timestamp}
	crudeAsBytes, _ := json.Marshal(crude)
	//TODO: check if crude already exists.
	err = stub.PutState(args[0], crudeAsBytes) //if an crude with the same ID already exists, this will override it.
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to add crude: %s", args[0]))
	}

	return shim.Success(nil)
}

/*
args[0] = fuelID like 'FuelXXXX'
arg1 = value,arg2 = quantity, arg3 = owner
arg4 = density,arg5 = type_of_fuel, arg6 = CrudeID (ancestor ID)
arg7 = timestamp.
*/
func (s *SmartContract) refine(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}
	AD, err := NewAssetDetails(args[1], args[2], args[3], "REFINED")
	if err != nil {
		return shim.Error(err.Error())
	}
	Density, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return shim.Error("Density should be a float number!")
	}
	Timestamp, err := RFCtoTime(args[7])
	if err != nil {
		return shim.Error(err.Error())
	}
	//ensure crudeID exists in db.
	crudebytes, _ := stub.GetState(args[6])
	if crudebytes == nil {
		return shim.Error("ID of crude doesn't exist!")
	}
	if fuelbytes, _ := stub.GetState(args[0]); fuelbytes != nil {
		return shim.Error("ID of fuel already exists.")
	}
	fuel := Fuel{AD, Density, args[5], args[6], Timestamp}
	fuelAsBytes, _ := json.Marshal(fuel)
	err = stub.PutState(args[0], fuelAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to add fuel: %s", args[0]))
	}
	return shim.Success(nil)
}

/*
arg1-3 = asset_details
arg4 = dest, arg5 = fuelID
arg6 = timestamp
*/
func (s *SmartContract) addFuelOrder(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}
	AD, err := NewAssetDetails(args[1], args[2], args[3], "READY_FOR_DISTRIBUTION")
	if err != nil {
		return shim.Error(err.Error())
	}
	if HasPrefixOrg(args[4]) == false {
		return shim.Error("Destination doesn't start with org!")
	}
	Proof := NewProof()
	//check that fuelID exists
	fuelbytes, _ := stub.GetState(args[5])
	if fuelbytes == nil {
		return shim.Error("FuelID doens't exist!")
	}
	Timestamp, err := RFCtoTime(args[6])
	if err != nil {
		return shim.Error(err.Error())
	}
	//check that fuelOrderID doens't exist
	if fuelOrderbytes, _ := stub.GetState(args[5]); fuelOrderbytes != nil {
		return shim.Error("FuelOrderID already exists")
	}

	fuelOrder := FuelOrder{AD, args[4], Proof, args[5], Timestamp}
	fuelAsBytes, _ := json.Marshal(fuelOrder)
	err = stub.PutState(args[0], fuelAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to add fuelOrder: %s", args[0]))
	}
	return shim.Success(nil)

}

/*
args of this invokation:
	PlanID
	TruckID
	{FuelOrderID,EstTime,Sloc,Dest}
	{FuelOrderID,EstTime,Sloc,Dest}
	.
	.
	.
	{FuelOrderID,EstTime,Sloc,Dest}
*/
func (s *SmartContract) deliverFuel(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) < 2 {
		return shim.Error("Expecting more args")
	}
	Veh := NewVehicle("Truck", args[1])
	orders := []FuelOrderID(args[2:])
	//check that client supplied properly the # of args
	if len(orders) == 0 {
		return shim.Error("At least one delivery should be specified")
	} else if len(orders)%4 != 0 {
		return shim.Error("Arguments dont match!Pattern should be {FuelOrderID,EstTime,Sloc,Dest}... ")
	}
	Plan := make(map[FuelOrderID]DeliveryDetails)
	//orders[i] = orderID , orders[i+1] = estTime , i+2 = sloc , i+3 = dest
	for i := 0; i < len(orders); i += 4 {
		fuelOrderbytes, _ := stub.GetState(orders[i])
		if fuelOrderbytes == nil {
			return shim.Error(fmt.Sprintf("FuelOrderID %s does not exist", orders[i]))
		} else {
			fuelOrder := FuelOrder{}
			json.Unmarshal(fuelOrderbytes, &fuelOrder)
			fuelOrder.AD.State = "ON_WAY"
			newFuelOrderbytes, _ := json.Marshal(fuelOrder)
			err := stub.PutState(orders[i], newFuelOrderbytes)
			if err != nil {
				return shim.Error(fmt.Sprint("Failed to add %s with different state", orders[i]))
			}

		}
		DD, err := NewDeliveryDetails(orders[i+1], orders[i+2], orders[i+3])
		if err != nil {
			return shim.Error(err.Error())
		}
		Plan[orders[i]] = DD
	}

	fuelDeliveryPlan := FuelDeliveryPlan{Veh, Plan}
	fuelDeliveryPlanAsBytes, _ := json.Marshal(fuelDeliveryPlan)
	err := stub.PutState(args[0], fuelDeliveryPlanAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprint("Failed to add Plan %s in db", args[0]))

	}

	return shim.Success(nil)

}

/*
 * The initLedger method *
func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {

	tim1, _ := time.Parse(time.RFC3339, "2019-03-27T14:12:47.921Z")
	tim2, _ := time.Parse(time.RFC3339, "2019-03-27T14:12:47.921Z")
	tim3, _ := time.Parse(time.RFC3339, "2019-03-27T14:12:47.921Z")
	asset := []Asset{
		Asset{Value: "92", Quantity: "200053", Owner: "1504054225", State: "READY_FOR_DISTRIBUTION", Type: "Car", Timestamp: tim1, EstTime: time.Time{}, Delay: 0, StartingLocation: "org2", Destination: "org1", ProofURL: "www.google.com", ProofHash: nil},
		Asset{Value: "34", Quantity: "204", Owner: "1504054225", State: "READY_FOR_DISTRIBUTION", Type: "Car", Timestamp: tim2, EstTime: time.Time{}, Delay: 0, StartingLocation: "org1", Destination: "org2", ProofURL: "www.google.com", ProofHash: nil},
		Asset{Value: "57", Quantity: "5", Owner: "1504054225", State: "READY_FOR_DISTRIBUTION", Type: "Car", Timestamp: tim3, EstTime: time.Time{}, Delay: 0, StartingLocation: "org1", Destination: "org2", ProofURL: "www.google.com", ProofHash: nil},
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
	currTime, err := time.Parse(time.RFC3339, args[2])
	if err != nil {
		return shim.Error("Timestamp format is wrong.RFC3339 Required!")
	}
	asset.Timestamp = currTime
	asset.EstTime = currTime.Add(dur)

	assetAsBytes, _ = json.Marshal(asset)
	err = APIstub.PutState(args[0], assetAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to change asset holder: %s", args[0]))
	}

	return shim.Success(nil)
}

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
	if asset.State != "ON_WAY" {
		return shim.Error("Trying to deliver an asset that isn't ON_WAY!")
	}
	asset.State = "DELIVERED"
	//asset.Timestamp = args[2]
	currTime, err := time.Parse(time.RFC3339, args[2])
	if err != nil {
		return shim.Error("Timestamp format is wrong.RFC3339 Required!")
	}
	asset.Timestamp = currTime
	//estTime, _ := time.Parse(time.RFC3339, asset.EstTime)
	dur := currTime.Sub(asset.EstTime)
	asset.Delay = dur.Seconds() //strconv.FormatFloat(dur.Seconds(), 'f', 6, 64) //negative delay means that asset was delivered before estimated time.

	assetAsBytes, _ = json.Marshal(asset)
	err = APIstub.PutState(args[0], assetAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to change asset holder: %s", args[0]))
	}

	return shim.Success(nil)
}

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

	//estTime, _ := time.Parse(time.RFC3339, asset.EstTime)
	asset.EstTime = asset.EstTime.Add(dur)
	//prev_delay, _ := strconv.ParseFloat(asset.Delay, 64)
	//newdelay := prev_delay + dur.Seconds()
	asset.Delay += dur.Seconds()

	assetAsBytes, _ = json.Marshal(asset)
	err = APIstub.PutState(args[0], assetAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to change asset holder: %s", args[0]))
	}

	return shim.Success(nil)
}
*/
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
