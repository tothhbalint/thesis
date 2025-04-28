package grpc_handler

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
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

func (v *Voter) Register(stream pb.ChaincodeSupport_RegisterServer) error {
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

	v.mu.Lock()
	v.versions[ccName] = NewVersion(stream)
	log.Printf("Registered: %s", ccName)
	v.mu.Unlock()

	v.listen(ccName, stream)
	return nil
}

func (v *Voter) listen(name string, stream pb.ChaincodeSupport_RegisterServer) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Stream closed for %s: %v", name, err)
			return
		}
		log.Printf("[%s] received: %s", name, msg.Type.String())
		v.mu.Lock()
		ch := v.versions[name].responseChannel
		ch <- msg
		v.mu.Unlock()
	}
}

func (v *Voter) Start() {
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
	pb.RegisterChaincodeSupportServer(grpcServer, v)
	log.Println("Registered")

	log.Println("Peer server listening on port 2051...")

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()
}
