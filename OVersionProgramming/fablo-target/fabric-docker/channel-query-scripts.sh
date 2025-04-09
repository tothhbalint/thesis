#!/usr/bin/env bash

source "$FABLO_NETWORK_ROOT/fabric-docker/scripts/channel-query-functions.sh"

set -eu

channelQuery() {
  echo "-> Channel query: " + "$@"

  if [ "$#" -eq 1 ]; then
    printChannelsHelp

  elif [ "$1" = "list" ] && [ "$2" = "org1" ] && [ "$3" = "peer0" ]; then

    peerChannelList "cli.org1.example.com" "peer0.org1.example.com:7041"

  elif
    [ "$1" = "list" ] && [ "$2" = "org2" ] && [ "$3" = "peer0" ]
  then

    peerChannelList "cli.org2.example.com" "peer0.org2.example.com:7061"

  elif
    [ "$1" = "list" ] && [ "$2" = "org3" ] && [ "$3" = "peer0" ]
  then

    peerChannelList "cli.org3.example.com" "peer0.org3.example.com:7081"

  #elif

#    [ "$1" = "getinfo" ] && [ "$2" = "version-channel-1" ] && [ "$3" = "org1" ] && [ "$4" = "peer0" ]
#  then
#
#    peerChannelGetInfo "version-channel-1" "cli.org1.example.com" "peer0.org1.example.com:7041"
#
#  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "version-channel-1" ] && [ "$4" = "org1" ] && [ "$5" = "peer0" ]; then
#    TARGET_FILE=${6:-"$channel-config.json"}
#
#    peerChannelFetchConfig "version-channel-1" "cli.org1.example.com" "$TARGET_FILE" "peer0.org1.example.com:7041"
#
#  elif [ "$1" = "fetch" ] && [ "$3" = "version-channel-1" ] && [ "$4" = "org1" ] && [ "$5" = "peer0" ]; then
#    BLOCK_NAME=$2
#    TARGET_FILE=${6:-"$BLOCK_NAME.block"}
#
#    peerChannelFetchBlock "version-channel-1" "cli.org1.example.com" "${BLOCK_NAME}" "peer0.org1.example.com:7041" "$TARGET_FILE"
#
#  elif
#    [ "$1" = "getinfo" ] && [ "$2" = "version-channel-2" ] && [ "$3" = "org2" ] && [ "$4" = "peer0" ]
#  then
#
#    peerChannelGetInfo "version-channel-2" "cli.org2.example.com" "peer0.org2.example.com:7061"
#
#  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "version-channel-2" ] && [ "$4" = "org2" ] && [ "$5" = "peer0" ]; then
#    TARGET_FILE=${6:-"$channel-config.json"}
#
#    peerChannelFetchConfig "version-channel-2" "cli.org2.example.com" "$TARGET_FILE" "peer0.org2.example.com:7061"
#
#  elif [ "$1" = "fetch" ] && [ "$3" = "version-channel-2" ] && [ "$4" = "org2" ] && [ "$5" = "peer0" ]; then
#    BLOCK_NAME=$2
#    TARGET_FILE=${6:-"$BLOCK_NAME.block"}
#
#    peerChannelFetchBlock "version-channel-2" "cli.org2.example.com" "${BLOCK_NAME}" "peer0.org2.example.com:7061" "$TARGET_FILE"
#
#  elif
#    [ "$1" = "getinfo" ] && [ "$2" = "version-channel-3" ] && [ "$3" = "org3" ] && [ "$4" = "peer0" ]
#  then
#
#    peerChannelGetInfo "version-channel-3" "cli.org3.example.com" "peer0.org3.example.com:7081"
#
#  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "version-channel-3" ] && [ "$4" = "org3" ] && [ "$5" = "peer0" ]; then
#    TARGET_FILE=${6:-"$channel-config.json"}
#
#    peerChannelFetchConfig "version-channel-3" "cli.org3.example.com" "$TARGET_FILE" "peer0.org3.example.com:7081"
#
#  elif [ "$1" = "fetch" ] && [ "$3" = "version-channel-3" ] && [ "$4" = "org3" ] && [ "$5" = "peer0" ]; then
#    BLOCK_NAME=$2
#    TARGET_FILE=${6:-"$BLOCK_NAME.block"}
#
#    peerChannelFetchBlock "version-channel-3" "cli.org3.example.com" "${BLOCK_NAME}" "peer0.org3.example.com:7081" "$TARGET_FILE"
#
  elif
    [ "$1" = "getinfo" ] && [ "$2" = "main-channel" ] && [ "$3" = "org1" ] && [ "$4" = "peer0" ]
  then

    peerChannelGetInfo "main-channel" "cli.org1.example.com" "peer0.org1.example.com:7041"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "main-channel" ] && [ "$4" = "org1" ] && [ "$5" = "peer0" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "main-channel" "cli.org1.example.com" "$TARGET_FILE" "peer0.org1.example.com:7041"

  elif [ "$1" = "fetch" ] && [ "$3" = "main-channel" ] && [ "$4" = "org1" ] && [ "$5" = "peer0" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "main-channel" "cli.org1.example.com" "${BLOCK_NAME}" "peer0.org1.example.com:7041" "$TARGET_FILE"

  elif
    [ "$1" = "getinfo" ] && [ "$2" = "main-channel" ] && [ "$3" = "org2" ] && [ "$4" = "peer0" ]
  then

    peerChannelGetInfo "main-channel" "cli.org2.example.com" "peer0.org2.example.com:7061"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "main-channel" ] && [ "$4" = "org2" ] && [ "$5" = "peer0" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "main-channel" "cli.org2.example.com" "$TARGET_FILE" "peer0.org2.example.com:7061"

  elif [ "$1" = "fetch" ] && [ "$3" = "main-channel" ] && [ "$4" = "org2" ] && [ "$5" = "peer0" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "main-channel" "cli.org2.example.com" "${BLOCK_NAME}" "peer0.org2.example.com:7061" "$TARGET_FILE"

  elif
    [ "$1" = "getinfo" ] && [ "$2" = "main-channel" ] && [ "$3" = "org3" ] && [ "$4" = "peer0" ]
  then

    peerChannelGetInfo "main-channel" "cli.org3.example.com" "peer0.org3.example.com:7081"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "main-channel" ] && [ "$4" = "org3" ] && [ "$5" = "peer0" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "main-channel" "cli.org3.example.com" "$TARGET_FILE" "peer0.org3.example.com:7081"

  elif [ "$1" = "fetch" ] && [ "$3" = "main-channel" ] && [ "$4" = "org3" ] && [ "$5" = "peer0" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "main-channel" "cli.org3.example.com" "${BLOCK_NAME}" "peer0.org3.example.com:7081" "$TARGET_FILE"

  else

    echo "$@"
    echo "$1, $2, $3, $4, $5, $6, $7, $#"
    printChannelsHelp
  fi

}

printChannelsHelp() {
  echo "Channel management commands:"
  echo ""

  echo "fablo channel list org1 peer0"
  echo -e "\t List channels on 'peer0' of 'Org1'".
  echo ""

  echo "fablo channel list org2 peer0"
  echo -e "\t List channels on 'peer0' of 'Org2'".
  echo ""

  echo "fablo channel list org3 peer0"
  echo -e "\t List channels on 'peer0' of 'Org3'".
  echo ""

  echo "fablo channel getinfo version-channel-1 org1 peer0"
  echo -e "\t Get channel info on 'peer0' of 'Org1'".
  echo ""
  echo "fablo channel fetch config version-channel-1 org1 peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Org1'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> version-channel-1 org1 peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Org1'".
  echo ""

  echo "fablo channel getinfo version-channel-2 org2 peer0"
  echo -e "\t Get channel info on 'peer0' of 'Org2'".
  echo ""
  echo "fablo channel fetch config version-channel-2 org2 peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Org2'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> version-channel-2 org2 peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Org2'".
  echo ""

  echo "fablo channel getinfo version-channel-3 org3 peer0"
  echo -e "\t Get channel info on 'peer0' of 'Org3'".
  echo ""
  echo "fablo channel fetch config version-channel-3 org3 peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Org3'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> version-channel-3 org3 peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Org3'".
  echo ""

  echo "fablo channel getinfo main-channel org1 peer0"
  echo -e "\t Get channel info on 'peer0' of 'Org1'".
  echo ""
  echo "fablo channel fetch config main-channel org1 peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Org1'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> main-channel org1 peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Org1'".
  echo ""

  echo "fablo channel getinfo main-channel org2 peer0"
  echo -e "\t Get channel info on 'peer0' of 'Org2'".
  echo ""
  echo "fablo channel fetch config main-channel org2 peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Org2'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> main-channel org2 peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Org2'".
  echo ""

  echo "fablo channel getinfo main-channel org3 peer0"
  echo -e "\t Get channel info on 'peer0' of 'Org3'".
  echo ""
  echo "fablo channel fetch config main-channel org3 peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Org3'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> main-channel org3 peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Org3'".
  echo ""

}
