package grpc_handler

import (
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

func (v *Version) handlePutState(msg *pb.ChaincodeMessage) *pb.ChaincodeMessage {
	putState := &pb.PutState{}
	if err := proto.Unmarshal(msg.Payload, putState); err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId}
	}
	log.Printf("Putting state: %s,%s", putState.Key, putState.Value)

	if putState.Collection != "" {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte("Private data isn't supported")}
	}

	if _, ok := v.rwSet.writeSet[putState.Key]; !ok {
		v.rwSet.writeSet[putState.Key] = NewState(putState.Value, nil)
	} else {
		v.rwSet.writeSet[putState.Key].Value = putState.Value
	}

	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: msg.Txid, ChannelId: msg.ChannelId}
}

func (v *Version) handleGetState(msg *pb.ChaincodeMessage, stub shim.ChaincodeStubInterface) *pb.ChaincodeMessage {
	getState := &pb.GetState{}
	if err := proto.Unmarshal(msg.Payload, getState); err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}

	if getState.Collection != "" {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte("Private data isn't supported")}
	}
	var res []byte
	if state, ok := v.rwSet.writeSet[getState.Key]; ok {
		res = state.Value
	} else {
		res = v.readCache.CheckRead(getState.Key)
	}

	if res == nil {
		val, err := stub.GetState(getState.Key)
		if err != nil {
			return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
		}
		v.readCache.AddRead(getState.Key, val)
		res = val
	} else {
		log.Printf("Got from cache: key :%s, val: %s", getState.GetKey(), res)
	}

	log.Printf("Got key:%s, with a response: %s", getState.GetKey(), string(res))

	v.rwSet.readSet[getState.GetKey()] = struct{}{}

	log.Printf("%s", res)

	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: res}
}

func (v *Version) handleDelState(msg *pb.ChaincodeMessage) *pb.ChaincodeMessage {
	delState := &pb.DelState{}
	if err := proto.Unmarshal(msg.Payload, delState); err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}

	if delState.Collection != "" {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte("Private data isn't supported")}
	}
	v.rwSet.writeSet[delState.Key] = NewDeletedState()
	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: msg.Txid, ChannelId: msg.ChannelId}
}

func (v *Version) handleGetSateByRange(msg *pb.ChaincodeMessage, stub shim.ChaincodeStubInterface) *pb.ChaincodeMessage {
	getStateByRange := &pb.GetStateByRange{}
	if err := proto.Unmarshal(msg.Payload, getStateByRange); err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}

	if getStateByRange.Collection != "" {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte("Private data isn't supported")}
	}

	metadata := &pb.QueryMetadata{}
	proto.Unmarshal(getStateByRange.GetMetadata(), metadata)
	var res shim.StateQueryIteratorInterface
	var err error
	var res_metadata *pb.QueryResponseMetadata
	if metadata.PageSize != 0 {
		if validateSimpleKeys(getStateByRange.StartKey, getStateByRange.EndKey) {
			objecttype, attributes, err := stub.SplitCompositeKey(getStateByRange.StartKey)
			if err != nil {
				return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
			}
			res, res_metadata, err = stub.GetStateByPartialCompositeKeyWithPagination(objecttype, attributes, metadata.PageSize, metadata.Bookmark)
		} else {
			res, res_metadata, err = stub.GetStateByRangeWithPagination(getStateByRange.StartKey, getStateByRange.EndKey, metadata.PageSize, metadata.Bookmark)
		}
	} else {
		if validateSimpleKeys(getStateByRange.StartKey, getStateByRange.EndKey) {
			objecttype, attributes, err := stub.SplitCompositeKey(getStateByRange.StartKey)
			if err != nil {
				return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
			}
			res, err = stub.GetStateByPartialCompositeKey(objecttype, attributes)
		} else {
			res, err = stub.GetStateByRange(getStateByRange.StartKey, getStateByRange.EndKey)
		}
	}
	iterator_id := uuid.New().String()
	v.iters[iterator_id] = &res
	queryResponse, err := v.createQueryResponse(res)
	queryResponse.Id = iterator_id
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	if res_metadata != nil {
		queryResponse.Metadata, err = proto.Marshal(res_metadata)
		log.Print(string(queryResponse.Metadata))
		if err != nil {
			return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
		}
	}

	result, err := proto.Marshal(queryResponse)
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}

	}
	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: result, Txid: msg.Txid, ChannelId: msg.ChannelId}
}

func (v *Version) createQueryResponse(iter shim.StateQueryIteratorInterface) (*pb.QueryResponse, error) {
	queryResponse := &pb.QueryResponse{}
	if iter.HasNext() {
		query_res, err := iter.Next()
		if err != nil {
			return nil, err
		}
		query_bytes, err := proto.Marshal(query_res)
		query_result_bytes := pb.QueryResultBytes{ResultBytes: query_bytes}
		queryResponse.Results = append(queryResponse.Results, &query_result_bytes)
		if err != nil {
			return nil, err
		}
		v.rwSet.readSet[query_res.Key] = struct{}{}
		v.readCache.AddRead(query_res.Key, query_res.Value)
	}
	queryResponse.HasMore = iter.HasNext()
	return queryResponse, nil
}

