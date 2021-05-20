#!/bin/sh

function getBackup() {
    result=$(peer chaincode query -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackup","Args":["'"$1"'"]}')
    hash=$(echo $result | python3 -c "import sys, json; print(json.load(sys.stdin)['hash'])")
    timestamp=$(echo $result | python3 -c "import sys, json; print(json.load(sys.stdin)['timestamp'])")
    ipfs get -o $BACKUP/$timestamp"_downloaded.tar.gz" $hash
}

function getBackupsWithTimestamps() {
    result=$(peer chaincode query -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackupsByTimestamps","Args":["'"$1"'", "'"$2"'", "'"$3"'"}')
    echo $result
}

function getAllDeviceBackups() {
    result=$(peer chaincode query -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackupsByDeviceID","Args":["'"$1"'"}')
    echo $result
}

getBackup peer0.org1.example.com_1621514887
getAllDeviceBackups peer0.org1.example.com
getBackupsWithTimestamps peer0.org1.example.com 1621514887 1621514890
