#!/usr/bin/env bash

generateArtifacts() {
  printHeadline "Generating basic configs" "U1F913"

  printItalics "Generating crypto material for Orderer" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-orderer.yaml" "peerOrganizations/orderer.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for Org1" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-org1.yaml" "peerOrganizations/org1.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for Org2" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-org2.yaml" "peerOrganizations/org2.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for Org3" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-org3.yaml" "peerOrganizations/org3.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating genesis block for group group1" "U1F3E0"
  genesisBlockCreate "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config" "Group1Genesis"

  # Create directories to avoid permission errors on linux
  mkdir -p "$FABLO_NETWORK_ROOT/fabric-config/chaincode-packages"
  mkdir -p "$FABLO_NETWORK_ROOT/fabric-config/config"
}

startNetwork() {
  printHeadline "Starting network" "U1F680"
  (cd "$FABLO_NETWORK_ROOT"/fabric-docker && docker compose up -d)
  sleep 4
}

generateChannelsArtifacts() {
  printHeadline "Generating config for 'main-channel'" "U1F913"
  createChannelTx "main-channel" "$FABLO_NETWORK_ROOT/fabric-config" "MainChannel" "$FABLO_NETWORK_ROOT/fabric-config/config"
}

installChannels() {
  printHeadline "Creating 'main-channel' on Org1/peer0" "U1F63B"
  docker exec -i cli.org1.example.com bash -c "source scripts/channel_fns.sh; createChannelAndJoinTls 'main-channel' 'Org1MSP' 'peer0.org1.example.com:7041' 'crypto/users/Admin@org1.example.com/msp' 'crypto/users/Admin@org1.example.com/tls' 'crypto-orderer/tlsca.orderer.example.com-cert.pem' 'orderer0.group1.orderer.example.com:7030';"

  printItalics "Joining 'main-channel' on Org2/peer0" "U1F638"
  docker exec -i cli.org2.example.com bash -c "source scripts/channel_fns.sh; fetchChannelAndJoinTls 'main-channel' 'Org2MSP' 'peer0.org2.example.com:7061' 'crypto/users/Admin@org2.example.com/msp' 'crypto/users/Admin@org2.example.com/tls' 'crypto-orderer/tlsca.orderer.example.com-cert.pem' 'orderer0.group1.orderer.example.com:7030';"
  printItalics "Joining 'main-channel' on Org3/peer0" "U1F638"
  docker exec -i cli.org3.example.com bash -c "source scripts/channel_fns.sh; fetchChannelAndJoinTls 'main-channel' 'Org3MSP' 'peer0.org3.example.com:7081' 'crypto/users/Admin@org3.example.com/msp' 'crypto/users/Admin@org3.example.com/tls' 'crypto-orderer/tlsca.orderer.example.com-cert.pem' 'orderer0.group1.orderer.example.com:7030';"
}

