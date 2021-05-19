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

    peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backupcc --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"UploadBackup","Args":["'"$CORE_PEER_ID"'","'"$hash"'","'"$previous_hash"'","'"$initial_backup_time"'"]}'

    export transaction_id=$(tail -n 1 $peer0_log | awk '{split($0,a,"  "); print a[2]}')
    echo $transaction_id
}


function monitorChanges() {
    echo "I'm monitoring $MONITOR directory"
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
        peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n backupcc --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"UploadBackup","Args":["'"$CORE_PEER_ID"'","'"$hash"'","'"$previous_hash"'","'"$backup_time"'"]}' 
        transaction_id=$(tail -n 1 $peer0_log | awk '{split($0,a,"  "); print a[2]}')
        echo $transaction_id
        
        # Sleep some seconds, to allow the complete update of the network.
        sleep 7
        # Execute the getBack.sh script whenever there is an update
        /bin/sh peer0_getBack.sh
        fi
    done
}



#-----------------------------------------------------------------------------------------------


#
# getBack is used to download the bakcup from ipfs and store it inside a folder
#

# CORE_PEER_ID, transaction_id are variables shared with backup.sh script
# download_folder is shared with execution.sh script
echo "The CORE_PEER_ID on getBack is: $CORE_PEER_ID"
echo "The download_folder on getBack is: $download_folder"
echo "The transaction_id on getBack is: $transaction_id"

# Get the path to query IPFS
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n ipfs --peerAddresses $CORE_PEER_ADDRESS --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE -c '{"function":"GetBackup","Args":["'"$CORE_PEER_ID"'","'"$transaction_id"'"]}'
CID=$(tail -n 1 $peer0_log | awk '{split($0,a,"  "); print a[2]}')
echo "CID: $CID"
prev_CID=$(tail -n 1 $peer0_log | awk '{split($0,a,"  "); print a[4]}')
echo "Previous transaction id: $prev_CID"
timestamp=$(tail -n 1 $peer0_log | awk '{split($0,a,"  "); print a[6]}')
echo "Timestamp: $timestamp"

# Query IPFS with the obtained path and put the backup into the correct folder 
ipfs get -o $download_folder/$timestamp".tar.gz" $CID