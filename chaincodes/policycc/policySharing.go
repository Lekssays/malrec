package main

import (
	"log"

	"github.com/Lekssays/malrec/chaincodes/policycc/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	policyChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating policy chaincode: %v", err)
	}

	if err := policyChaincode.Start(); err != nil {
		log.Panicf("Error starting policy chaincode: %v", err)
	}
}