installChaincodes() {
  if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node")" ]; then
    local version="0.0.1"
    printHeadline "Packaging chaincode 'version_cc'" "U1F60E"
    chaincodeBuild "version_cc" "node" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node" "16"
    chaincodePackage "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc" "$version" "node" printHeadline "Installing 'version_cc' for Org1" "U1F60E"
    chaincodeInstall "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
    chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc" "$version" "orderer0.group1.orderer.example.com:7030" "OutOf(2,'Org1MSP.member','Org2MSP.member','Org3MSP.member')" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
  else
    echo "Warning! Skipping chaincode 'version_cc_1' installation. Chaincode directory is empty."
    echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node'"
  fi
  if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-go")" ]; then
    local version="0.0.1"
    printHeadline "Packaging chaincode 'version_cc'" "U1F60E"
    chaincodeBuild "version_cc" "golang" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-go" "16"
    chaincodePackage "cli.org2.example.com" "peer0.org2.example.com:7061" "version_cc" "$version" "golang" printHeadline "Installing 'version_cc_2' for Org2" "U1F60E"
    printHeadline "Installing 'version_cc_2' for Org2" "U1F60E"
    chaincodeInstall "cli.org2.example.com" "peer0.org2.example.com:7061" "version_cc" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
    chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "main-channel" "version_cc" "$version" "orderer0.group1.orderer.example.com:7030" "OutOf(2,'Org1MSP.member','Org2MSP.member','Org3MSP.member')" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
  else
    echo "Warning! Skipping chaincode 'version_cc_2' installation. Chaincode directory is empty."
    echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-go'"
  fi
  if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node-faulty")" ]; then
    local version="0.0.1"
    printHeadline "Packaging chaincode 'version_cc'" "U1F60E"
    chaincodeBuild "version_cc" "node" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node-faulty" "16"
    chaincodePackage "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc" "$version" "node" printHeadline "Installing 'version_cc_3' for Org1" "U1F60E"
    chaincodeInstall "cli.org3.example.com" "peer0.org3.example.com:7081" "version_cc" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
    chaincodeApprove "cli.org3.example.com" "peer0.org3.example.com:7081" "main-channel" "version_cc" "$version" "orderer0.group1.orderer.example.com:7030" "OutOf(2,'Org1MSP.member','Org2MSP.member','Org3MSP.member')" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
    printItalics "Committing chaincode 'version_cc' on channel 'main-channel' as 'Org1'" "U1F618"
    chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc" "$version" "orderer0.group1.orderer.example.com:7030" "OutOf(2,'Org1MSP.member','Org2MSP.member','Org3MSP.member')" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" "peer0.org1.example.com:7041,peer0.org2.example.com:7061,peer0.org3.example.com:7081" "crypto-peer/peer0.org1.example.com/tls/ca.crt,crypto-peer/peer0.org2.example.com/tls/ca.crt,crypto-peer/peer0.org3.example.com/tls/ca.crt" ""
  else
    echo "Warning! Skipping chaincode 'version_cc_3' installation. Chaincode directory is empty."
    echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node-faulty'"
  fi

}

installChaincode() {
  local chaincodeName="$1"
  if [ -z "$chaincodeName" ]; then
    echo "Error: chaincode name is not provided"
    exit 1
  fi

  local version="$2"
  if [ -z "$version" ]; then
    echo "Error: chaincode version is not provided"
    exit 1
  fi

  if [ "$chaincodeName" = "version_cc_1" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node")" ]; then
      printHeadline "Packaging chaincode 'version_cc_1'" "U1F60E"
      chaincodeBuild "version_cc_1" "node" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node" "16"
      chaincodePackage "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_1" "$version" "node" printHeadline "Installing 'version_cc_1' for Org1" "U1F60E"
      chaincodeInstall "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_1" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_1" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_1' for Org2" "U1F60E"
      chaincodeInstall "cli.org2.example.com" "peer0.org2.example.com:7061" "version_cc_1" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "main-channel" "version_cc_1" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_1' for Org3" "U1F60E"
      chaincodeInstall "cli.org3.example.com" "peer0.org3.example.com:7081" "version_cc_1" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org3.example.com" "peer0.org3.example.com:7081" "main-channel" "version_cc_1" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printItalics "Committing chaincode 'version_cc_1' on channel 'main-channel' as 'Org1'" "U1F618"
      chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_1" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" "peer0.org1.example.com:7041,peer0.org2.example.com:7061,peer0.org3.example.com:7081" "crypto-peer/peer0.org1.example.com/tls/ca.crt,crypto-peer/peer0.org2.example.com/tls/ca.crt,crypto-peer/peer0.org3.example.com/tls/ca.crt" ""

    else
      echo "Warning! Skipping chaincode 'version_cc_1' install. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node'"
    fi
  fi
  if [ "$chaincodeName" = "version_cc_2" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-go")" ]; then
      printHeadline "Packaging chaincode 'version_cc_2'" "U1F60E"
      chaincodeBuild "version_cc_2" "golang" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-go" "16"
      chaincodePackage "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_2" "$version" "golang" printHeadline "Installing 'version_cc_2' for Org1" "U1F60E"
      chaincodeInstall "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_2" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_2" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_2' for Org2" "U1F60E"
      chaincodeInstall "cli.org2.example.com" "peer0.org2.example.com:7061" "version_cc_2" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "main-channel" "version_cc_2" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_2' for Org3" "U1F60E"
      chaincodeInstall "cli.org3.example.com" "peer0.org3.example.com:7081" "version_cc_2" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org3.example.com" "peer0.org3.example.com:7081" "main-channel" "version_cc_2" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printItalics "Committing chaincode 'version_cc_2' on channel 'main-channel' as 'Org1'" "U1F618"
      chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_2" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" "peer0.org1.example.com:7041,peer0.org2.example.com:7061,peer0.org3.example.com:7081" "crypto-peer/peer0.org1.example.com/tls/ca.crt,crypto-peer/peer0.org2.example.com/tls/ca.crt,crypto-peer/peer0.org3.example.com/tls/ca.crt" ""

    else
      echo "Warning! Skipping chaincode 'version_cc_2' install. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-go'"
    fi
  fi
  if [ "$chaincodeName" = "version_cc_3" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node-faulty")" ]; then
      printHeadline "Packaging chaincode 'version_cc_3'" "U1F60E"
      chaincodeBuild "version_cc_3" "node" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node-faulty" "16"
      chaincodePackage "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_3" "$version" "node" printHeadline "Installing 'version_cc_3' for Org1" "U1F60E"
      chaincodeInstall "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_3" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_3" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_3' for Org2" "U1F60E"
      chaincodeInstall "cli.org2.example.com" "peer0.org2.example.com:7061" "version_cc_3" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "main-channel" "version_cc_3" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_3' for Org3" "U1F60E"
      chaincodeInstall "cli.org3.example.com" "peer0.org3.example.com:7081" "version_cc_3" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org3.example.com" "peer0.org3.example.com:7081" "main-channel" "version_cc_3" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printItalics "Committing chaincode 'version_cc_3' on channel 'main-channel' as 'Org1'" "U1F618"
      chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_3" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" "peer0.org1.example.com:7041,peer0.org2.example.com:7061,peer0.org3.example.com:7081" "crypto-peer/peer0.org1.example.com/tls/ca.crt,crypto-peer/peer0.org2.example.com/tls/ca.crt,crypto-peer/peer0.org3.example.com/tls/ca.crt" ""

    else
      echo "Warning! Skipping chaincode 'version_cc_3' install. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node-faulty'"
    fi
  fi
}

