package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

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
	BackupID     string `json:"backupID"`
	DeviceID     string `json:"deviceID"`
	Hash         string `json:"hash"`
	PreviousHash string `json:"previousHash"`
	Timestamp    string `json:"timestamp"`
	IsValid      bool   `json:"isValid"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("InitLedger")
	return nil
}

// CreateBackup adds a new backup to the world state with given details
func (s *SmartContract) CreateBackup(ctx contractapi.TransactionContextInterface, deviceID string, hash string) (string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	backupID := fmt.Sprintf("%s_%s", deviceID, timestamp)
	previousHash, _ := getPreviousHash(ctx, deviceID)
	fmt.Printf("Previous Hash: %s", previousHash)

	backup := Backup{
		BackupID:     backupID,
		DeviceID:     deviceID,
		Hash:         hash,
		PreviousHash: previousHash,
		Timestamp:    timestamp,
		IsValid:      true,
	}

	backupAsBytes, err := ctx.GetStub().GetState(hash)
	if err != nil {
		return "", err
	} else if backupAsBytes != nil {
		fmt.Println("The backup already exists: " + hash)
		return "", err
	}

	backupJSON, err := json.Marshal(backup)
	if err != nil {
		fmt.Printf("error: %v", err)
		return "", err
	}

	err = ctx.GetStub().PutState(backupID, backupJSON)
	if err != nil {
		return "", err
	}

	deviceBackupIndexKey, err := ctx.GetStub().CreateCompositeKey("deviceID~backupID", []string{backup.DeviceID, backup.BackupID})
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(deviceBackupIndexKey, []byte{0x00})
	if err != nil {
		return "", err
	}

	timestampIndexKey, err := ctx.GetStub().CreateCompositeKey("timestamp~backupID", []string{backup.Timestamp, backup.BackupID})
	if err != nil {
		return "", err
	}

	return backupID, ctx.GetStub().PutState(timestampIndexKey, []byte{0x00})
}

// QueryBackup returns the backup stored in the world state with given backupID
func (s *SmartContract) QueryBackup(ctx contractapi.TransactionContextInterface, backupID string) (*Backup, error) {
	backupAsBytes, err := ctx.GetStub().GetState(backupID)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if backupAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", backupID)
	}

	backup := new(Backup)
	_ = json.Unmarshal(backupAsBytes, backup)
	return backup, nil
}

// QueryBackupsByDeviceID returns the backups stored in the world state with given deviceID
func (s *SmartContract) QueryBackupsByDeviceID(ctx contractapi.TransactionContextInterface, deviceID string) ([]*Backup, error) {
	queryString := fmt.Sprintf(`{"selector":{"deviceID":"%s"}}`, deviceID)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var backups []*Backup
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var backup Backup
		err = json.Unmarshal(queryResult.Value, &backup)
		if err != nil {
			return nil, err
		}
		backups = append(backups, &backup)
	}

	return backups, nil
}

// QueryBackupsByTimestamps returns the VALID backups stored in the world state with given deviceID, start timestamp, and end timestamp
func (s *SmartContract) QueryBackupsByTimestamps(ctx contractapi.TransactionContextInterface, deviceID string, startTime string, endTime string) ([]*Backup, error) {
	queryString := fmt.Sprintf(`{"selector":{"deviceID":"%s","timestamp":{"$gte": "%s","$lte": "%s"},"isValid":{"$eq":true}}}`, deviceID, startTime, endTime)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var backups []*Backup
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var backup Backup
		err = json.Unmarshal(queryResult.Value, &backup)
		if err != nil {
			return nil, err
		}
		backups = append(backups, &backup)
	}

	return backups, nil
}

// getPreviousHash returns the previous hash given a deviceID
func getPreviousHash(ctx contractapi.TransactionContextInterface, deviceID string) (string, error) {
	queryString := fmt.Sprintf(`{"selector":{"deviceID":"%s"}}`, deviceID)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()

	var backups []*Backup
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}
		var backup Backup
		err = json.Unmarshal(queryResult.Value, &backup)
		if err != nil {
			return "", err
		}
		backups = append(backups, &backup)
	}

	if len(backups) == 0 {
		return "null", err
	}

	return backups[len(backups)-1].PreviousHash, nil
}
