package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// SmartContract is an empty struct to hold smart contract logic.
type SmartContract struct {
}

// Vehicle is the struct for vehicle registration and activation information.
type Vehicle struct {
	Serial          string `json:serial`
	Owner           string `json:owner`
	VehicleType     string `json:vehicletype`
	CurrentLocation string `json:currentlocation`
	InService       bool   `json:inService`
}

// Trip is a struct containing the information for any trips.
type Trip struct {
	Serial    string  `json:serial`
	StartTime int64   `json:startTime`
	StartLat  float64 `json:startLat`
	StartLong float64 `json:startLong`
	EndTime   int64   `json:endTime`
	EndLat    float64 `json:endLat`
	EndLong   float64 `json:endLong`
}

// main is the entry point for the chaincode.
func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error starting transaction chaincode: %s", err)
	}
}

// Init is the function called for an initial invoke.
func (t *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke is where we route to the specific method user is attempting to call.
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "registerVehicle" {
		return s.registerVehicle(APIstub, args)
	} else if function == "seeAllVehicles" {
		return s.seeAllVehicles(APIstub, args)
	} else if function == "getHistoryForVehicle" {
		return s.getHistoryForVehicle(APIstub, args)
	} else if function == "takeOutOfService" {
		return s.takeOutOfService(APIstub, args)
	} else if function == "putInService" {
		return s.putInService(APIstub, args)
	} else if function == "updateLocation" {
		return s.updateLocation(APIstub, args)
	} else if function == "recordTrip" {
		return s.recordTrip(APIstub, args)
	} else if function == "getTripsForVehicle" {
		return s.getTripsForVehicle(APIstub, args)
	} else if function == "queryLastTrip" {
		return s.queryLastTrip(APIstub, args)
	} else if function == "getAllTripsAllVehicles" {
		return s.getAllTripsAllVehicles(APIstub, args)
	}
	// If none of the handler functions were invoked, send back an error.
	return shim.Error("Invalid Smart Contract function name.")
}

// seeAllVehicles brings up all vehicle records. Confirm this is fixed.
func (s *SmartContract) seeAllVehicles(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	startKey := "V"
	endKey := "Vzzzzzzzzzzz"
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
		// Add a comma before array members, suppress it for the first array member
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

	return shim.Success(buffer.Bytes())
}

// registerVehicle is registers a vehicle record to the blockchain.
func (s *SmartContract) registerVehicle(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Returning an error if there aren't the right number of arguments.
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// We get the serial number for the vehicle, the company that owns the vehicle, and the type of vehicle. We mark its current location as "oos" (for out of service) and mark InService as false.
	var k = "V" + args[0]
	var identity = Vehicle{Serial: args[0], Owner: args[1], VehicleType: args[2], CurrentLocation: "oos", InService: false}

	// json.Marshal is returning a JSON encoding of the identity struct.
	identityAsBytes, _ := json.Marshal(identity)

	// Create a key that identifies this record as a vehicle.
	var k = "V" + args[0]

	// Put the record on the chain, and catch any error returned.
	err := APIstub.PutState(k, identityAsBytes)

	// Got some logging happening here...
	var logger = shim.NewLogger("vehicleReg")
	logger.Info("Lance Debug - This is the identityAsBytes...")
	logger.Info(identityAsBytes)

	// If there is any error from creating a record for this on the chain, exit on that error.
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record the vehicle: %s", args[0]))
	}

	///////////// TRYING OUT SETTING AN EVENT /////////////
	err = APIstub.SetEvent("DeviceRegistered", identityAsBytes)
	///=================================================

	// If all went well, return success.
	return shim.Success(nil)
}

