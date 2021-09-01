#!/bin/bash
set -e

export COMPOSE_PROJECT_NAME=malrec
export PROJECT_DIRECTORY=$PWD
export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=$PROJECT_DIRECTORY/network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export ORDERER_ADDRESS=0.0.0.0:7050
export ORDERER_HOSTNAME=orderer.example.com
export ORG1_TLS_ROOTCERT_FILE=$PROJECT_DIRECTORY/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export ORG2_TLS_ROOTCERT_FILE=$PROJECT_DIRECTORY/network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export ORG3_TLS_ROOTCERT_FILE=$PROJECT_DIRECTORY/network/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
export FABRIC_CFG_PATH=$PROJECT_DIRECTORY/network/config/
export CCVERSION=1.0
export CHANNEL_NAME=mychannel
export PATH=$PROJECT_DIRECTORY/bin:$PATH

# Peers and Organizations Settings
ORGS=3
PEERS=2
ENDORSEMENT_POLICY="OR ('Org1MSP.member','Org2MSP.member', 'Org3MSP.member')"
PEERS_CERTIFICATES="--peerAddresses 0.0.0.0:1151 --tlsRootCertFiles ${PWD}/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses 0.0.0.0:1251 --tlsRootCertFiles ${PWD}/network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --peerAddresses 0.0.0.0:1351 --tlsRootCertFiles ${PWD}/network/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt"

script_path=`dirname "$0"`

if [ ! -d "$script_path/chaincodes" ]; then
  echo "chaincodes/ folder was not found."
  exit 1
fi

function printHelp() {
  echo "Usage: "
  echo "  network.sh <Mode>"
  echo "    Modes:"
  echo "      "$'\e[0;32m'up$'\e[0m' - brings up fabric orderer and peer nodes. with channel creation.
  echo "      "$'\e[0;32m'down$'\e[0m' - clears the network with docker-compose down
  echo "      "$'\e[0;32m'restart$'\e[0m' - restarts the network
  echo "      "$'\e[0;32m'clear$'\e[0m' - clears the state of the blockchain
  echo "      "$'\e[0;32m'netstat$'\e[0m' - checks network status
  echo "      "$'\e[0;32m'deployCC$'\e[0m' - installs and instantiate the chaincode
  echo "      "$'\e[0;32m'invokeCC$'\e[0m' - invokes the chaincode
  echo "      "$'\e[0;32m'createC$'\e[0m' - creates channel
}

function setVariables() {
  orgId=$1
  if [ $orgId = 10 ]; then
    port=1051
  else
    port=$((1051+orgId*100))
  fi
  export CORE_PEER_ADDRESS=0.0.0.0:${port}
  export CORE_PEER_LOCALMSPID="Org${orgId}MSP"
  export PEER_ORG_CA=$PROJECT_DIRECTORY/network/crypto-config/peerOrganizations/org${orgId}.example.com/peers/peer0.org${orgId}.example.com/tls/ca.crt
  export CORE_PEER_TLS_ROOTCERT_FILE=$PEER_ORG_CA
  export CORE_PEER_MSPCONFIGPATH=$PROJECT_DIRECTORY/network/crypto-config/peerOrganizations/org${orgId}.example.com/users/Admin@org${orgId}.example.com/msp    
}

function restartNetwork() {
  docker rm -f $(docker ps -aq) || true
  docker volume rm $(docker volume ls)  || true
}

function networkDown() {
  docker stop $(docker ps -a -q) || true
  docker rm $(docker ps -a -q) || true
}

function clearNetwork {
  docker rm -f $(docker ps -aq) || true
  docker rmi -f $(docker images -a -q) || true
  docker volume rm $(docker volume ls)  || true  
}

function generateBlocks() {
  cd $PROJECT_DIRECTORY/network/
  echo "Generating crypto-config files..."
  if [ -d "crypto-config" ]; then
    rm -rf ./crypto-config/
  fi
  cryptogen generate --config=crypto-config.yaml

  echo "Generating genesis block..."
  cd $PROJECT_DIRECTORY/network/config/
  if test -f "genesis.block"; then
      rm *.block
  fi
  if test -f "$CHANNEL_NAME.tx"; then
      rm $CHANNEL_NAME.tx
  fi
  configtxgen -outputBlock genesis.block -profile ThreeOrgsOrdererGenesis -channelID system-channel -configPath=.

  echo "Generating channel block..."
  configtxgen -profile ThreeOrgsChannel -outputCreateChannelTx $CHANNEL_NAME.tx -channelID $CHANNEL_NAME -configPath=.
}

function networkUp() {
  cd $PROJECT_DIRECTORY/network/

  docker-compose -f docker-compose.yml up -d

  echo "**********************************"
  echo "********* Network Status *********"
  echo "**********************************"

  docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

}

function checkNetworkStatus() {
  docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
}

