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
    previous_hash="null"

    output=$(peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"createBackup","Args":["'"$CORE_PEER_ID"'","'"$hash"'","'"$previous_hash"'","'"$initial_backup_time"'"]}')

    echo "--------------"
    echo "output = $output"
    #export transaction_id=$(tail -n 1 $peer0_log | awk '{split($0,a,"  "); print a[2]}')
    #echo $transaction_id
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
        peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backup --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"UploadBackup","Args":["'"$CORE_PEER_ID"'","'"$hash"'","'"$previous_hash"'","'"$backup_time"'"]}' 
        transaction_id=$(tail -n 1 $peer0_log | awk '{split($0,a,"  "); print a[2]}')
        echo $transaction_id
        fi
    done
}

function getBackup() {
    echo "The CORE_PEER_ID on getBack is: $CORE_PEER_ID"
    echo "The download_folder on getBack is: $download_folder"
    echo "The transaction_id on getBack is: $transaction_id"

    # Get the path to query IPFS
    peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backupcc --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"GetBackup","Args":["'"$CORE_PEER_ID"'","'"$transaction_id"'"]}'
    CID=$(tail -n 1 $peer0_log | awk '{split($0,a,"  "); print a[2]}')
    echo "File Hash: $hash"
    prev_CID=$(tail -n 1 $peer0_log | awk '{split($0,a,"  "); print a[4]}')
    echo "Previous transaction id: $previous_hash"
    timestamp=$(tail -n 1 $peer0_log | awk '{split($0,a,"  "); print a[6]}')
    echo "Timestamp: $timestamp"

    # Query IPFS with the obtained path and put the backup into the correct folder 
    ipfs get -o $download_folder/$timestamp".tar.gz" $hash
}

initBackup