// getHistoryForVehicle return history for vehicle record.
func (t *SmartContract) getHistoryForVehicle(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	// Make sure there is at least one argument, the vehicle serial number.
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Provide vehicle serial number.")
	}

	// Taking the serial number from the arguments and making it the recordKey.
	recordKey := "V" + args[0]

	// Getting the results for this key.
	resultsIterator, err := stub.GetHistoryForKey(recordKey)

	// If there is an error getting that record, return the error.
	if err != nil {
		return shim.Error(err.Error())
	}

	// Deferring a close on this iterator until the function returns.
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the key/value pair
	var buffer bytes.Buffer
	buffer.WriteString("[")

	// This boolean is used to let us know if we're on the first item for vehicle.
	bArrayMemberAlreadyWritten := false

	// Now we loop through the history results...
	for resultsIterator.HasNext() {
		// The next item becomes the response, and we check for any error.
		response, err := resultsIterator.Next()

		// If there's an error, return that error.
		if err != nil {
			return shim.Error(err.Error())
		}

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		// Write the Transaction ID for this record.
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		// Write the value for the transaction.
		buffer.WriteString(", \"Value\":")

		// If it was a delete operation on given transaction, we should set the
		// corresponding value to NULL. Otherwise, we will write the response.Value
		// as-is (with the value itself as JSON)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		// Write the timestamp of the transaction.
		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")
		buffer.WriteString("}")
		// Mark our first-item-used check boolean to true.
		bArrayMemberAlreadyWritten = true
	}
	// Close out the array JSON payload.
	buffer.WriteString("]")
	// Return the history buffer.
	return shim.Success(buffer.Bytes())
}

// takeOutOfService marks a registered vehicle as taken out of service.
func (s *SmartContract) takeOutOfService(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Returning an error if there aren't the right number of arguments.
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1, the vehicle serial number.")
	}

	// Turn the serial number entered into the vehicle record key.
	recordKey := "V" + args[0]

	// Get the current state for the vehicle.
	vehicleAsBytes, _ := APIstub.GetState(recordKey)

	// Make sure we got a result. If we didn't, throw an error.
	if vehicleAsBytes == nil {
		return shim.Error("Could not locate vehicle with that serial number")
	}

	// Setting up a new empty Vehicle struct.
	vehicle := Vehicle{}

	// Unmarshalling/parsing the vehicle state into the struct.
	json.Unmarshal(vehicleAsBytes, &vehicle)

	// Adding in the current location and marking the InService boolean to true
	vehicle.CurrentLocation = "oos"
	vehicle.InService = false

	// Putting this back into JSON.
	vehicleAsBytes, _ = json.Marshal(vehicle)

	// Putting that JSON for the updated vehicle info on the chain, capturing any error.
	err := APIstub.PutState(args[0], vehicleAsBytes)

	// Got some logging happening here...
	var logger = shim.NewLogger("vehicleReg")
	logger.Info("Lance Debug - This is the vehicleAsBytes...")
	logger.Info(vehicleAsBytes)

	// If there is any error from creating a record for this on the chain, exit on that error.
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record the vehicle: %s", args[0]))
	}

	// If all went well, return success.
	return shim.Success(nil)
}

// putInService to get a registered vehicle put into service.
func (s *SmartContract) putInService(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Returning an error if there aren't the right number of arguments, the serial number (args[0]), and current location(args[1])
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. The vehicle serial number and its current location.")
	}

	// Turn the serial number entered into the vehicle record key.
	recordKey := "V" + args[0]

	// Get the current state for the vehicle.
	vehicleAsBytes, _ := APIstub.GetState(recordKey)

	// Make sure we got a result. If we didn't, throw an error.
	if vehicleAsBytes == nil {
		return shim.Error("Could not locate Vehicle with that serial number")
	}

	// Setting up a new empty Vehicle struct.
	vehicle := Vehicle{}

	// Unmarshalling/parsing the vehicle state into the struct.
	json.Unmarshal(vehicleAsBytes, &vehicle)

	// Adding in the current location and marking the InService boolean to true
	vehicle.CurrentLocation = args[1]
	vehicle.InService = true

	// Putting this back into JSON.
	vehicleAsBytes, _ = json.Marshal(vehicle)

	// Putting that JSON for the updated vehicle info on the chain, capturing any error.
	err := APIstub.PutState(recordKey, vehicleAsBytes)

	// Got some logging happening here...
	var logger = shim.NewLogger("vehicleReg")
	logger.Info("Lance Debug - This is the vehicleAsBytes...")
	logger.Info(vehicleAsBytes)

	// If there is any error from creating a record for this on the chain, exit on that error.
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record the vehicle: %s", args[0]))
	}

	// If all went well, return success.
	return shim.Success(nil)
}

