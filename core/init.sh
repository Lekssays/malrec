#!/bin/sh

ipfs init

export MONITOR="/monitor"
export BACKUP="/backup"

initial_backup_time=$(date +%s)
initial_backup_file=$initial_backup_time".tar.gz"
tar cvzf ${BACKUP}/${initial_backup_file} ${MONITOR}
file_to_upload=${BACKUP}/${initial_backup_file}
echo "File to upload = ${file_to_upload}"
hash=$(ipfs add -Q -r ${file_to_upload})
echo "File hash: $hash"
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"createBackup","Args":["'"$CORE_PEER_ID"'","'"$hash"'"]}'
