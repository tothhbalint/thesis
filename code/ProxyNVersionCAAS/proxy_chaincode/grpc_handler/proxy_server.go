package grpc_handler

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	serverInterval     = time.Duration(2) * time.Hour    // 2 hours - gRPC default
	serverTimeout      = time.Duration(20) * time.Second // 20 sec - gRPC default
	serverMinInterval  = time.Duration(1) * time.Minute
	connectionTimeout  = 5 * time.Second
	dialTimeout        = 10 * time.Second
	maxRecvMessageSize = 100 * 1024 * 1024 // 100 MiB
	maxSendMessageSize = 100 * 1024 * 1024 // 100 MiB
)

type ProxyServer struct {
	pb.UnimplementedChaincodeSupportServer
	versions     sync.Map
	responseChan map[string]chan *pb.ChaincodeMessage
	mu           sync.Mutex
}

func NewProxyServer() *ProxyServer {
	return &ProxyServer{
		responseChan: make(map[string]chan *pb.ChaincodeMessage),
	}
}

func (s *ProxyServer) Register(stream pb.ChaincodeSupport_RegisterServer) error {
	log.Println("Received Register stream from chaincode")

	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	if msg.Type != pb.ChaincodeMessage_REGISTER {
		return fmt.Errorf("expected REGISTER message")
	}

	var ccID pb.ChaincodeID
	if err := proto.Unmarshal(msg.Payload, &ccID); err != nil {
		return err
	}
	ccName := ccID.Name

	stream.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_REGISTERED})
	stream.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_READY})

	s.mu.Lock()
	s.versions.Store(ccName, stream)
	log.Printf("Registered: %s", ccName)
	s.responseChan[ccName] = make(chan *pb.ChaincodeMessage, 1)
	s.mu.Unlock()

	s.listen(ccName, stream)
	return nil
}

func (s *ProxyServer) listen(name string, stream pb.ChaincodeSupport_RegisterServer) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Stream closed for %s: %v", name, err)
			return
		}
		log.Printf("[%s] received: %s", name, msg.Type.String())
		s.mu.Lock()
		if ch, ok := s.responseChan[name]; ok {
			ch <- msg
		}
		s.mu.Unlock()
	}
}
func (s *ProxyServer) HasStream(chaincodeName string) (pb.ChaincodeSupport_RegisterServer, bool) {
	val, ok := s.versions.Load(chaincodeName)
	if !ok {
		return nil, false
	}
	return val.(pb.ChaincodeSupport_RegisterServer), true
}

func (s *ProxyServer) Init() pb.Response {
	s.responseChan = make(map[string]chan *pb.ChaincodeMessage)

	lis, err := net.Listen("tcp", ":2051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	serverKeepAliveParameters := keepalive.ServerParameters{
		Time:    1 * time.Minute,
		Timeout: 20 * time.Second,
	}

	var serverOpts []grpc.ServerOption
	serverOpts = append(serverOpts, grpc.KeepaliveParams(serverKeepAliveParameters))

	// Default properties follow - let's start simple and stick with defaults for now.
	// These match Fabric peer side properties. We can expose these as user properties
	// if needed

	// set max send and recv msg sizes
	serverOpts = append(serverOpts, grpc.MaxSendMsgSize(maxSendMessageSize))
	serverOpts = append(serverOpts, grpc.MaxRecvMsgSize(maxRecvMessageSize))

	//set enforcement policy
	kep := keepalive.EnforcementPolicy{
		MinTime: serverMinInterval,
		// allow keepalive w/o rpc
		PermitWithoutStream: true,
	}
	serverOpts = append(serverOpts, grpc.KeepaliveEnforcementPolicy(kep))

	//set default connection timeout
	serverOpts = append(serverOpts, grpc.ConnectionTimeout(connectionTimeout))

	grpcServer := grpc.NewServer(serverOpts...)
	log.Println("Registering")
	pb.RegisterChaincodeSupportServer(grpcServer, s)
	log.Println("Registered")

	log.Println("Peer server listening on port 2051...")

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	// Create channels to signal when each stream is registered
	v1StreamReady := make(chan bool)
	v2StreamReady := make(chan bool)

	// Goroutine to check for v1 stream registration
	go func() {
		for {
			_, ok := s.HasStream("v1")
			if ok {
				v1StreamReady <- true
				return
			}
			log.Println("Waiting for v1 to register...")
			time.Sleep(1 * time.Second)
		}
	}()

	// Goroutine to check for v2 stream registration
	go func() {
		for {
			_, ok := s.HasStream("v2")
			if ok {
				v2StreamReady <- true
				return
			}
			log.Println("Waiting for v2 to register...")
			time.Sleep(1 * time.Second)
		}
	}()

	// Wait for both streams to be registered
	select {
	case <-v1StreamReady:
		log.Println("v1 stream registered successfully.")
	case <-v2StreamReady:
		log.Println("v2 stream registered successfully.")
	}

	// Once both are registered, exit the loop
	log.Println("Both streams are registered. Continuing...")

	return pb.Response{Status: 200, Message: "gRPC server started"}
}

