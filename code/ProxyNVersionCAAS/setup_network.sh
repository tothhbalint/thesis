#!/bin/bash

set -e  # Exit on error

FABRIC_SAMPLES_DIR="fabric-samples"

# Step 1: Check if fabric-samples folder already exists
if [ ! -d "$FABRIC_SAMPLES_DIR" ]; then
  echo "➡️ Downloading Fabric samples and binaries..."

  # Download the install-fabric script and Fabric binaries/samples
  curl -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh
  chmod +x install-fabric.sh
  ./install-fabric.sh binary samples
  # Step 2: Change to the test-network directory
  cd $FABRIC_SAMPLES_DIR/test-network
  
  echo "➡️ Applying test-network patch..."
  
  # Step 3: Apply your patch
  git apply ../../test-network.diff
else
  cd $FABRIC_SAMPLES_DIR/test-network
  ./network.sh down
  echo "🔹 Fabric samples folder already exists, skipping download."
fi


# Step 4: Export environment variables
echo "➡️ Setting environment variables..."

export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

# Step 5: Bring up Fabric network with CouchDB
echo "➡️ Starting Fabric network with CouchDB..."
./network.sh up createChannel -ca -s couchdb

# Step 6: Deploy chaincode
echo "➡️ Deploying Chaincode as a Service (CCaaS)..."
./network.sh deployCCAAS -ccn peer -ccp ../../proxy_chaincode

echo "✅ Fabric network is up and chaincode is deployed!"

# Step 7: (Optional) Go back to your project root or start docker compose
# cd ../..
# docker-compose up -d