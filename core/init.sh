#!/bin/sh

ipfs init

mkdir -p /encrypted_backup

export MONITOR="/monitor"
export BACKUP="/backup"
export ENC_BACKUP="/encrypted_backup"
export PUB_KEY="/key/public_key"

COUNTER_ID=0
BACKUP_ID="${CORE_PEER_ID}:${COUNTER_ID}"
initial_backup_time=$(date +%s)
initial_backup_file=$initial_backup_time".tar.gz"
tar cvzf ${BACKUP}/${initial_backup_file} ${MONITOR}
file_to_upload=${BACKUP}/${initial_backup_file}

# Encrypt file
encrypted_file=${ENC_BACKUP}/${initial_backup_file}
echo ${encrypted_file}
eciespy -e -D ${BACKUP}/${initial_backup_file} -O ${encrypted_file} -k ${PUB_KEY}

echo "File to upload = ${encrypted_file}"
hash=$(ipfs add -Q -r ${encrypted_file})
echo "File hash: $hash"
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"createBackup","Args":["'"$BACKUP_ID"'", "'"$CORE_PEER_ID"'","'"$hash"'"]}'

