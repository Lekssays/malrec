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
`$ peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"createBackup","Args":["BACKUP_52892114", "peer0.org1.example.com","QmdXYvmSEXrA9EoFBDQJRqrYiBLF6UB5o5M3pBSM4xJMuH"]}'`

#### Get a Backup by backupID
`$ peer chaincode query -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackup","Args":["BACKUP_52892114"]}'`

#### Get Backups of a Specific Device by deviceID
`$ peer chaincode query -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackupsByDeviceID","Args":["peer0.org1.example.com"]}'`

#### Get Backups of a Specific Device during a Timestamp Range
`$ peer chaincode query -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackupsByTimestamps","Args":["peer0.org1.example.com", "1621522261", "1621522261"]}'`

#### Delete a Malicious Backup from the State
`$ peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"DeleteBackup","Args":["BACKUP_52892114"]}'`

### Test the System with Hyperledger Caliper
#### Run a Benchmark
- Go to caliper directory: `$ cd caliper`
- Initialize a project: `npm init -y`
- Install caliper: `npm install --only=prod @hyperledger/caliper-cli@0.4.0`
- Bind it: `npx caliper bind --caliper-bind-sut fabric:2.1 --caliper-bind-cwd ./`
- Run: `$ npx caliper launch manager --caliper-workspace ./ --caliper-networkconfig networkConfig.yaml --caliper-benchconfig benchmarks/queryBackup.yaml  --caliper-fabric-gateway-enabled --caliper-flow-only-test`
- Check `caliper/report.html` for the results of the tests. 

## Monitoring
Prometheus is enabled in the project as a monitoring framework. In addition, Grafana is added for better visualization. You can access Prometheus at: `http://0.0.0.0:9090` and Grafaba at `http://0.0.0.0:3000`.