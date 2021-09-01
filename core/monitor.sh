#!/bin/sh
export MONITOR="/monitor"
export BACKUP="/backup"
export ENC_BACKUP="/encrypted_backup"
export PUB_KEY="/key/public_key"

echo "I'm monitoring $MONITOR directory..."
COUNTER_ID=1
BACKUP_ID="${CORE_PEER_ID}:${COUNTER_ID}"
inotifywait -m -e modify ${MONITOR} |
while read path action file; do
    echo "New directory: $file appeared in directory $path via $action"
    if [[ $(expr match "$file" '.*txt$') ]];
    then
        let COUNTER_ID++
        BACKUP_ID="${CORE_PEER_ID}:${COUNTER_ID}"
        backup_time=$(date +%s)
        backup_file=$backup_time".tar.gz"
        backup_path=${BACKUP}/${backup_file}
        tar cvzf ${backup_path} ${MONITOR}/${file}
        file_to_upload=${backup_path}

        # Encrypt file
        encrypted_file=${ENC_BACKUP}/${backup_file}
        eciespy -e -D ${BACKUP}/${backup_file} -O ${encrypted_file} -k ${PUB_KEY}

        hash=$(ipfs add -Q -r ${encrypted_file}})
        echo "File hash: $hash"
        peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"createBackup","Args":["'"$BACKUP_ID"'", "'"$CORE_PEER_ID"'","'"$hash"'"]}'
    fi
done
