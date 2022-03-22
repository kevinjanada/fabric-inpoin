package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const lpKeyPrefix = "lp"

const PLATFORM_FEE_KEY = "lp~platformFee"

const PLATFORM_TOKEN_ID_KEY = "lp~platformTokenId"

const lpTokenBalancePrefix = "lp~balance"

type LiquidityPool struct {
	TokenID             uint64  `json:"token_id"`
	TokenSupply         float64 `json:"token_supply"`
	TokenPlatformSupply float64 `json:"token_platform_supply"`
	CreatorID           string  `json:"creator_id"`
	ExchangeRate        float64 `json:"exchange_rate"`
}

type ExchangeResult struct {
	FromTokenID     uint64
	FromTokenAmount float64
	ToTokenID       uint64
	ToTokenAmount   float64
	ExchangeRate    float64
	PlatformFee     float64
}

func (s *SmartContract) CreateLP(ctx contractapi.TransactionContextInterface, tokenId uint64, tokenSupply float64, tokenPlatformSupply float64, exchangeRate float64) (*LiquidityPool, error) {

	// Get ID of submitting client identity
	lpCreatorId, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return nil, fmt.Errorf("failed to get client id: %v", err)
	}

	token1Name, err := s.GetTokenName(ctx, tokenId)
	if err != nil {
		return nil, err
	}
	if token1Name == "" {
		return nil, fmt.Errorf("token with id %v does not exist", tokenId)
	}

	lp := &LiquidityPool{
		CreatorID:           lpCreatorId,
		TokenID:             tokenId,
		TokenSupply:         tokenSupply,
		TokenPlatformSupply: tokenPlatformSupply,
		ExchangeRate:        exchangeRate,
	}

	// LP is identified by tokenId
	lpIdString := strconv.FormatUint(uint64(tokenId), 10)
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

	// Add token balance to LP
	err = s.AddToLP(ctx, lpCreatorId, tokenId, tokenSupply)
	if err != nil {
		return nil, err
	}

	// Add token platform balance to LP
	tokenPlatformId, err := s.GetPlatformTokenID(ctx)
	if err != nil {
		return nil, err
	}
	err = s.AddToLP(ctx, lpCreatorId, tokenPlatformId, tokenPlatformSupply)
	if err != nil {
		return nil, err
	}

	return lp, nil
}

func (s *SmartContract) AddToLP(ctx contractapi.TransactionContextInterface, adderId string, tokenId uint64, amount float64) error {
	tokenIdString := strconv.FormatUint(uint64(tokenId), 10)
	lpTokenBalanceKey, err := ctx.GetStub().CreateCompositeKey(lpTokenBalancePrefix, []string{tokenIdString})
	if err != nil {
		return err
	}
	err = addBalance(ctx, adderId, lpTokenBalanceKey, tokenId, amount)
	if err != nil {
		return err
	}
	err = removeBalance(ctx, adderId, []uint64{tokenId}, []float64{amount})
	if err != nil {
		return err
	}
	return nil
}

func (s *SmartContract) TakeFromLP(ctx contractapi.TransactionContextInterface, takerId string, tokenId uint64, amount float64) error {
	tokenIdString := strconv.FormatUint(uint64(tokenId), 10)
	lpTokenBalanceKey, err := ctx.GetStub().CreateCompositeKey(lpTokenBalancePrefix, []string{tokenIdString})
	err = addBalance(ctx, lpTokenBalanceKey, takerId, tokenId, amount)
	if err != nil {
		return err
	}
	err = removeBalance(ctx, lpTokenBalanceKey, []uint64{tokenId}, []float64{amount})
	if err != nil {
		return err
	}
	return nil
}

func (s *SmartContract) GetLPByTokenID(ctx contractapi.TransactionContextInterface, tokenId uint64) (*LiquidityPool, error) {
	lpIdString := strconv.FormatUint(uint64(tokenId), 10)
	lpKey, err := ctx.GetStub().CreateCompositeKey(lpKeyPrefix, []string{lpIdString})
	if err != nil {
		return nil, fmt.Errorf("failed to create the composite key for prefix %s: %v", lpKeyPrefix, err)
	}
	lpBytes, err := ctx.GetStub().GetState(lpKey)
	if err != nil {
		return nil, err
	}
	var lp LiquidityPool
	err = json.Unmarshal(lpBytes, &lp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode approval JSON of lp tokenId of %v: %v", tokenId, err)
	}
	return &lp, nil
}

func (s *SmartContract) SetPlatformFeeAmount(ctx contractapi.TransactionContextInterface, platformFee float64) (float64, error) {
	err := ctx.GetStub().PutState(PLATFORM_FEE_KEY, []byte(strconv.FormatFloat(platformFee, 'e', 2, 64)))
	if err != nil {
		return 0, err
	}

	return platformFee, nil
}

func (s *SmartContract) GetPlatformFeeAmount(ctx contractapi.TransactionContextInterface) (float64, error) {
	feeBytes, err := ctx.GetStub().GetState(PLATFORM_FEE_KEY)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(string(feeBytes), 64)
}

