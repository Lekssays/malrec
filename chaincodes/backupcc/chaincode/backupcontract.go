package chaincode

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/common/util"
)

/* Smartcontract provides functions for managing an Backup */
type SmartContract struct {
	/*
	 * It is the Contract struct defined into contractapi package.
	 */
	contractapi.Contract
}

type Backup struct {
	BackupID     string   `json:"backupID"`
	DeviceID     string   `json:"deviceID"`
	Hash         string   `json:"hash"`
	PreviousHash string   `json:"previousHash"`
	Timestamp    int64    `json:"timestamp"`
	IsValid      bool     `json:"isValid"`
	Signature    string   `json:"signature"`
	Paths        []string `json:"paths"`
	Size         int      `json:"size"`
}

type Policy struct {
	PolicyID  string `json:"policyID"`
	Replicas  int    `json:"replicas"`
	Frequency int    `json:"frequency"`
	Offsite   int    `json:"offsite"`
	Size      int    `json:"size"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("InitLedger")
	return nil
}

// CreateBackup adds a new backup to the world state with given details
func (s *SmartContract) CreateBackup(ctx contractapi.TransactionContextInterface, backupID string, deviceID string, hash string, paths string, signature string, size string) (string, error) {
	timestamp, err := strconv.ParseInt(fmt.Sprintf("%d", time.Now().Unix()), 10, 64)
	if err != nil {
		return "", err
	}

	sz, err := strconv.ParseInt(size, 10, 32)
	if err != nil {
		return "", err
	}

	previousHash, _ := s.GetPreviousHash(ctx, deviceID)
	fmt.Printf("Previous Hash: %s\n", previousHash)

	splittedPaths := strings.Split(paths, ";")
	backup := Backup{
		BackupID:     backupID,
		DeviceID:     deviceID,
		Hash:         hash,
		PreviousHash: previousHash,
		Timestamp:    timestamp,
		IsValid:      true,
		Paths:        splittedPaths,
		Signature:    signature,
		Size:         int(sz),
	}

	// check validity of the previous hash
	ts := fmt.Sprintf(`%d`, backup.Timestamp)
	backups, err := s.QueryBackupsByTimestamps(ctx, backup.DeviceID, "000000000", ts)
	sort.SliceStable(backups, func(i, j int) bool {
		return backups[i].Timestamp < backups[j].Timestamp
	})

	if len(backups) > 1 {
		if backups[len(backups)-1].Hash != backup.PreviousHash || !backups[len(backups)-1].IsValid {
			return "", errors.New("backup's previous hash is invalid!")
		}
	}

	// check policy compliance
	policyID := fmt.Sprintf(`%s_policy`, backup.DeviceID)
	chainCodeArgs := util.ToChaincodeArgs("ReadPolicy", policyID)
	response := ctx.GetStub().InvokeChaincode("policy", chainCodeArgs, "mychannel")

	var policy Policy
	var msg string
	json.Unmarshal([]byte(response.Payload), &policy)

	if len(backup.Paths) < int(policy.Replicas) {
		msg = fmt.Sprintf("number of replicas does not satisfy the policy requirements! (want %d got %d)", int(policy.Replicas), len(backup.Paths))
		return "", errors.New(msg)
	}

	if backup.Size > int(policy.Size) {
		msg = fmt.Sprintf("backup size does not satisfy the policy requirements! (want %d got %d)", int(policy.Size), backup.Size)
		return "", errors.New(msg)
	}

	offsite := 0
	onsite := fmt.Sprintf("https://%s", backup.DeviceID)
	for i := 0; i < len(backup.Paths); i++ {
		if !strings.Contains(backup.Paths[i], onsite) {
			offsite += 1
		}
	}

	if offsite < int(policy.Offsite) {
		msg = fmt.Sprintf("backup offsite replicas do not satisfy the policy requirements! (want %d got %d)", int(policy.Offsite), offsite)
		return "", errors.New(msg)
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

func (s *SmartContract) InvalidateBackup(ctx contractapi.TransactionContextInterface, backupID string) (bool, error) {
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

	backup.IsValid = false

	backupJSON, err = json.Marshal(backup)
	if err != nil {
		fmt.Printf("error: %v", err)
		return false, err
	}

	err = ctx.GetStub().PutState(backupID, backupJSON)
	if err != nil {
		fmt.Printf("error: %v", err)
		return false, err
	}

	previousValidHash := backup.PreviousHash
	ts := fmt.Sprintf(`%d`, backup.Timestamp+1)
	backups, err := s.QueryBackupsByTimestamps(ctx, backup.DeviceID, ts, "9999999999")
	sort.SliceStable(backups, func(i, j int) bool {
		return backups[i].Timestamp < backups[j].Timestamp
	})

	backupNextJSON, err := ctx.GetStub().GetState(backups[0].BackupID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	if backupNextJSON == nil {
		return false, fmt.Errorf("the backup %s does not exist", backups[0].BackupID)
	}

	var backupNext Backup
	err = json.Unmarshal(backupNextJSON, &backupNext)
	if err != nil {
		return false, err
	}

	backupNext.PreviousHash = previousValidHash

	backupNextJSON, err = json.Marshal(backupNext)
	if err != nil {
		fmt.Printf("error: %v", err)
		return false, err
	}

	err = ctx.GetStub().PutState(backups[0].BackupID, backupNextJSON)
	if err != nil {
		fmt.Printf("error: %v", err)
		return false, err
	}

	return true, err
}
