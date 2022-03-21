package chaincode

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (s *SmartContract) CreateLP(ctx contractapi.TransactionContextInterface, tokenId1 uint64, tokenId2 uint64) error {
	fmt.Println("TODO:")
	return nil
}

func (s *SmartContract) AddToLP(ctx contractapi.TransactionContextInterface, tokenId1 uint64, tokenId2 uint64, amount1 uint64, amount2 uint64) error {
	fmt.Println("TODO:")
	return nil
}
