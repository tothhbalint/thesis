package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

func (s *SmartContract) CrossInvoke(ctx contractapi.TransactionContextInterface) (string, error) {
	chaincodeName := "chaincode1"
	functionName := "put"
	args := []string{"name", "Willy Wonka"}

	result := ctx.GetStub().InvokeChaincode(chaincodeName, [][]byte{[]byte(functionName), []byte(args[0]), []byte(args[1])}, "my-channel1")
	if result.Status != 200 {
		return "", fmt.Errorf("failed to invoke chaincode. status: %d, message: %s", result.Status, result.Message)
	}
	return string(result.Payload), nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create cross_invoker chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting cross_invoker chaincode: %s", err.Error())
	}
}
