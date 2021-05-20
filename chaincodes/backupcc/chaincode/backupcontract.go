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
func (s *SmartContract) CreateBackup(ctx contractapi.TransactionContextInterface, deviceID string, hash string, previousHash string,
	timestamp string) (string, error) {
	backupID := fmt.Sprintf("%s_%s", deviceID, timestamp)

	fmt.Println("backupID = %s", backupID)

	//todo(ahmed): write a function to get the previous hash
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

	indexName := "deviceID~backupID"
	deviceBackupIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{backup.DeviceID, backup.BackupID})
	if err != nil {
		return "", err
	}
	value := []byte{0x00}

	return backupID, ctx.GetStub().PutState(deviceBackupIndexKey, value)
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

// // queryBackupsByDeviceId returns the backups stored in the world state of a specific deviceID
// func (t *SmartContract) QueryBackupsByDeviceId(ctx contractapi.TransactionContextInterface, deviceID string) pb.Response {
// 	queryString := fmt.Sprintf("{\"selector\":{\"deviceID\":\"%s\"}}", deviceID)

// 	queryResults, err := getQueryResultForQueryString(ctx, queryString)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	return shim.Success(queryResults)
// }

// // getQueryResultForQueryString executes the passed in query string.
// // Result set is built and returned as a byte array containing the JSON results.
// func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]byte, error) {

// 	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

// 	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resultsIterator.Close()

// 	buffer, err := constructQueryResponseFromIterator(resultsIterator)
// 	if err != nil {
// 		return nil, err
// 	}

// 	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

// 	return buffer.Bytes(), nil
// }

// // constructQueryResponseFromIterator constructs a JSON array containing query results from
// // a given result iterator
// func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
// 	// buffer is a JSON array containing QueryResults
// 	var buffer bytes.Buffer
// 	buffer.WriteString("[")

// 	bArrayMemberAlreadyWritten := false
// 	for resultsIterator.HasNext() {
// 		queryResponse, err := resultsIterator.Next()
// 		if err != nil {
// 			return nil, err
// 		}
// 		// Add a comma before array members, suppress it for the first array member
// 		if bArrayMemberAlreadyWritten {
// 			buffer.WriteString(",")
// 		}
// 		buffer.WriteString("{\"Key\":")
// 		buffer.WriteString("\"")
// 		buffer.WriteString(queryResponse.Key)
// 		buffer.WriteString("\"")

// 		buffer.WriteString(", \"Record\":")
// 		// Record is a JSON object, so we write as-is
// 		buffer.WriteString(string(queryResponse.Value))
// 		buffer.WriteString("}")
// 		bArrayMemberAlreadyWritten = true
// 	}
// 	buffer.WriteString("]")

// 	return &buffer, nil
// }

// QueryAllBackups returns all backups found in world state
func (s *SmartContract) QueryAllBackupsByKeys(ctx contractapi.TransactionContextInterface, startKey string, endKey string) ([]*Backup, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
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

// // QueryBackupsByDeviceID returns all backups found in world state for a specific device
// func (s *SmartContract) QueryBackupsByDeviceID(ctx contractapi.TransactionContextInterface, deviceID string) ([]*Backup, error) {
// 	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("deviceID~backupID", []string{deviceID})
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resultsIterator.Close()

// 	var backups []*Backup
// 	for resultsIterator.HasNext() {
// 		queryResponse, err := resultsIterator.Next()
// 		if err != nil {
// 			return nil, err
// 		}

// 		objectType, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
// 		if err != nil {
// 			return nil, err
// 		}

// 		fmt.Printf("- found a backup from index:%s deviceID:%s backupID:%s\n", objectType, compositeKeyParts[0], compositeKeyParts[1])

// 		var backup Backup
// 		err = json.Unmarshal(queryResponse.Value, &backup)
// 		if err != nil {
// 			return nil, err
// 		}
// 		backups = append(backups, &backup)
// 	}

// 	return backups, nil
// }

func (t *SmartContract) QueryBackupsByDeviceID(ctx contractapi.TransactionContextInterface, deviceID string) ([]*Backup, error) {
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
