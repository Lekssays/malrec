package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

/* Smartcontract provides functions for managing an Asset */
type SmartContract struct {
	/*
	 * It is the Contract struct defined into contractapi package.
	 */
	contractapi.Contract
}

type Asset struct {
	DeviceID     string `json:"deviceID"`
	CID          string `json:"CID"`
	Previous_CID string `json:"previous_CID"`
	Timestamp    string `json:"timestamp"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("InitLedger")
	return nil
}

func (s *SmartContract) UploadBackup(ctx contractapi.TransactionContextInterface, device_id string, file_path string, previous_path string,
	timestamp string) error {

	// Create the asset: the content of the transaction
	asset := Asset{
		DeviceID:     device_id,
		CID:          file_path,
		Previous_CID: previous_path,
		Timestamp:    timestamp,
	}
	// Add the asset to the ledger
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		fmt.Printf("error: %v", err)
		return err
	}
	ctx.GetStub().PutState(device_id, assetJSON)

	transaction_id := ctx.GetStub().GetTxID()
	fmt.Println("Previous Transaction ID: ", transaction_id)

	// Define the composite key
	composite_key, err := ctx.GetStub().CreateCompositeKey(device_id, []string{device_id})
	if err != nil {
		fmt.Printf("composite key not created: %v", err)
		return err
	}
	value := []byte{0x00}

	return ctx.GetStub().PutState(composite_key, value)
}

func (s *SmartContract) GetBackup(ctx contractapi.TransactionContextInterface, device_id string, transaction_id string) (string, error) {
	// Check whether the device_id exists
	assetJSON, err := ctx.GetStub().GetState(device_id)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state : %v", err)
	}
	if assetJSON == nil {
		return "", fmt.Errorf("the asset %s does not exists", device_id)
	}

	// Get all the transaction concerning the device_id
	transactionHystory, err := ctx.GetStub().GetHistoryForKey(device_id)
	if err != nil {
		return "", fmt.Errorf("failed to read the history of the device %s: %v", device_id, err)
	}

	for transactionHystory.HasNext() {
		// Get info about the first transaction available
		transaction, err := transactionHystory.Next()
		if err != nil {
			return "", fmt.Errorf("failed to read the history of the device %s: %v", device_id, err)
		}
		if transaction.GetTxId() == transaction_id {
			values := transaction.Value
			var asset Asset
			err = json.Unmarshal(values, &asset)
			if err != nil {
				return "", err
			}
			fmt.Println("Path: ", asset.CID, " Previous path: ", asset.Previous_CID, " Timestamp: ", asset.Timestamp)
			return string(values), nil
		}

	}
	fmt.Println("Inexistent transaction_ID: ", transaction_id)
	return "", nil
}

/*
 * AssetExists allows us to check whether an asset exists in the world state database.
 * This function returns 2 elements: a boolean and an error.
 */
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, device_id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(device_id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
