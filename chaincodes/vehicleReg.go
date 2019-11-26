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

type SmartContract struct {
}

type Vehicle struct {
  Serial          string  `json:serial`
  Owner           string  `json:owner`
  VehicleType     string  `json:vehicletype`
  CurrentLocation string  `json:currentlocation`
  InService       bool    `json:inService`
}

func main() {
  err := shim.Start(new(SmartContract))
  if err != nil {
    fmt.Printf("Error starting transaction Trace chaincode: %s", err)
  }
}

func (t *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
  return shim.Success(nil)
}

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
  }
  // If none of the handler functions were invoked, send back an error.
  return shim.Error("Invalid Smart Contract function name.")
}

// Function to see all the vehicles registered on this ledger.
func (s *SmartContract) seeAllVehicles(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

  startKey := ""
  endKey := ""
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
  fmt.Printf("- queryAllTransaction:\n%s\n", buffer.String())
  return shim.Success(buffer.Bytes())
}


// Function to register the vehicle.
func (s *SmartContract) registerVehicle(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
  // Returning an error if there aren't the right number of arguments.
  if len(args) != 3 {
    return shim.Error("Incorrect number of arguments. Expecting 3")
  }
  // We get the serial number for the vehicle (which becomes the key for record), the company that owns the vehicle, and the type of vehicle. We mark its current locatio as "oos" (for out of service) and mark InService as false.
  var identity = Vehicle {Serial:args[0],Owner:args[1],VehicleType:args[2],CurrentLocation:"oos",InService:false}
  // json.Marshal is returning a JSON encoding of the identity struct.
  identityAsBytes, _ := json.Marshal(identity)
  // Put the record on the chain, and catch any error returned.
  err := APIstub.PutState(args[0], identityAsBytes)
  // Got some logging happening here...
  var logger = shim.NewLogger("vehicleReg")
  logger.Info("Lance Debug - This is the identityAsBytes...")
  logger.Info(identityAsBytes)
  // If there is any error from creating a record for this on the chain, exit on that error.
  if err != nil {
    return shim.Error(fmt.Sprintf("Failed to record the vehicle: %s", args[0]))
  }
  // If all went well, return success.
  return shim.Success(nil)
}


// Here we can get the history for all trips for a particular vehicle.
func (t *SmartContract) getHistoryForVehicle(stub shim.ChaincodeStubInterface, args []string) sc.Response {

  // Make sure there is at least one argument, the vehicle serial number.
  if len(args) < 1 {
    return shim.Error("Incorrect number of arguments. Provide vehicle serial number.")
  }
  // Taking the serial number from the arguments and making it the recordKey.
  recordKey := args[0]
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
    // Write down if this transaction was deleted or not.
    buffer.WriteString(", \"IsDelete\":")
    buffer.WriteString("\"")
    buffer.WriteString(strconv.FormatBool(response.IsDelete))
    buffer.WriteString("\"")
    // Close out the JSON of this item.
    buffer.WriteString("}")
    // Mark our first-item-used check boolean to true.
    bArrayMemberAlreadyWritten = true
  }
  // Close out the array JSON payload.
  buffer.WriteString("]")
  // Return the history buffer.
  return shim.Success(buffer.Bytes())
}

//////////////////////////////// FIX THIS
// Function to mark a registered vehicle as taken out of service.
func (s *SmartContract) takeOutOfService(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
  // Returning an error if there aren't the right number of arguments.
  if len(args) != 1 {
    return shim.Error("Incorrect number of arguments. Expecting 1, the vehicle serial number.")
  }
  // Get the current state for the vehicle.
  vehicleAsBytes, _ := APIstub.GetState(args[0])
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

// Function to get a registered vehicle put into service.
func (s *SmartContract) putInService(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
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
  vehicle.InService = true
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

// Function to get a registered vehicle put into service.
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