func (s *ProxyServer) handlePutState(msg *pb.ChaincodeMessage) *pb.ChaincodeMessage {
	putState := &pb.PutState{}
	if err := proto.Unmarshal(msg.Payload, putState); err != nil {
		return nil
	}
	log.Printf("Putting state: %s,%s", putState.Key, putState.Value)

	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: msg.Txid, ChannelId: msg.ChannelId}
}

// TODO HANDLE ERRORS
func (s *ProxyServer) handleGetState(msg *pb.ChaincodeMessage, stub shim.ChaincodeStubInterface) *pb.ChaincodeMessage {
	getState := &pb.GetState{}
	if err := proto.Unmarshal(msg.Payload, getState); err != nil {
		return nil
	}
	res, err := stub.GetState(getState.GetKey())
	if err != nil {
		return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: res}
	}

	return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: msg.Txid, ChannelId: msg.ChannelId, Payload: res}
}

func (s *ProxyServer) Invoke(chaincodeName string, stub shim.ChaincodeStubInterface) pb.Response {
	fun := "put"
	args := []string{"1", "1"}
	stub.GetChannelID()
	txID := stub.GetTxID()
	txTimeStame, err := stub.GetTxTimestamp()
	if err != nil {
		fmt.Errorf("failed getting tx timestamp")
	}
	proposal, err := stub.GetSignedProposal()
	if err != nil {
		fmt.Errorf("failed getting signed proposal")
	}

	input := &pb.ChaincodeInput{
		Args: [][]byte{[]byte(fun)},
	}
	for _, arg := range args {
		input.Args = append(input.Args, []byte(arg))
	}

	payload, err := proto.Marshal(input)
	if err != nil {
		return pb.Response{Status: 500, Message: "marshal error"}
	}

	msg := &pb.ChaincodeMessage{
		Type:      pb.ChaincodeMessage_TRANSACTION,
		Payload:   payload,
		Txid:      txID,
		Timestamp: txTimeStame,
		Proposal:  proposal,
		ChannelId: stub.GetChannelID(),
	}

	streamVal, ok := s.versions.Load(chaincodeName)
	s.mu.Lock()
	ch, chOK := s.responseChan[chaincodeName]
	s.mu.Unlock()

	if !ok || !chOK {
		return pb.Response{Status: 500, Message: "stream/channel missing"}
	}

	stream := streamVal.(pb.ChaincodeSupport_RegisterServer)
	if err := stream.Send(msg); err != nil {
		return pb.Response{Status: 500, Message: "send error"}
	}
	for {
		select {
		case msg = <-ch:
			switch msg.Type {
			case pb.ChaincodeMessage_COMPLETED:
				return pb.Response{Status: 200, Message: "Transaction completed"}
			case pb.ChaincodeMessage_PUT_STATE:
				stream.Send(s.handlePutState(msg))
			case pb.ChaincodeMessage_GET_STATE:
				stream.Send(s.handleGetState(msg, stub))
			}
		case <-time.After(5 * time.Second):
			return pb.Response{Status: 408, Message: "Timeout"}
		}
	}
}
