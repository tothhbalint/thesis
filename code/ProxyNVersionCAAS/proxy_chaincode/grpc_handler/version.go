package grpc_handler

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// State struct to handle putstate
type State struct {
	Value    []byte
	Deleted  bool
	Metadata []byte
}

func NewState(value []byte, metadata []byte) *State {
	return &State{Value: value, Deleted: false, Metadata: metadata}
}

func NewDeletedState() *State {
	return &State{Value: nil, Deleted: true, Metadata: nil}
}

func (s *State) compareStates(other *State) bool {
	if !bytes.Equal(s.Value, other.Value) {
		log.Printf("Value doesn't equal")
		return false
	} else if s.Deleted != other.Deleted {
		log.Printf("Deleted doesn't equal")
		return false
	} else if !bytes.Equal(s.Metadata, other.Metadata) {
		log.Printf("Metadata doesn't equal")
		return false
	}
	return true
}

type RWSet struct {
	readSet  map[string]struct{}
	writeSet map[string]*State
	// TODO : privateReadSet map[string]string (collection, key)
	// TODO : privateWriteSet map[string]PrivateState
}

func NewRWSet() *RWSet {
	return &RWSet{readSet: make(map[string]struct{}), writeSet: make(map[string]*State)}
}

func (rwSet *RWSet) CompareRWSets(other *RWSet) bool {
	for read := range other.readSet {
		if _, ok := rwSet.readSet[read]; !ok {
			return false
		}
	}

	for writeKey, writeValue := range other.writeSet {
		val, ok := rwSet.writeSet[writeKey]
		if !ok || !val.compareStates(writeValue) {
			return false
		}
	}
	return true
}

type Version struct {
	stream          pb.ChaincodeSupport_RegisterServer
	responseChannel chan *pb.ChaincodeMessage
	rwSet           *RWSet
	readCache       *ReadCache
	iters           map[string]*shim.StateQueryIteratorInterface
}

func NewVersion(stream pb.ChaincodeSupport_RegisterServer) *Version {
	return &Version{stream: stream, responseChannel: make(chan *pb.ChaincodeMessage), iters: make(map[string]*shim.StateQueryIteratorInterface)}
}

func (v *Version) Invoke(msg *pb.ChaincodeMessage, stub shim.ChaincodeStubInterface, readCache *ReadCache) ([]byte, error) {
	ch := v.responseChannel
	v.rwSet = NewRWSet()
	v.readCache = readCache

	if err := v.stream.Send(msg); err != nil {
		return nil, fmt.Errorf("failed sending message")
	}
	for {
		select {
		case msg = <-ch:
			switch msg.Type {
			case pb.ChaincodeMessage_COMPLETED:
				return msg.Payload, nil
			case pb.ChaincodeMessage_PUT_STATE:
				v.stream.Send(v.handlePutState(msg))
			case pb.ChaincodeMessage_GET_STATE:
				v.stream.Send(v.handleGetState(msg, stub))
			case pb.ChaincodeMessage_DEL_STATE:
				v.stream.Send(v.handleDelState(msg))
			case pb.ChaincodeMessage_GET_STATE_BY_RANGE:
				v.stream.Send(v.handleGetSateByRange(msg, stub))
			case pb.ChaincodeMessage_QUERY_STATE_CLOSE:
				v.stream.Send(v.handleQueryStateClose(msg))
			case pb.ChaincodeMessage_PUT_STATE_METADATA:
				v.stream.Send(v.handlePutStateMetadata(msg))
			case pb.ChaincodeMessage_GET_STATE_METADATA:
				v.stream.Send(v.handleGetStateMetadata(msg, stub))
			case pb.ChaincodeMessage_GET_QUERY_RESULT:
				v.stream.Send(v.handleGetQueryResult(msg, stub))
			case pb.ChaincodeMessage_GET_HISTORY_FOR_KEY:
				v.stream.Send(v.handleGetHistoryForKey(msg, stub))
			case pb.ChaincodeMessage_QUERY_STATE_NEXT:
				v.stream.Send(v.handleQueryStateNext(msg))
			}
		case <-time.After(5 * time.Second):
			return nil, fmt.Errorf("transaction timed out")
		}
	}
}