function createChannel() {
  setVariables 1
  
  echo "Creating channel..."
  peer channel create -o $ORDERER_ADDRESS  --ordererTLSHostnameOverride $ORDERER_HOSTNAME -c $CHANNEL_NAME -f $PROJECT_DIRECTORY/network/config/$CHANNEL_NAME.tx --outputBlock $PROJECT_DIRECTORY/network/config/$CHANNEL_NAME.block --tls true --cafile $ORDERER_CA

  echo "Joining peers to channel..."
  for orgId in $(seq $ORGS);
  do
      setVariables $orgId
      peer channel join -b $PROJECT_DIRECTORY/network/config/$CHANNEL_NAME.block 
  done
}

function deployChaincode() {
  cd $PROJECT_DIRECTORY
  echo "Packaging $1 chaincode..."
  peer lifecycle chaincode package $1.tar.gz --path ./chaincodes/$1cc --lang golang --label ${1}_${CCVERSION}
  
  echo "Installing $1 chaincode on peers..."
  for orgId in $(seq $ORGS);
  do
      setVariables $orgId
      peer lifecycle chaincode install $1.tar.gz
  done
  
  echo "Exporting $1 Package ID..."
  CC_PACKAGE_ID=$(peer lifecycle chaincode queryinstalled)
  CC_PACKAGE_ID=${CC_PACKAGE_ID%,*}
  CC_PACKAGE_ID=${CC_PACKAGE_ID#*:}
  CC_PACKAGE_ID=${CC_PACKAGE_ID##* }
  export CC_PACKAGE_ID=$CC_PACKAGE_ID

  echo "Approving $1 chaincode for Organizations..."
  for orgId in $(seq $ORGS);
  do
    setVariables $orgId
    peer lifecycle chaincode approveformyorg -o $ORDERER_ADDRESS --ordererTLSHostnameOverride $ORDERER_HOSTNAME --channelID $CHANNEL_NAME --name $1 --version $CCVERSION --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile $ORDERER_CA --signature-policy "OR ('Org1MSP.member','Org2MSP.member', 'Org3MSP.member')"
  done

  echo "Check for $1 commit readiness..."
  peer lifecycle chaincode checkcommitreadiness --channelID $CHANNEL_NAME --name $1 --version $CCVERSION --sequence 1 --tls --cafile $ORDERER_CA --signature-policy "OR ('Org1MSP.member','Org2MSP.member', 'Org3MSP.member')" --output json
  
  echo "Committing $1 chaincode definition to channel..."
  peer lifecycle chaincode commit -o $ORDERER_ADDRESS --ordererTLSHostnameOverride $ORDERER_HOSTNAME --channelID $CHANNEL_NAME --name $1 --version $CCVERSION --sequence 1 --tls --cafile $ORDERER_CA $PEERS_CERTIFICATES --signature-policy "OR ('Org1MSP.member','Org2MSP.member', 'Org3MSP.member')"
  peer lifecycle chaincode querycommitted --channelID $CHANNEL_NAME --name $1 --cafile $ORDERER_CA
}

function invokeChaincode() {
  cd $PROJECT_DIRECTORY
  setVariables 1
  echo "Invoke $1 chaincode..."
  peer chaincode invoke -o $ORDERER_ADDRESS --ordererTLSHostnameOverride $ORDERER_HOSTNAME --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $1 $PEERS_CERTIFICATES -c '{"function":"InitLedger","Args":[]}'
}

function key_generation(){
  for orgId in $(seq $ORGS)
  do
    for ((peerId=0; peerId<$PEERS; peerId++));
    do
      echo "Key generation on peer$peerId.org$orgId.example.com..."
      docker exec peer$peerId.org$orgId.example.com /bin/sh -c "/bin/sh /core/key_generation.sh"
    done
  done
}

function initIPFS() {
  for orgId in $(seq $ORGS);
  do
    for ((peerId=0; peerId<$PEERS; peerId++));
    do
      echo "Installing IPFS on peer$peerId.org$orgId.example.com..."
      docker exec -d peer$peerId.org$orgId.example.com /bin/sh -c "/bin/sh /core/init.sh"
    done
  done
}


function startBackupMonitoring() {
  for orgId in $(seq $ORGS);
  do
    for ((peerId=0; peerId<$PEERS; peerId++));
    do
      echo "Starting backup engine on peer$peerId.org$orgId.example.com..."
      docker exec -d peer$peerId.org$orgId.example.com /bin/sh -c "/bin/sh /core/monitor.sh"
    done
  done
}

if [[ $# -lt 1 ]] ; then
  printHelp
  exit 0
else
  MODE=$1
  shift
fi

if [ "${MODE}" == "down" ]; then
  networkDown
elif [ "${MODE}" == "deployCC" ]; then
  deployChaincode "backup"
  deployChaincode "malware"
elif [ "${MODE}" == "invokeCC" ]; then
  invokeChaincode "backup"
  invokeChaincode "malware"
elif [ "${MODE}" == "netstat" ]; then
  checkNetworkStatus
elif [ "${MODE}" == "createC" ]; then
  createChannel
elif [ "${MODE}" == "clear" ]; then
  clearNetwork
elif [ "${MODE}" == "up" ]; then
  restartNetwork
  generateBlocks
  networkUp
  createChannel
  deployChaincode "backup"
  invokeChaincode "backup"
  deployChaincode "malware"
  invokeChaincode "malware"
  key_generation
  initIPFS
  startBackupMonitoring
else
  printHelp
  exit 1
fi