runDevModeChaincode() {
  local chaincodeName=$1
  if [ -z "$chaincodeName" ]; then
    echo "Error: chaincode name is not provided"
    exit 1
  fi

  if [ "$chaincodeName" = "version_cc_1" ]; then
    local version="0.0.1"
    printHeadline "Approving 'version_cc_1' for Org1 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_1" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printHeadline "Approving 'version_cc_1' for Org2 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "main-channel" "version_cc_1" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printHeadline "Approving 'version_cc_1' for Org3 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org3.example.com" "peer0.org3.example.com:7081" "main-channel" "version_cc_1" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printItalics "Committing chaincode 'version_cc_1' on channel 'main-channel' as 'Org1' (dev mode)" "U1F618"
    chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_1" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer0.org1.example.com:7041,peer0.org2.example.com:7061,peer0.org3.example.com:7081" "" ""

  fi
  if [ "$chaincodeName" = "version_cc_2" ]; then
    local version="0.0.1"
    printHeadline "Approving 'version_cc_2' for Org1 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_2" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printHeadline "Approving 'version_cc_2' for Org2 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "main-channel" "version_cc_2" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printHeadline "Approving 'version_cc_2' for Org3 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org3.example.com" "peer0.org3.example.com:7081" "main-channel" "version_cc_2" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printItalics "Committing chaincode 'version_cc_2' on channel 'main-channel' as 'Org1' (dev mode)" "U1F618"
    chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_2" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer0.org1.example.com:7041,peer0.org2.example.com:7061,peer0.org3.example.com:7081" "" ""

  fi
  if [ "$chaincodeName" = "version_cc_3" ]; then
    local version="0.0.1"
    printHeadline "Approving 'version_cc_3' for Org1 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_3" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printHeadline "Approving 'version_cc_3' for Org2 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "main-channel" "version_cc_3" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printHeadline "Approving 'version_cc_3' for Org3 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org3.example.com" "peer0.org3.example.com:7081" "main-channel" "version_cc_3" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printItalics "Committing chaincode 'version_cc_3' on channel 'main-channel' as 'Org1' (dev mode)" "U1F618"
    chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_3" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer0.org1.example.com:7041,peer0.org2.example.com:7061,peer0.org3.example.com:7081" "" ""

  fi
}

