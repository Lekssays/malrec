package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract defines the structure of a smart contract
type SmartContract struct {
	contractapi.Contract
}

type Policy struct {
	PolicyID  string `json:"policyID"`
	Replicas  int    `json:"replicas"`
	Frequency int    `json:"frequency"`
	Offsite   int    `json:"offsite"`
	Size      int    `json:"size"`
}

// InitLedger adds a base set of policy entries to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("Init Ledger")
	return nil
}

// CreatePolicy issues a new policy entry to the world state with given details.
func (s *SmartContract) CreatePolicy(ctx contractapi.TransactionContextInterface, policyID string, replicas string, frequency string, offsite string, size string) error {
	exists, err := s.PolicyExists(ctx, policyID)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("the policy %s already exists", policyID)
	}

	f, err := strconv.Atoi(frequency)
	if err != nil {
		return err
	}

	r, err := strconv.Atoi(replicas)
	if err != nil {
		return err
	}

	o, err := strconv.Atoi(offsite)
	if err != nil {
		return err
	}

	sz, err := strconv.Atoi(size)
	if err != nil {
		return err
	}

	policy := Policy{
		PolicyID:  policyID,
		Replicas:  r,
		Frequency: f,
		Offsite:   o,
		Size:      sz,
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(policyID, policyJSON)
}

// ReadPolicy returns the policy entry stored in the world state with given id.
func (s *SmartContract) ReadPolicy(ctx contractapi.TransactionContextInterface, policyID string) (*Policy, error) {
	policyJSON, err := ctx.GetStub().GetState(policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if policyJSON == nil {
		return nil, fmt.Errorf("the policy %s does not exist", policyID)
	}

	var policy Policy
	err = json.Unmarshal(policyJSON, &policy)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

// PolicyExists returns true when a policy entry with given ID exists in world state
func (s *SmartContract) PolicyExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	policyJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return policyJSON != nil, nil
}
