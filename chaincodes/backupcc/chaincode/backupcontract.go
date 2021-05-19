package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

/* Smartcontract provides functions for managing an Backup */
type SmartContract struct {
	/*
	 * It is the Contract struct defined into contractapi package.
	 */
	contractapi.Contract
}

type Backup struct {
	DeviceID     string `json:"deviceID"`
	Hash         string `json:"hash"`
	PreviousHash string `json:"previous_hash"`
	Timestamp    string `json:"timestamp"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("InitLedger")
	return nil
}

func (s *SmartContract) UploadBackup(ctx contractapi.TransactionContextInterface, deviceID string, hash string, previousHash string,
	timestamp string) error {

	// Create the backup: the content of the transaction
	backup := Backup{
		DeviceID:     deviceID,
		Hash:         hash,
		PreviousHash: previousHash,
		Timestamp:    timestamp,
	}

	// Add the backup to the ledger
	backupJSON, err := json.Marshal(backup)
	if err != nil {
		fmt.Printf("error: %v", err)
		return err
	}
	ctx.GetStub().PutState(deviceID, backupJSON)

	transaction_id := ctx.GetStub().GetTxID()
	fmt.Println("Previous Transaction ID: ", transaction_id)

	// Define the composite key
	compositeKey, err := ctx.GetStub().CreateCompositeKey(deviceID, []string{deviceID})
	if err != nil {
		fmt.Printf("composite key not created: %v", err)
		return err
	}
	value := []byte{0x00}

	return ctx.GetStub().PutState(compositeKey, value)
}

func (s *SmartContract) GetBackup(ctx contractapi.TransactionContextInterface, deviceID string, transactionID string) (string, error) {
	// Check whether the deviceID exists
	backupJSON, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state : %v", err)
	}
	if backupJSON == nil {
		return "", fmt.Errorf("the backup %s does not exists", deviceID)
	}

	// Get all the transaction concerning the deviceID
	transactionHistory, err := ctx.GetStub().GetHistoryForKey(deviceID)
	if err != nil {
		return "", fmt.Errorf("failed to read the history of the device %s: %v", deviceID, err)
	}

	for transactionHistory.HasNext() {
		// Get info about the first transaction available
		transaction, err := transactionHistory.Next()
		if err != nil {
			return "", fmt.Errorf("failed to read the history of the device %s: %v", deviceID, err)
		}
		if transaction.GetTxId() == transactionID {
			values := transaction.Value
			var backup Backup
			err = json.Unmarshal(values, &backup)
			if err != nil {
				return "", err
			}
			fmt.Println("Hash: ", backup.Hash, " Previous Hash: ", backup.PreviousHash, " Timestamp: ", backup.Timestamp)
			return string(values), nil
		}

	}
	fmt.Println("Inexistent transaction_ID: ", transactionID)
	return "", nil
}

/*
 * BackupExists allows us to check whether an backup exists in the world state database.
 * This function returns 2 elements: a boolean and an error.
 */
func (s *SmartContract) BackupExists(ctx contractapi.TransactionContextInterface, deviceID string) (bool, error) {
	backupJSON, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return backupJSON != nil, nil
}

// GetAllBackups returns all backups found in world state
func (s *SmartContract) GetAllBackups(ctx contractapi.TransactionContextInterface) ([]*Backup, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all backups in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var backups []*Backup
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var backup Backup
		err = json.Unmarshal(queryResponse.Value, &backup)
		if err != nil {
			return nil, err
		}
		backups = append(backups, &backup)
	}

	return backups, nil
}
