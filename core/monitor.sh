#!/bin/sh
export MONITOR="/monitor"
export BACKUP="/backup"

echo "I'm monitoring $MONITOR directory..."
inotifywait -m -e modify ${MONITOR} |
while read path action file; do
    echo "New directory: $file appeared in directory $path via $action"
    # It is performed when a .txt file is added/modified
    if [[ $(expr match "$file" '.*txt$') ]];
    then
    backup_time=$(date +%s)
    backup_file=$backup_time".tar.gz"
    backup_path=${BACKUP}/${backup_file}
    tar cvzf ${backup_path} ${MONITOR}/${file}
    # Send to IPFS 
    previous_hash=$hash
    file_to_upload=${backup_path}
    CID=$(ipfs add -Q -r ${file_to_upload})
    echo "File hash: $hash"
    # Invoke the chaincode
    peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"createBackup","Args":["'"$CORE_PEER_ID"'","'"$hash"'"]}'
    fi
done
