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
