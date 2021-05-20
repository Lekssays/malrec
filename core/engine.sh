#!/bin/sh
 
export MONITOR="/monitor"
export BACKUP="/backup"


function initBackup() {
    initial_backup_time=$(date +%s)
    initial_backup_file=$initial_backup_time".tar.gz"
    tar cvzf ${BACKUP}/${initial_backup_file} ${MONITOR}
    file_to_upload=${BACKUP}/${initial_backup_file}
    echo "File to upload = ${file_to_upload}"
    hash=$(ipfs add -Q -r ${file_to_upload})
    echo "File hash: $hash"
    peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"createBackup","Args":["'"$CORE_PEER_ID"'","'"$hash"'"]}'
}


function monitorChanges() {
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
}

function getBackup() {
    result=$(peer chaincode query -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"QueryBackup","Args":["'"$1"'"]}')
    hash=$(echo $result | python3 -c "import sys, json; print(json.load(sys.stdin)['hash'])")
    timestamp=$(echo $result | python3 -c "import sys, json; print(json.load(sys.stdin)['timestamp'])")
    ipfs get -o $BACKUP/$timestamp"_downloaded.tar.gz" $hash
}

initBackup
#getBackup peer0.org1.example.com_1621514887