package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/common/chaincode"
	"github.com/hyperledger/fabric/core/chaincode"
)

type SmartContract struct {
	contractapi.Contract
}

func (s *SmartContract) Add(ctx contractapi.TransactionContextInterface, key string, value string) (string, error) {
	args := make([][]byte, 3)
	args[0] = []byte("put")
	args[1] = []byte(key)
	args[2] = []byte(value)
	result := ctx.GetStub().InvokeChaincode("version_cc", args, "main-channel")
	if result.Status != 200 {
		return "", fmt.Errorf("failed to invoke chaincode. status: %d,message: %s", result.Status, result.Message)
	}

	return string(result.Payload), nil
}

func (s *SmartContract) Check(ctx contractapi.TransactionContextInterface) error {
	result := "OK"
	return ctx.GetStub().SetEvent("transaction_success", []byte(result))
}

func main() {
	chaincodeSupport := &chaincode.ChaincodeSupport()
	chaincodeSupport.
		cs, err := chaincode.ChaincodeSupport()
	if err != nil {
		fmt.Printf("Error create proxy chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting proxy chaincode: %s", err.Error())
	}
}
