package main

import (
	"log"

	"github.com/Lekssays/malrec/chaincodes/backupcc/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	malwareChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating backup chaincode: %v", err)
	}

	if err := malwareChaincode.Start(); err != nil {
		log.Panicf("Error starting backup chaincode: %v", err)
	}
}
