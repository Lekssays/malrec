# MalRec: A Blockchain-based Malware Recovery Framework for Internet of Things

## Cite
```
@inproceedings{lekssays2022malrec,
  title={MalRec: A Blockchain-based Malware Recovery Framework for Internet of Things},
  author={Lekssays, Ahmed and Sirigu, Giorgia and Carminati, Barbara and Ferrari, Elena},
  booktitle={Proceedings of the 17th International Conference on Availability, Reliability and Security},
  pages={1--8},
  year={2022}
}
```

## Getting Started

### Prerequisites
- Install docker
- Install docker-compose
- Install golang 1.14+ and add it to your PATH
- Install npm and add it to your path

### Run the system
- Go to project directory `cd malrec`
- Run the command: `./system.sh up`
- To display other options: `./system.sh`

### Supported Queries

#### Setup
- Go to project directory `cd malrec`
- Run 
```
$ export CORE_PEER_TLS_ENABLED=true \
export CORE_PEER_LOCALMSPID='Org1MSP' \
export CORE_PEER_TLS_ROOTCERT_FILE=$(pwd)/network/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
export CORE_PEER_MSPCONFIGPATH=$(pwd)/network/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp \
export CORE_PEER_ADDRESS=0.0.0.0:1151 \
export FABRIC_CFG_PATH=$(pwd)/network/config/ \
export ORDERER_CA=$(pwd)/network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
export ORDERER_ADDRESS=0.0.0.0:7050 \
export ORDERER_HOSTNAME=orderer.example.com
```

#### Add a Policy
`$ peer chaincode invoke -o 0.0.0.0:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n policy --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"CreatePolicy","Args":["peer0.org1.example.com_policy", "3","1", "1", "1024"]}'`

#### Add a Backup
`$ peer chaincode invoke -o 0.0.0.0:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"CreateBackup","Args":["BACKUP_52892114", "peer0.org1.example.com","QmdXYvmSEXrA9EoFBDQJRqrYiBLF6UB5o5M3pBSM4xJMuH", "https://drive.google.com/zerer;https://s3.amazonaws.com/bucketmalrec/ahsdlsdps;https://peer0.org1.example.com/hsqfhqsfaz", "some signature", "600"]}'`

#### Get a Backup by backupID
`$ peer chaincode query -o 0.0.0.0:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackup","Args":["BACKUP_52892114"]}'`

#### Get Backups of a Specific Device by deviceID
`$ peer chaincode query -o 0.0.0.0:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackupsByDeviceID","Args":["peer0.org1.example.com"]}'`

#### Get Backups of a Specific Device during a Timestamp Range
`$ peer chaincode query -o 0.0.0.0:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackupsByTimestamps","Args":["peer0.org1.example.com", "1621522261", "1621522261"]}'`

#### Add a Malware
Adding a malware will automatically invalidate the backups of the corresponding device in the infection period. It perfoms a range query with timestamps where `end_timestamp = current_timestamp`  and `start_timestamp = end_timestamp - period`. `period` is added as an argument to the query which denotes the estimated duration of infection in seconds (e.g., 600s in the example below).

`$ peer chaincode invoke -o 0.0.0.0:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n malware --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"CreateMalware","Args":["MALWARE_52892114", "600","peer0.org1.example.com", "some checksum"]}'`

### Test the System with Hyperledger Caliper
#### Run a Benchmark
- Go to caliper directory: `$ cd caliper`
- Initialize a project: `$ npm init -y`
- Install caliper: `$ npm install --only=prod @hyperledger/caliper-cli@0.4.0`
- Bind it: `$ npx caliper bind --caliper-bind-sut fabric:2.1 --caliper-bind-cwd ./`
- To test queryBackup run: `$ npx caliper launch manager --caliper-workspace ./ --caliper-networkconfig networkConfig.yaml --caliper-benchconfig benchmarks/queryBackup.yaml  --caliper-fabric-gateway-enabled --caliper-flow-only-test`
- To test queryBackupsByDeviceID run: `$ npx caliper launch manager --caliper-workspace ./ --caliper-networkconfig networkConfig.yaml --caliper-benchconfig benchmarks/queryBackupsByDeviceID.yaml  --caliper-fabric-gateway-enabled --caliper-flow-only-test`
- To test queryBackupsByTimestamps run: `$ npx caliper launch manager --caliper-workspace ./ --caliper-networkconfig networkConfig.yaml --caliper-benchconfig benchmarks/queryBackupsByTimestamps.yaml  --caliper-fabric-gateway-enabled --caliper-flow-only-test`
- Check `caliper/report.html` for the results of the tests. 

## Monitoring
Prometheus is enabled in the project as a monitoring framework. In addition, Grafana is added for better visualization. You can access Prometheus at: `http://0.0.0.0:9090` and Grafana at `http://0.0.0.0:3000`.
