package grpc_handler

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type ReadCache struct {
	mu    sync.Mutex
	reads map[string][]byte
}

func NewReadCache() *ReadCache {
	return &ReadCache{reads: make(map[string][]byte)}
}

func (rc *ReadCache) AddRead(key string, value []byte) {
	rc.mu.Lock()
	rc.reads[key] = value
	rc.mu.Unlock()
}

func (rc *ReadCache) CheckRead(key string) []byte {
	rc.mu.Lock()
	val, ok := rc.reads[key]
	rc.mu.Unlock()
	if !ok {
		return nil
	}
	return val
}

type Voter struct {
	versions map[string]*Version
	mu       sync.Mutex
}

func NewVoter() *Voter {
	// TODO Hang until atleas three has registered
	v := new(Voter)
	v.versions = make(map[string]*Version)
	v.Start()
	return v
}

func craftInvokeMessage(stub shim.ChaincodeStubInterface) (*pb.ChaincodeMessage, error) {
	fun, args := stub.GetFunctionAndParameters()
	txID := stub.GetTxID()
	txTimeStame, err := stub.GetTxTimestamp()
	if err != nil {
		return nil, fmt.Errorf("failed getting tx timestamp")
	}
	proposal, err := stub.GetSignedProposal()
	if err != nil {
		return nil, fmt.Errorf("failed getting signed proposal")
	}

	input := &pb.ChaincodeInput{
		Args: [][]byte{[]byte(fun)},
	}
	for _, arg := range args {
		input.Args = append(input.Args, []byte(arg))
	}

	payload, err := proto.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("payload marshal error")
	}

	msg := &pb.ChaincodeMessage{
		Type:      pb.ChaincodeMessage_TRANSACTION,
		Payload:   payload,
		Txid:      txID,
		Timestamp: txTimeStame,
		Proposal:  proposal,
		ChannelId: stub.GetChannelID(),
	}

	return msg, nil
}

func (s *Voter) Invoke(stub shim.ChaincodeStubInterface) ([]byte, error) {
	msg, err := craftInvokeMessage(stub)
	if err != nil {
		return nil, fmt.Errorf("failed crafting invoke message: %v", err)
	}

	results := make(map[string][]byte)
	rCache := NewReadCache()

	for version_name, version := range s.versions {
		res, err := version.Invoke(msg, stub, rCache)
		if err != nil {
			return nil, fmt.Errorf("failed invoking chaincode: %s, error: %v", version_name, err)
		}
		results[version_name] = res
	}

	for _, result := range results {
		log.Print(string(result))
	}

	// Vote on results
	rwSet, result := s.CompareResults(results)
	if rwSet == nil {
		return nil, fmt.Errorf("multiple wrong RW Sets")
	}

	// Write the winning state
	for key, state := range rwSet.writeSet {
		if state.Deleted {
			log.Printf("Deleting key:%s", key)
			if err := stub.DelState(key); err != nil {
				log.Printf("Failed deleting from ledger: %v", err)
			}
		} else if state.Metadata != nil {
			log.Printf("Writing metadata")
			if err := stub.SetStateValidationParameter(key, state.Metadata); err != nil {
				log.Printf("Failed writing to ledger: %v", err)
			}
		} else {
			log.Printf("Writting key:%s, value:%s", key, state.Value)
			if err := stub.PutState(key, state.Value); err != nil {
				log.Printf("Failed writing to ledger: %v", err)
			}
		}
	}

	// cleanup the versions states
	for _, version := range s.versions {
		// remove old rwSet
		version.rwSet = nil
	}
	rCache = nil // Vote on results
	return result, nil
}

func (v *Voter) CompareResults(results map[string][]byte) (*RWSet, []byte) {
	maxVote := 0
	var majorityRWSet *RWSet
	var majorityResult []byte

	// Vote majority RWSet
	for version_name, version := range v.versions {
		if maxVote == 0 {
			majorityRWSet = version.rwSet
			majorityResult = results[version_name]
			maxVote++
		} else if majorityRWSet.CompareRWSets(version.rwSet) && bytes.Equal(majorityResult, results[version_name]) {
			maxVote++
		} else {
			maxVote--
		}
	}

	// Compare matches for the majority results
	var losingVersions []string
	for version_name, version := range v.versions {
		if !majorityRWSet.CompareRWSets(version.rwSet) || !bytes.Equal(majorityResult, results[version_name]) {
			losingVersions = append(losingVersions, version_name)
			log.Printf("losing version: %s", version_name)
		}
	}

	if len(losingVersions) > len(v.versions)/2 {
		return nil, nil
	}
	return majorityRWSet, majorityResult
}
