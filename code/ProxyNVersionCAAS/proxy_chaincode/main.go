package main

import (
	"fmt"
	"proxy_chaincode/grpc_handler"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// NVersionProxyChaincode example simple Chaincode implementation
type NVersionProxyChaincode struct {
	s grpc_handler.ProxyServer
}

func (cc *NVersionProxyChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return cc.s.Init()
}

func (cc *NVersionProxyChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//fun, args := stub.GetFunctionAndParameters()
	cc.s.Invoke("v2", stub)
	return cc.s.Invoke("v1", stub)
}

// NOTE - parameters such as ccid and endpoint information are hard coded here for illustration. This can be passed in a variety of standard ways
func main() {
	//The ccid is assigned to the chaincode on install (using the “peer lifecycle chaincode install <package>” command) for instance
	ccid := "mycc:fcbf8724572d42e859a7dd9a7cd8e2efb84058292017df6e3d89178b64e6c831"
	cc := new(NVersionProxyChaincode)
	server := &shim.ChaincodeServer{
		CCID:    ccid,
		Address: "peer:9999",
		CC:      cc,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	fakeStub := new(shim.ChaincodeStubInterface)

	cc.Init(*fakeStub)

	err := server.Start()
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}

	// for testing

	//cc := new(NVersionProxyChaincode)
	//_stub := new(shim.ChaincodeStubInterface)

	//res := cc.Init(*_stub)
	//fmt.Println(res.Message)
	//time.Sleep(1000)
	//res = cc.Invoke(*_stub)
	//fmt.Println(res.Message)
	//select {}
}