upgradeChaincode() {
  local chaincodeName="$1"
  if [ -z "$chaincodeName" ]; then
    echo "Error: chaincode name is not provided"
    exit 1
  fi

  local version="$2"
  if [ -z "$version" ]; then
    echo "Error: chaincode version is not provided"
    exit 1
  fi

  if [ "$chaincodeName" = "version_cc_1" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node")" ]; then
      printHeadline "Packaging chaincode 'version_cc_1'" "U1F60E"
      chaincodeBuild "version_cc_1" "node" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node" "16"
      chaincodePackage "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_1" "$version" "node" printHeadline "Installing 'version_cc_1' for Org1" "U1F60E"
      chaincodeInstall "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_1" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_1" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_1' for Org2" "U1F60E"
      chaincodeInstall "cli.org2.example.com" "peer0.org2.example.com:7061" "version_cc_1" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "main-channel" "version_cc_1" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_1' for Org3" "U1F60E"
      chaincodeInstall "cli.org3.example.com" "peer0.org3.example.com:7081" "version_cc_1" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org3.example.com" "peer0.org3.example.com:7081" "main-channel" "version_cc_1" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printItalics "Committing chaincode 'version_cc_1' on channel 'main-channel' as 'Org1'" "U1F618"
      chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_1" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" "peer0.org1.example.com:7041,peer0.org2.example.com:7061,peer0.org3.example.com:7081" "crypto-peer/peer0.org1.example.com/tls/ca.crt,crypto-peer/peer0.org2.example.com/tls/ca.crt,crypto-peer/peer0.org3.example.com/tls/ca.crt" ""

    else
      echo "Warning! Skipping chaincode 'version_cc_1' upgrade. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node'"
    fi
  fi
  if [ "$chaincodeName" = "version_cc_2" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-go")" ]; then
      printHeadline "Packaging chaincode 'version_cc_2'" "U1F60E"
      chaincodeBuild "version_cc_2" "golang" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-go" "16"
      chaincodePackage "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_2" "$version" "golang" printHeadline "Installing 'version_cc_2' for Org1" "U1F60E"
      chaincodeInstall "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_2" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_2" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_2' for Org2" "U1F60E"
      chaincodeInstall "cli.org2.example.com" "peer0.org2.example.com:7061" "version_cc_2" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "main-channel" "version_cc_2" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_2' for Org3" "U1F60E"
      chaincodeInstall "cli.org3.example.com" "peer0.org3.example.com:7081" "version_cc_2" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org3.example.com" "peer0.org3.example.com:7081" "main-channel" "version_cc_2" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printItalics "Committing chaincode 'version_cc_2' on channel 'main-channel' as 'Org1'" "U1F618"
      chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_2" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" "peer0.org1.example.com:7041,peer0.org2.example.com:7061,peer0.org3.example.com:7081" "crypto-peer/peer0.org1.example.com/tls/ca.crt,crypto-peer/peer0.org2.example.com/tls/ca.crt,crypto-peer/peer0.org3.example.com/tls/ca.crt" ""

    else
      echo "Warning! Skipping chaincode 'version_cc_2' upgrade. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-go'"
    fi
  fi
  if [ "$chaincodeName" = "version_cc_3" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node-faulty")" ]; then
      printHeadline "Packaging chaincode 'version_cc_3'" "U1F60E"
      chaincodeBuild "version_cc_3" "node" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node-faulty" "16"
      chaincodePackage "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_3" "$version" "node" printHeadline "Installing 'version_cc_3' for Org1" "U1F60E"
      chaincodeInstall "cli.org1.example.com" "peer0.org1.example.com:7041" "version_cc_3" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_3" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_3' for Org2" "U1F60E"
      chaincodeInstall "cli.org2.example.com" "peer0.org2.example.com:7061" "version_cc_3" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "main-channel" "version_cc_3" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printHeadline "Installing 'version_cc_3' for Org3" "U1F60E"
      chaincodeInstall "cli.org3.example.com" "peer0.org3.example.com:7081" "version_cc_3" "$version" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
      chaincodeApprove "cli.org3.example.com" "peer0.org3.example.com:7081" "main-channel" "version_cc_3" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" ""
      printItalics "Committing chaincode 'version_cc_3' on channel 'main-channel' as 'Org1'" "U1F618"
      chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "main-channel" "version_cc_3" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "crypto-orderer/tlsca.orderer.example.com-cert.pem" "peer0.org1.example.com:7041,peer0.org2.example.com:7061,peer0.org3.example.com:7081" "crypto-peer/peer0.org1.example.com/tls/ca.crt,crypto-peer/peer0.org2.example.com/tls/ca.crt,crypto-peer/peer0.org3.example.com/tls/ca.crt" ""

    else
      echo "Warning! Skipping chaincode 'version_cc_3' upgrade. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-kv-node-faulty'"
    fi
  fi
}

