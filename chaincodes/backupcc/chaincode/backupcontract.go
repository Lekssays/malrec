package chaincode

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
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
	Timestamp    int64  `json:"timestamp"`
	IsValid      bool   `json:"isValid"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("InitLedger")
	return nil
}

// CreateBackup adds a new backup to the world state with given details
func (s *SmartContract) CreateBackup(ctx contractapi.TransactionContextInterface, backupID string, deviceID string, hash string) (string, error) {
	timestamp, err := strconv.ParseInt(fmt.Sprintf("%d", time.Now().Unix()), 10, 64)
	if err != nil {
		return "", err
	}

	previousHash, _ := s.GetPreviousHash(ctx, deviceID)
	fmt.Printf("Previous Hash: %s\n", previousHash)

	backup := Backup{
		BackupID:     backupID,
		DeviceID:     deviceID,
		Hash:         hash,
		PreviousHash: previousHash,
		Timestamp:    timestamp,
		IsValid:      true,
	}

	backupAsBytes, err := ctx.GetStub().GetState(backupID)
	if err != nil {
		return "", err
	} else if backupAsBytes != nil {
		fmt.Println("The backup already exists: " + backupID)
		return "", err
	}

	backupJSON, err := json.Marshal(backup)
	if err != nil {
		fmt.Printf("error: %v", err)
		return "", err
	}

	err = ctx.GetStub().PutState(backupID, backupJSON)

	return "", err
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
	startTimestamp, err := strconv.ParseInt(startTime, 10, 64)
	if err != nil {
		return nil, err
	}

	endTimestamp, err := strconv.ParseInt(endTime, 10, 64)
	if err != nil {
		return nil, err
	}

	queryString := fmt.Sprintf(`{"selector":{"deviceID":"%s","timestamp":{"$gte": %v,"$lte": %v},"isValid":{"$eq":true}}}`, deviceID, startTimestamp, endTimestamp)
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

// GetPreviousHash returns the previous hash given a deviceID
func (s *SmartContract) GetPreviousHash(ctx contractapi.TransactionContextInterface, deviceID string) (string, error) {
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

		fmt.Println(backup.Hash)

	}

	if len(backups) == 0 {
		return "null", err
	}

	// Order the backups list by timestamps and then get the latest Hash value
	sort.Slice(backups, func(i, j int) bool { return backups[i].Timestamp < backups[j].Timestamp })

	return backups[len(backups)-1].Hash, nil
}

func (s *SmartContract) DeleteBackup(ctx contractapi.TransactionContextInterface, backupID string) (bool, error) {
	backupJSON, err := ctx.GetStub().GetState(backupID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	if backupJSON == nil {
		return false, fmt.Errorf("the backup %s does not exist", backupID)
	}

	var backup Backup
	err = json.Unmarshal(backupJSON, &backup)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().DelState(backup.BackupID)
	if err != nil {
		return false, fmt.Errorf("Failed to delete state:" + err.Error())
	}

	deviceBackupIndexKey, err := ctx.GetStub().CreateCompositeKey("deviceID~backupID", []string{backup.DeviceID, backup.BackupID})
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().DelState(deviceBackupIndexKey)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	timestampIndexKey, err := ctx.GetStub().CreateCompositeKey("timestamp~backupID", []string{strconv.FormatInt(backup.Timestamp, 10), backup.BackupID})
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().DelState(timestampIndexKey)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	return true, err
}
