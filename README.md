# malrec
A Blockchain-based Framework for Malware Recovery

## Getting Started

### Prerequisits
- Make sure `cryptogen` and `configtxgen` are added to your PATH.

### Run the system
- Run the command: `./system.sh up`
- To display other options: `./system.sh`

### Supported Queries

#### Add a Backup
`$ peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"createBackup","Args":["peer0.org1.example.com","QmdXYvmSEXrA9EoFBDQJRqrYiBLF6UB5o5M3pBSM4xJMuH"]}'`

#### Get a Backup by backupID
`$ peer chaincode query -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackup","Args":["peer0.org1.example.com_1621512826"]}'`

#### Get Backups of a Specific Device by deviceID
`$ peer chaincode query -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackupsByDeviceID","Args":["peer0.org1.example.com"]}'`

#### Get Backups of a Specific Device during a Timestamp Range
`$ peer chaincode query -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackupsByTimestamps","Args":["peer0.org1.example.com", "1621522261", "1621522261"]}'`