package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const lpKeyPrefix = "lp"

// Keep track of lps owned by owner
// const lpsByOwnerPrefix = "lp~owner"

var lpIdCount uint64 = 0

type LiquidityPool struct {
	ID           uint64  `json:"id"`
	CreatorId    string  `json:"creator_id"`
	Token1Id     uint64  `json:"token_a"`
	Token2Id     uint64  `json:"token_b"`
	Token1Supply float64 `json:"token_a_supply"`
	Token2Supply float64 `json:"token_b_supply"`
}

func (s *SmartContract) CreateLP(ctx contractapi.TransactionContextInterface, lpId uint64, token1Id uint64, token2Id uint64) (*LiquidityPool, error) {

	// Get ID of submitting client identity
	lpCreatorId, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return nil, fmt.Errorf("failed to get client id: %v", err)
	}

	token1Name, err := s.GetTokenName(ctx, token1Id)
	if err != nil {
		return nil, err
	}
	if token1Name == "" {
		return nil, fmt.Errorf("token with id %v does not exist", token1Id)
	}

	token2Name, err := s.GetTokenName(ctx, token1Id)
	if err != nil {
		return nil, err
	}
	if token2Name == "" {
		return nil, fmt.Errorf("token with id %v does not exist", token2Id)
	}

	lpId := lpIdCount + 1
	lp := &LiquidityPool{
		ID:           lpId,
		CreatorId:    lpCreatorId,
		Token1Id:     token1Id,
		Token2Id:     token2Id,
		Token1Supply: 0,
		Token2Supply: 0,
	}

	lpIdString := strconv.FormatUint(uint64(lpId), 10)
	lpKey, err := ctx.GetStub().CreateCompositeKey(lpKeyPrefix, []string{lpIdString})
	if err != nil {
		return nil, fmt.Errorf("failed to create the composite key for prefix %s: %v", lpKeyPrefix, err)
	}

	lpJson, err := json.Marshal(lp)
	if err != nil {
		return nil, err
	}

	err = ctx.GetStub().PutState(lpKey, lpJson)
	if err != nil {
		return nil, err
	}

	// Increment tokenIdCount
	lpIdCount++

	return lp, nil
}

func (s *SmartContract) AddToLP(ctx contractapi.TransactionContextInterface, tokenId1 uint64, tokenId2 uint64, amount1 uint64, amount2 uint64) error {
	fmt.Println("TODO:")
	return nil
}
