package main

import (
	"encoding/json"
	"fmt"
	"log"

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

func (s *SmartContract) Delete(ctx contractapi.TransactionContextInterface, key string) (map[string]string, error) {
	// Check if the key exists in the world state
	existingValue, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, fmt.Errorf("failed to read state for key %s: %v", key, err)
	}
	if existingValue == nil {
		return map[string]string{"error": "NOT_FOUND"}, nil
	}

	// Delete the key from the world state
	err = ctx.GetStub().DelState(key)
	if err != nil {
		return nil, fmt.Errorf("failed to delete state for key %s: %v", key, err)
	}

	// Return success response
	return map[string]string{"success": "DELETED"}, nil
}

func (s *SmartContract) GetStateByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) (map[string][]map[string]string, error) {
	const pageSize = 1
	var bookmark string
	var results []map[string]string

	for {
		// Get an iterator for the current page
		resultsIterator, responseMetadata, err := ctx.GetStub().GetStateByRangeWithPagination(startKey, endKey, pageSize, bookmark)
		log.Print(responseMetadata.Bookmark)
		if err != nil {
			return nil, fmt.Errorf("failed to get state by range with pagination: %v", err)
		}
		defer resultsIterator.Close()

		// Iterate through the results (there will be at most `pageSize` results)
		for resultsIterator.HasNext() {
			queryResponse, err := resultsIterator.Next()
			if err != nil {
				return nil, fmt.Errorf("failed to iterate over paginated results: %v", err)
			}

			// Append the key-value pair to the results
			results = append(results, map[string]string{
				"key":   queryResponse.Key,
				"value": string(queryResponse.Value),
			})
		}

		// Check if there are more pages
		if responseMetadata.Bookmark == "" {
			break // No more pages
		}
		bookmark = responseMetadata.Bookmark // Move to next page
	}

	return map[string][]map[string]string{"success": results}, nil
}

func (s *SmartContract) Ttest(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()
	iter, err := stub.GetStateByRange("", "")
	if err != nil {
		fmt.Printf("GetStateByPartialCompositeKey error: %v\n", err)
	} else {
		defer iter.Close()
		for iter.HasNext() {
			res, err := iter.Next()
			if err != nil {
				fmt.Printf("Iterator next error: %v\n", err)
				break
			}
			fmt.Printf("PartialCompositeKey result: key=%s, value=%s\n", res.Key, string(res.Value))
		}
	}
	return nil
}

func (s *SmartContract) Test(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()

	// Simple Ledger Calls
	state, err := stub.GetState("someKey")
	fmt.Printf("GetState: %v, err: %v\n", state, err)

	err = stub.PutState("someKey", []byte("someValue"))
	fmt.Printf("PutState: err: %v\n", err)

	err = stub.DelState("someKey")
	fmt.Printf("DelState: err: %v\n", err)

	res, err := stub.GetStateByRange("", "")
	fmt.Printf("GetStateByRange: %v, err: %v\n", res, err)

	res, err = stub.GetQueryResult("{\"selector\":{\"docType\":\"someType\"}}")
	fmt.Printf("GetQueryResult: %v, err: %v\n", res, err)

	resH, err := stub.GetHistoryForKey("someKey")
	fmt.Printf("GetHistoryForKey: %v, err: %v\n", resH, err)

	index := "owner~car"
	err = stub.PutState("CAR01", []byte("someValue"))
	fmt.Printf("PutState: err: %v\n", err)

	// Composite Key usage
	ckey, err := stub.CreateCompositeKey(index, []string{"BRAD", "CAR"})
	fmt.Printf("CreateCompositeKey: %s, err: %v\n", ckey, err)

	err = stub.PutState(ckey, []byte{0x00})
	fmt.Printf("PutState: err: %v\n", err)

	objectType, attributes, err := stub.SplitCompositeKey(ckey)
	fmt.Printf("SplitCompositeKey: objectType=%s, attributes=%v, err=%v\n", objectType, attributes, err)

	// Partial Composite Key
	iter, err := stub.GetStateByPartialCompositeKey(index, []string{"BRAD"})
	if err != nil {
		fmt.Printf("GetStateByPartialCompositeKey error: %v\n", err)
	} else {
		defer iter.Close()
		for iter.HasNext() {
			res, err := iter.Next()
			if err != nil {
				fmt.Printf("Iterator next error: %v\n", err)
				break
			}
			fmt.Printf("PartialCompositeKey result: key=%s, value=%s\n", res.Key, string(res.Value))
		}
	}

	// Partial Composite Key with Pagination
	pageSize := int32(1)
	bookmark := ""
	iterPaginated, metadata, err := stub.GetStateByPartialCompositeKeyWithPagination(objectType, attributes, pageSize, bookmark)
	if err != nil {
		fmt.Printf("GetStateByPartialCompositeKeyWithPagination error: %v\n", err)
	} else {
		defer iterPaginated.Close()
		for iterPaginated.HasNext() {
			res, err := iterPaginated.Next()
			if err != nil {
				fmt.Printf("Iterator pagination next error: %v\n", err)
				break
			}
			fmt.Printf("Paginated result: key=%s, value=%s\n", res.Key, string(res.Value))
		}
		fmt.Printf("Pagination metadata: fetchedRecordsCount=%d, bookmark=%s\n", metadata.FetchedRecordsCount, metadata.Bookmark)
	}

	return nil
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
	fmt.Println("Chaincode is running")
}
