# Usage:
## Requirements
### Setup with script
```
./setup_network.sh
```
### Setup by hand
First get the fabric samples/binaries:
```
curl -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh && chmod +x install-fabric.sh
```
The binaries and the fabric samples are also required

Change the working directory to fabric-samples/test-network

Apply the patch to the test-network folder
```
git apply ../../test-network.diff
```

Set the export variables
```
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
```

Start the fabric network using couchdb 
```
./network.sh up createChannel -ca -s couchdb
```

Install the chaincode
```
./network deployCCAAS -ccn peer -ccp ../../proxy_chaincode
```

Now the docker compose project can be started, from the project root

## Test functions

```
./invoke.sh function args..
```

### Available functions:
- put key value
	- put a key with value
- get key
	- get a key
- test 
	- run most available stub functions, won't result in anything, but tests that every function works
