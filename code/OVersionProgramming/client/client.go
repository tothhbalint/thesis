/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID        = "Org3MSP"
	cryptoPath   = "../fablo-target/fabric-config/crypto-config/peerOrganizations/org3.example.com"
	certPath     = cryptoPath + "/users/User1@org3.example.com/msp/signcerts"
	keyPath      = cryptoPath + "/users/User1@org3.example.com/msp/keystore"
	tlsCertPath  = cryptoPath + "/peers/peer0.org3.example.com/tls/ca.crt"
	peerEndpoint = "localhost:7081"
	gatewayPeer  = "peer0.org3.example.com"
)

func main() {
	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithHash(hash.SHA256),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(30*time.Second),
		client.WithEndorseTimeout(30*time.Second),
		client.WithSubmitTimeout(30*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "version_cc"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	channelName := "main-channel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	// Get a network instance
	network := gw.GetNetwork(channelName)

	contract := network.GetContract(chaincodeName)

	exampleErrorHandling(contract)
}

func newGrpcConnection() *grpc.ClientConn {
	certificatePEM, err := os.ReadFile(tlsCertPath)
	if err != nil {
		panic(fmt.Errorf("failed to read TLS certifcate file: %w", err))
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.NewClient(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificatePEM, err := readFirstFile(certPath)
	if err != nil {
		panic(fmt.Errorf("failed to read certificate file: %w", err))
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	privateKeyPEM, err := readFirstFile(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

func readFirstFile(dirPath string) ([]byte, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}

	fileNames, err := dir.Readdirnames(1)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(path.Join(dirPath, fileNames[0]))
}

// Submit transaction, passing in the wrong number of arguments ,expected to throw an error containing details of any error responses from the smart contract.
func exampleErrorHandling(contract *client.Contract) {
	proposal, err := contract.NewProposal("put", client.WithArguments("1", "1"))
	if err != nil {
		panic(fmt.Errorf("failed creating New proposal: %w", err))
	}

	evaluate_result, err := proposal.Evaluate()
	if err != nil {
		panic(fmt.Errorf("failed evaluating proposal: %w", err))
	}
	log.Print(string(evaluate_result))

	txn, err := proposal.Endorse()
	if err != nil {
		panic(fmt.Errorf("failed to endorse proposal: %w", err))
	}

	commit_result, err := txn.Submit()
	if err != nil {
		panic(fmt.Errorf("failed to submit txn: %w", err))
	}

	commit_bytes, err := commit_result.Bytes()
	if err != nil {
		panic(fmt.Errorf("failed to get commit bytes: %w", err))
	}

	commit_status, err := commit_result.Status()
	if err != nil {
		panic(fmt.Errorf("failed to get commit status: %w", err))
	}

	log.Printf("transaction result: %s, commit bytes: %s, commit status: %s", string(commit_bytes), commit_status.Code.String(), txn.Result())

	result, err := contract.SubmitTransaction("get", "1")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}
	fmt.Printf("%s", result)

	result, err = contract.SubmitTransaction("get", "asdf")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}
	fmt.Printf("%s", result)
}