notifyOrgsAboutChannels() {

  printHeadline "Creating new channel config blocks" "U1F537"
  createNewChannelUpdateTx "main-channel" "Org1MSP" "MainChannel" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"
  createNewChannelUpdateTx "main-channel" "Org2MSP" "MainChannel" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"
  createNewChannelUpdateTx "main-channel" "Org3MSP" "MainChannel" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"

  printHeadline "Notyfing orgs about channels" "U1F4E2"
  notifyOrgAboutNewChannelTls "main-channel" "Org1MSP" "cli.org1.example.com" "peer0.org1.example.com" "orderer0.group1.orderer.example.com:7030" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
  notifyOrgAboutNewChannelTls "main-channel" "Org2MSP" "cli.org2.example.com" "peer0.org2.example.com" "orderer0.group1.orderer.example.com:7030" "crypto-orderer/tlsca.orderer.example.com-cert.pem"
  notifyOrgAboutNewChannelTls "main-channel" "Org3MSP" "cli.org3.example.com" "peer0.org3.example.com" "orderer0.group1.orderer.example.com:7030" "crypto-orderer/tlsca.orderer.example.com-cert.pem"

  printHeadline "Deleting new channel config blocks" "U1F52A"
  deleteNewChannelUpdateTx "main-channel" "Org1MSP" "cli.org1.example.com"
  deleteNewChannelUpdateTx "main-channel" "Org2MSP" "cli.org2.example.com"
  deleteNewChannelUpdateTx "main-channel" "Org3MSP" "cli.org3.example.com"

}

printStartSuccessInfo() {
  printHeadline "Done! Enjoy your fresh network" "U1F984"
}

stopNetwork() {
  printHeadline "Stopping network" "U1F68F"
  (cd "$FABLO_NETWORK_ROOT"/fabric-docker && docker compose stop)
  sleep 4
}

networkDown() {
  printHeadline "Destroying network" "U1F916"
  (cd "$FABLO_NETWORK_ROOT"/fabric-docker && docker compose down)

  printf "Removing chaincode containers & images... \U1F5D1 \n"
  for container in $(docker ps -a | grep "dev-peer0.org1.example.com-version_cc_1" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.org1.example.com-version_cc_1*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.org2.example.com-version_cc_1" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.org2.example.com-version_cc_1*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.org3.example.com-version_cc_1" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.org3.example.com-version_cc_1*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.org1.example.com-version_cc_2" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.org1.example.com-version_cc_2*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.org2.example.com-version_cc_2" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.org2.example.com-version_cc_2*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.org3.example.com-version_cc_2" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.org3.example.com-version_cc_2*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.org1.example.com-version_cc_3" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.org1.example.com-version_cc_3*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.org2.example.com-version_cc_3" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.org2.example.com-version_cc_3*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.org3.example.com-version_cc_3" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.org3.example.com-version_cc_3*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done

  printf "Removing generated configs... \U1F5D1 \n"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/config"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/crypto-config"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/chaincode-packages"

  printHeadline "Done! Network was purged" "U1F5D1"
}