// updateLocation to change the location for a registered vehicle. THIS MIGHT BE ABLE TO VANISH...
func (s *SmartContract) updateLocation(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Returning an error if there aren't the right number of arguments, the serial number (args[0]), and current location(args[1])
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. The vehicle serial number and its current location.")
	}
	// Get the current state for the vehicle.
	vehicleAsBytes, _ := APIstub.GetState(args[0])
	// Make sure we got a result. If we didn't, throw an error.
	if vehicleAsBytes == nil {
		return shim.Error("Could not locate Vehicle with that serial number")
	}
	// Setting up a new empty Vehicle struct.
	vehicle := Vehicle{}
	// Unmarshalling/parsing the vehicle state into the struct.
	json.Unmarshal(vehicleAsBytes, &vehicle)
	// Adding in the current location and marking the InService boolean to true
	vehicle.CurrentLocation = args[1]
	// Putting this back into JSON.
	vehicleAsBytes, _ = json.Marshal(vehicle)
	// Putting that JSON for the updated vehicle info on the chain, capturing any error.
	err := APIstub.PutState(args[0], vehicleAsBytes)
	// Got some logging happening here...
	var logger = shim.NewLogger("vehicleReg")
	logger.Info("Lance Debug - This is the vehicleAsBytes...")
	logger.Info(vehicleAsBytes)
	// If there is any error from creating a record for this on the chain, exit on that error.
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record the vehicle: %s", args[0]))
	}
	// If all went well, return success.
	return shim.Success(nil)
}

// getAllTripsAllVehicles allows us to see all the vehicles registered on this
// ledger.
func (s *SmartContract) getAllTripsAllVehicles(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	startKey := "T"
	endKey := "Tzzzzzzzzzzz"
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
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		//    buffer.WriteString(string(queryResponse.Value))
		resultsIterator2, err := APIstub.GetHistoryForKey(queryResponse.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		// Begin writing an array of the vehicle trip history.
		buffer.WriteString("[")
		bArrayMemberAlreadyWritten2 := false
		for resultsIterator2.HasNext() {
			tripResponse, err := resultsIterator2.Next()
			// If there is an error, throw it.
			if err != nil {
				return shim.Error(err.Error())
			}
			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten2 == true {
				buffer.WriteString(",")
			}
			// Write the transaction ID.
			buffer.WriteString("{\"TxId\":")
			buffer.WriteString("\"")
			buffer.WriteString(tripResponse.TxId)
			buffer.WriteString("\"")
			// Write the value for the string.
			buffer.WriteString(", \"Value\":")
			// If there was a delete operation on given key, then we need to set the
			// corresponding value null. Otherwise, we will write the response.Value
			// as-is (as the Value itself a JSON for the trip.)
			if tripResponse.IsDelete {
				buffer.WriteString("null")
			} else {
				buffer.WriteString(string(tripResponse.Value))
			}
			// Write the timestamp when the trip was committed.
			buffer.WriteString(", \"Timestamp\":")
			buffer.WriteString("\"")
			buffer.WriteString(time.Unix(tripResponse.Timestamp.Seconds, int64(tripResponse.Timestamp.Nanos)).String())
			buffer.WriteString("\"")
			// Close out the item
			buffer.WriteString("}")
			// Set the boolean to true so we know that at least one item has been written and to include commas in the future.
			bArrayMemberAlreadyWritten2 = true
		}
		buffer.WriteString("]")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

// recordTrip will record a new trip <==== PICK THIS UP TOMORROW.
func (s *SmartContract) recordTrip(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Making sure we have the right number of arguments provided, otherwise provide error.
	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}
	// Creating a logger to get some debug info logged out.
	var logger = shim.NewLogger("tripLogger")

	// Get the vehicle record.
	key := "V" + args[0]
	_, err := APIstub.GetState(key)
	// If there is any error getting the key,
	if err != nil {
		return shim.Error(err.Error())
	}

	// Convert the strings from the initial JSON into their correct values with strconv.

	theStartTime, _ := strconv.ParseInt(args[1], 10, 64)
	theEndTime, _ := strconv.ParseInt(args[4], 10, 64)
	theStartLat, _ := strconv.ParseFloat(args[2], 64)
	theStartLong, _ := strconv.ParseFloat(args[3], 64)
	theEndLat, _ := strconv.ParseFloat(args[5], 64)
	theEndLong, _ := strconv.ParseFloat(args[6], 64)

	// Put the trip info into a Trip struct.
	var singleTrip = Trip{Serial: args[0], StartTime: theStartTime, StartLat: theStartLat, StartLong: theStartLong, EndTime: theEndTime, EndLat: theEndLat, EndLong: theEndLong}
	// Marshalling this into JSON for recording on the chain.
	singleTripAsBytes, _ := json.Marshal(singleTrip)
	// Some debug logging.
	logger.Info("Lance Debug - This is the singleTripAsBytes...")
	logger.Info(singleTripAsBytes)
	// Putting the new trip into the database. Catching any error that may occur.
	err = APIstub.PutState(args[0], singleTripAsBytes)
	// If there was an error, return that error
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record trip. Error: %s", err))
	}
	// If everything worked, record a success.
	return shim.Success(nil)
}