func (s *SmartContract) SetPlatformTokenID(ctx contractapi.TransactionContextInterface, tokenId uint64) (uint64, error) {
	err := ctx.GetStub().PutState(PLATFORM_TOKEN_ID_KEY, []byte(strconv.FormatUint(tokenId, 10)))
	if err != nil {
		return 0, err
	}

	return tokenId, nil
}

func (s *SmartContract) GetPlatformTokenID(ctx contractapi.TransactionContextInterface) (uint64, error) {
	tokenIdBytes, err := ctx.GetStub().GetState(PLATFORM_TOKEN_ID_KEY)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(string(tokenIdBytes), 10, 64)
}

func (s *SmartContract) SaveLPState(ctx contractapi.TransactionContextInterface, lp *LiquidityPool) error {
	lpIdString := strconv.FormatUint(uint64(lp.TokenID), 10)
	lpKey, err := ctx.GetStub().CreateCompositeKey(lpKeyPrefix, []string{lpIdString})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", lpKeyPrefix, err)
	}
	lpJson, err := json.Marshal(lp)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(lpKey, lpJson)
	if err != nil {
		return err
	}
	return nil
}

func (s *SmartContract) Exchange(
	ctx contractapi.TransactionContextInterface,
	fromTokenId uint64,
	toTokenId uint64,
	amount float64,
) (result *ExchangeResult, err error) {

	// Get ID of submitting client identity
	exchangerId, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return nil, err
	}

	platformTokenId, err := s.GetPlatformTokenID(ctx)
	if err != nil {
		return nil, err
	}

	PLATFORM_FEE, err := s.GetPlatformFeeAmount(ctx)
	if err != nil {
		return nil, err
	}

	exchangeResult := &ExchangeResult{
		FromTokenID:     fromTokenId,
		FromTokenAmount: amount,
		ToTokenID:       toTokenId,
		ToTokenAmount:   0,
		ExchangeRate:    0,
		PlatformFee:     0,
	}

	if fromTokenId == platformTokenId || toTokenId == platformTokenId {
		if toTokenId == platformTokenId {
			// -------------------
			// Token X -> BUMNPoin
			// -------------------
			lp, err := s.GetLPByTokenID(ctx, fromTokenId)
			if err != nil {
				return nil, err
			}

			// Calculate amounts of each token to be given to user and platform provider
			exchangeRate := lp.ExchangeRate
			grossExchangeAmount := amount * exchangeRate
			platformFeeAmount := PLATFORM_FEE * 1
			toTokenAmount := grossExchangeAmount - platformFeeAmount

			exchangeResult.ToTokenAmount = toTokenAmount
			exchangeResult.ExchangeRate = exchangeRate
			exchangeResult.PlatformFee = platformFeeAmount

			// Check if amount covers platformFeeAmount
			if toTokenAmount < 0 {
				return nil, fmt.Errorf(
					"amount %v of tokenId %v to exchange does not cover platform fee %v of tokenId %v",
					amount, fromTokenId, platformFeeAmount, toTokenId,
				)
			}

			// Send fromToken amount to LP
			err = s.AddToLP(ctx, exchangerId, fromTokenId, amount)
			if err != nil {
				return nil, err
			}

			// send toToken amount from LP to user
			err = s.TakeFromLP(ctx, exchangerId, toTokenId, toTokenAmount)
			if err != nil {
				return nil, err
			}

			// send toToken as fees to platformProvider
			platformTokenCreatorId, err := s.GetTokenCreator(ctx, platformTokenId)
			if err != nil {
				return nil, err
			}
			err = s.TakeFromLP(ctx, platformTokenCreatorId, toTokenId, platformFeeAmount)
			if err != nil {
				return nil, err
			}

			lp.TokenSupply += amount
			lp.TokenPlatformSupply -= grossExchangeAmount
			err = s.SaveLPState(ctx, lp)
			if err != nil {
				return nil, err
			}
		}

		if fromTokenId == platformTokenId {
			lp, err := s.GetLPByTokenID(ctx, toTokenId)
			if err != nil {
				return nil, err
			}
			exchangeRate := 1 / lp.ExchangeRate

			// Calculate amounts of each token to be given to user and platform provider
			grossExchangeAmount := amount * exchangeRate
			platformFeeAmount := PLATFORM_FEE * exchangeRate
			toTokenAmount := grossExchangeAmount - platformFeeAmount

			exchangeResult.ToTokenAmount = toTokenAmount
			exchangeResult.ExchangeRate = exchangeRate
			exchangeResult.PlatformFee = platformFeeAmount

			// Check if amount covers platformFeeAmount
			if toTokenAmount < 0 {
				return nil, fmt.Errorf(
					"amount %v of tokenId %v to exchange does not cover platform fee %v of tokenId %v",
					amount, fromTokenId, platformFeeAmount, toTokenId,
				)
			}

			// User:BUMN (amt) --> LP
			// Add amount to supply from user
			err = s.AddToLP(ctx, exchangerId, fromTokenId, amount)
			if err != nil {
				return nil, err
			}

			// LP --> User:TokenX (amt - fee)
			//	\--> Platform:TokenX (fee)
			// Take out exchangeAmount from supply and send (exchangeAmount - platformFeeAmount) to user
			err = s.TakeFromLP(ctx, exchangerId, toTokenId, toTokenAmount)
			if err != nil {
				return nil, err
			}
			// send platformFeeAmount to provider
			platformTokenCreatorId, err := s.GetTokenCreator(ctx, platformTokenId)
			if err != nil {
				return nil, err
			}
			err = s.TakeFromLP(ctx, platformTokenCreatorId, toTokenId, platformFeeAmount)
			if err != nil {
				return nil, err
			}

			lp.TokenPlatformSupply += amount
			lp.TokenSupply -= grossExchangeAmount
			err = s.SaveLPState(ctx, lp)
			if err != nil {
				return nil, err
			}
		}
	}

	if fromTokenId != platformTokenId && toTokenId != platformTokenId {
		// -----
		// First Exchange
		// ---------
		finalTokenId := toTokenId
		toTokenId = platformTokenId // Set toTokenId to platformToken for the first exchange
		lp, err := s.GetLPByTokenID(ctx, fromTokenId)
		if err != nil {
			return nil, err
		}
		exchangeRate := lp.ExchangeRate

		// No Fees for first exchange
		grossExchangeAmount := amount * exchangeRate
		platformFeeAmount := PLATFORM_FEE * 0
		toTokenAmount := grossExchangeAmount - platformFeeAmount

		// Check if amount covers platformFeeAmount
		finalLp, err := s.GetLPByTokenID(ctx, finalTokenId)
		if err != nil {
			return nil, err
		}
		finalExchangeRate := 1 / finalLp.ExchangeRate
		finalGrossExchangeAmount := grossExchangeAmount * finalExchangeRate
		finalPlatformFeeAmount := PLATFORM_FEE * finalExchangeRate
		finalToTokenAmount := finalGrossExchangeAmount - finalPlatformFeeAmount
		if finalToTokenAmount < 0 {
			return nil, fmt.Errorf(
				"amount %v of tokenId %v to exchange does not cover platform fee %v of tokenId %v",
				amount, fromTokenId, finalPlatformFeeAmount, finalTokenId,
			)
		}

		// send fromToken amount from user to LP
		err = s.AddToLP(ctx, exchangerId, fromTokenId, amount)
		if err != nil {
			return nil, err
		}

		// send toToken amount from LP to user
		err = s.TakeFromLP(ctx, exchangerId, toTokenId, toTokenAmount)
		if err != nil {
			return nil, err
		}

		// send toToken as fees to platformProvider
		platformTokenCreatorId, err := s.GetTokenCreator(ctx, platformTokenId)
		if err != nil {
			return nil, err
		}
		err = s.TakeFromLP(ctx, platformTokenCreatorId, toTokenId, platformFeeAmount)
		if err != nil {
			return nil, err
		}

		lp.TokenSupply += amount
		lp.TokenPlatformSupply -= grossExchangeAmount
		err = s.SaveLPState(ctx, lp)
		if err != nil {
			return nil, err
		}

		// ----
		// Second Exchange
		// ----------
		// amount is now the result of the first exchange
		amount = toTokenAmount

		// Switch tokenIds
		fromTokenId = toTokenId
		toTokenId = finalTokenId

		lpTokenId := toTokenId
		lp, err = s.GetLPByTokenID(ctx, lpTokenId)
		if err != nil {
			return nil, err
		}
		exchangeRate = 1 / lp.ExchangeRate

		// User:BUMN (amt) --> LP
		// Add amount to supply from user
		err = s.AddToLP(ctx, exchangerId, fromTokenId, amount)
		if err != nil {
			return nil, err
		}

		// Calculate amounts of each token to be given to user and platform provider
		grossExchangeAmount = amount * exchangeRate
		platformFeeAmount = PLATFORM_FEE * exchangeRate
		toTokenAmount = grossExchangeAmount - platformFeeAmount

		exchangeResult.ToTokenAmount = toTokenAmount
		exchangeResult.ExchangeRate = exchangeRate
		exchangeResult.PlatformFee = platformFeeAmount

		// LP --> User:TokenX (amt - fee)
		//	\--> Platform:TokenX (fee)
		// Take out exchangeAmount from supply
		// send (exchangeAmount - platformFeeAmount) to user
		err = s.TakeFromLP(ctx, exchangerId, toTokenId, toTokenAmount)
		if err != nil {
			return nil, err
		}
		// send platformFeeAmount to provider
		platformTokenCreatorId, err = s.GetTokenCreator(ctx, platformTokenId)
		if err != nil {
			return nil, err
		}
		err = s.TakeFromLP(ctx, platformTokenCreatorId, toTokenId, platformFeeAmount)
		if err != nil {
			return nil, err
		}

		lp.TokenPlatformSupply += amount
		lp.TokenSupply -= grossExchangeAmount
		err = s.SaveLPState(ctx, lp)
		if err != nil {
			return nil, err
		}
	}

	return exchangeResult, nil
}
