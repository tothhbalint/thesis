#!/bin/bash

# Step 1: Check if at least the function name is provided
if [ $# -lt 1 ]; then
  echo "Usage: $0 <function> [<arg1> <arg2> ...]"
  exit 1
fi

# Step 2: Set the function to the first argument
FUNCTION=$1

# Step 3: Collect the rest of the arguments into an array
ARGS=("${@:2}")

# Step 4: Create a JSON array of arguments, ensuring each argument is wrapped in quotes
ARGS_JSON="["
for arg in "${ARGS[@]}"; do
  ARGS_JSON+="\"$arg\","
done
ARGS_JSON="${ARGS_JSON%,}]"  # Remove the last comma and close the JSON array

cd fabric-samples/test-network

# Step 5: Set the environment variables
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

# Step 6: Run the chaincode invoke command with function and arguments
peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.example.com \
  --tls \
  --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" \
  -C mychannel \
  -n peer \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" \
  -c "{\"function\":\"$FUNCTION\",\"Args\":$ARGS_JSON}"

# Step 7: Print the result
echo "Chaincode invoke completed for function: $FUNCTION with arguments: $ARGS_JSON"