package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Put stores a key-value pair in the world state and returns a success message
func (s *SmartContract) Put(ctx contractapi.TransactionContextInterface, key string, value string) (string, error) {
	err := ctx.GetStub().PutState(key, []byte(value))
	if err != nil {
		return "", fmt.Errorf("failed to put state: %v", err)
	}

	// Create success response
	resp := map[string]string{"success": "OK"}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %v", err)
	}

	return string(respBytes), nil
}

// Get retrieves the value from the ledger using the provided key
func (s *SmartContract) Get(ctx contractapi.TransactionContextInterface, key string) (map[string]string, error) {
	// Get the state (value) associated with the key
	value, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get state for key %s: %s", key, err)
	}
	// If the value does not exist or is empty, return an error
	if value == nil || len(value) == 0 {
		return map[string]string{"error": "NOT_FOUND"}, nil
	}

	// Otherwise, return the value as a string in the map
	return map[string]string{"success": string(value)}, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create version chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting version chaincode: %s", err.Error())
	}
}
