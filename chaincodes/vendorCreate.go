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

// SmartContract is an empty struct for holding
type SmartContract struct {
}

// Vendor is a struct for the vendor info to be entered in EBS + a boolean if they are updated.
type Vendor struct {
	VendorID   string `json:vendorID` // Change from below here.
	VendorName string `json:owner`
	DUNS       string `json:duns`
	Approved   bool   `json:approved`
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
	if function == "registerVendor" {
		return s.registerVendor(APIstub, args)
	} else if function == "approveVendor" {
		return s.approveVendor(APIstub, args)
	} else if function == "seeAllVendors" {
		return s.seeAllVendors(APIstub, args)
	} else if function == "getVendorHistory" {
		return s.getVendorHistory(APIstub, args)
	}
	// If none of the handler functions were invoked, send back an error.
	return shim.Error("Invalid Smart Contract function name.")
}

// registerVendor is the function to register a new vendor on the blockchain.
func (s *SmartContract) registerVendor(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	////////////////////// FIX THIS..... //////////////////////////
	// Returning an error if there aren't the right number of arguments.
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	// Fill out vendor struct.
	var identity = Vendor{VendorID: args[0], VendorName: args[1], DUNS: args[2], Approved: false}
	// json.Marshal is returning a JSON encoding of the identity struct.
	identityAsBytes, _ := json.Marshal(identity)
	// Put the record on the chain, and catch any error returned.
	err := APIstub.PutState(args[0], identityAsBytes)

	// If there is any error from creating a record for this on the chain, exit on that error.
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record the vendor: %s", args[0]))
	}
	// If all went well, return success.
	return shim.Success(nil)
}

// approveVendor is the function to approve or deny a vendor. If vendor is approved, we emit an event to the OIC API.
func (s *SmartContract) approveVendor(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// Returning an error if we don't have just 1 argument, the serial
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1. The vendor serial number.")
	}

	// Get the current state for the vendor.
	vendorAsBytes, _ := APIstub.GetState(args[0])
	// Make sure we got a result. If we didn't, throw an error.
	if vendorAsBytes == nil {
		return shim.Error("Could not locate vendor with that serial number")
	}

	// Setting up a new empty Vendor struct.
	vendor := Vendor{}
	// Unmarshalling/parsing the Vendor state into the struct.
	json.Unmarshal(vendorAsBytes, &vendor)
	// Adding in the current location and marking the InService boolean to true
	vendor.Approved = true

	// Putting that JSON for the updated vendor info on the chain, capturing any error.
	err := APIstub.PutState(args[0], vendorAsBytes)

	// If there is any error from creating a record for this on the chain, exit on that error.
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record the approval: %s", args[0]))
	}

	// Send the event to OIC
	err = APIstub.SetEvent("VendorRegistered", vendorAsBytes)
	///=================================================
	return shim.Success(nil)
}

// Function to see all the vendors registered on this ledger.
func (s *SmartContract) seeAllVendors(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

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

// Here we can get the history for all revisions for a particular vendor.
func (t *SmartContract) getVendorHistory(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	// Make sure there is at least one argument, the vendor serial number.
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Provide vendor serial number.")
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
	// This boolean is used to let us know if we're on the first item for vendor.
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