func (t *SmartContract) getTripsForVehicle(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Making sure we have at least one argument, the vehicle serial number
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 argument, the vehicle serial number.")
	}
	// Putting the key into a variable for ease-of-use.
	recordKey := args[0]
	// Getting the history for the record. Catching any potential error.
	resultsIterator, err := stub.GetHistoryForKey(recordKey)
	// If there is an error, return that error.
	if err != nil {
		return shim.Error(err.Error())
	}
	// Close out the resultsIterator at the end of the function.
	defer resultsIterator.Close()
	// buffer is a JSON array containing historic values for the key/value pair
	var buffer bytes.Buffer
	buffer.WriteString("[")
	// This boolean is to let us know if we're in the first item in the record, for formatting purposes.
	bArrayMemberAlreadyWritten := false
	// Loop through the results.
	for resultsIterator.HasNext() {
		// Grab the next item and any error that may come up.
		response, err := resultsIterator.Next()
		// If there is an error, throw it.
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		// Write the transaction ID.
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")
		// Write the value for the string.
		buffer.WriteString(", \"Value\":")
		// If there was a delete operation on given key, then we need to set the
		// corresponding value null. Otherwise, we will write the response.Value
		// as-is (as the Value itself a JSON for the trip.)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}
		// Write the timestamp when the trip was committed.
		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")
		// Write down if the trip was deleted.
		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")
		// Close out the item
		buffer.WriteString("}")
		// Set the boolean to true so we know that at least one item has been written and to include commas in the future.
		bArrayMemberAlreadyWritten = true
	}
	// Close out the JSON for the results.
	buffer.WriteString("]")
	// Creating a logger to get some debug info logged out.
	var logger = shim.NewLogger("tripLogger")
	// Writing the results in the logger.
	logger.Info("- getHistoryForRecord returning:\n%s\n", buffer.String())
	// Return the buffer we've written all of this information to.
	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryLastTrip(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Making sure that we have the right number of arguments.
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	// Put the serial number into a variable.
	key := args[0]
	// Grab the state of the key/serial number, which should be the last trip.
	tripAsBytes, _ := APIstub.GetState(key)
	// If the state is nil, return an error.
	if tripAsBytes == nil {
		return shim.Error("Could not locate a trip for this serial number.")
	}
	// Return that trip.
	return shim.Success(tripAsBytes)
}
