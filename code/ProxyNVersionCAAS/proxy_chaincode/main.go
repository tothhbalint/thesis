package main

import (
	"fmt"
	"log"
	"os"
	"proxy_chaincode/grpc_handler"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// NVersionProxyChaincode example simple Chaincode implementation
type NVersionProxyChaincode struct {
	voter *grpc_handler.Voter
}

func (cc *NVersionProxyChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return pb.Response{Status: 200, Message: "Successfully initialized chaincodes!"}
}

func (cc *NVersionProxyChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	log.Printf("STARTING CC INVOKE")
	res, err := cc.voter.Invoke(stub)
	if err != nil {
		return pb.Response{Status: 500, Message: "Failed invoking chaincodes", Payload: []byte(err.Error())}
	}
	return pb.Response{Status: 200, Message: "Successfully invoked chaincodes!", Payload: res}
}

// NOTE - parameters such as ccid and endpoint information are hard coded here for illustration. This can be passed in a variety of standard ways
func main() {
	//The ccid is assigned to the chaincode on install (using the “peer lifecycle chaincode install <package>” command) for instance
	ccid := os.Getenv("CHAINCODE_ID")
	cc := new(NVersionProxyChaincode)
	server := &shim.ChaincodeServer{
		CCID:    ccid,
		Address: "0.0.0.0:9999",
		CC:      cc,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	//fakeStub := new(shim.ChaincodeStubInterface)
	cc.voter = grpc_handler.NewVoter()

	//time.Sleep(10 * time.Second)
	//cc.Invoke(*fakeStub)

	err := server.Start()
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