// Return 0 payload response, as there is no query context
func (v *Version) handleQueryStateClose(msg *pb.ChaincodeMessage) *pb.ChaincodeMessage {
	queryStateClose := &pb.QueryStateClose{}
	if err := proto.Unmarshal(msg.Payload, queryStateClose); err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	iter, ok := v.iters[queryStateClose.Id]
	if ok {
		(*iter).Close()
		delete(v.iters, queryStateClose.Id)
	}
	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: msg.Txid, ChannelId: msg.ChannelId}
}

func (v *Version) handleQueryStateNext(msg *pb.ChaincodeMessage) *pb.ChaincodeMessage {
	queryStateNext := &pb.QueryStateNext{}
	if err := proto.Unmarshal(msg.Payload, queryStateNext); err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	queryResponse, err := v.createQueryResponse(*v.iters[queryStateNext.Id])
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	queryResponse.Id = queryStateNext.Id
	result, err := proto.Marshal(queryResponse)
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: result, Txid: msg.Txid, ChannelId: msg.ChannelId}
}

func (v *Version) handleGetStateMetadata(msg *pb.ChaincodeMessage, stub shim.ChaincodeStubInterface) *pb.ChaincodeMessage {
	getStateMetadata := &pb.GetStateMetadata{}
	if err := proto.Unmarshal(msg.Payload, getStateMetadata); err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	var metadata []byte
	var err error
	if getStateMetadata.Collection == "" {
		metadata, err = stub.GetStateValidationParameter(getStateMetadata.Key)
	} else {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte("Private data isn't supported")}
	}
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}

	// Send response msg back to chaincode. GetState will not trigger event
	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: metadata, Txid: msg.Txid, ChannelId: msg.ChannelId}
}

// This should be voted on - Or the version chaincodes shouldn't have access to it
// TODO: Use the new state struct
func (v *Version) handlePutStateMetadata(msg *pb.ChaincodeMessage) *pb.ChaincodeMessage {
	PutStateMetadata := &pb.PutStateMetadata{}
	if err := proto.Unmarshal(msg.Payload, PutStateMetadata); err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	metadata, err := proto.Marshal(PutStateMetadata.Metadata)
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	if PutStateMetadata.Collection == "" {
		if _, ok := v.rwSet.writeSet[PutStateMetadata.Key]; !ok {
			v.rwSet.writeSet[PutStateMetadata.Key] = NewState(nil, metadata)
		} else {
			v.rwSet.writeSet[PutStateMetadata.Key].Metadata = metadata
		}
	} else {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte("Private data isn't supported")}
	}

	// Send response msg back to chaincode. GetState will not trigger event
	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: msg.Txid, ChannelId: msg.ChannelId}
}

func (v *Version) handleGetQueryResult(msg *pb.ChaincodeMessage, stub shim.ChaincodeStubInterface) *pb.ChaincodeMessage {
	getQueryResult := &pb.GetQueryResult{}
	if err := proto.Unmarshal(msg.Payload, getQueryResult); err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}

	if getQueryResult.Collection != "" {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte("Private data isn't supported")}
	}

	res, err := stub.GetQueryResult(getQueryResult.Query)
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	iterator_id := uuid.New().String()
	v.iters[iterator_id] = &res
	queryResponse, err := v.createQueryResponse(res)
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	queryResponse.Id = iterator_id
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}

	result, err := proto.Marshal(queryResponse)
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}

	}
	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: result, Txid: msg.Txid, ChannelId: msg.ChannelId}
}

func (v *Version) handleGetHistoryForKey(msg *pb.ChaincodeMessage, stub shim.ChaincodeStubInterface) *pb.ChaincodeMessage {
	getHistoryForKey := &pb.GetHistoryForKey{}
	if err := proto.Unmarshal(msg.GetPayload(), getHistoryForKey); err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}

	res, err := stub.GetHistoryForKey(getHistoryForKey.GetKey())
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	defer res.Close()
	queryResponse := &pb.QueryResponse{}
	for res.HasNext() {
		query_res, err := res.Next()
		if err != nil {
			return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
		}
		query_bytes, err := proto.Marshal(query_res)
		query_result_bytes := pb.QueryResultBytes{ResultBytes: query_bytes}
		queryResponse.Results = append(queryResponse.Results, &query_result_bytes)
		if err != nil {
			return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
		}
	}
	result, err := proto.Marshal(queryResponse)
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}

	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: []byte(err.Error())}
	}
	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: result, Txid: msg.Txid, ChannelId: msg.ChannelId}

}

func validateSimpleKeys(simpleKeys ...string) bool {
	for _, key := range simpleKeys {
		if len(key) > 0 && key[0] == "\x00"[0] {
			return true
		}
	}
	return false
}
