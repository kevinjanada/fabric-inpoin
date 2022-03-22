package main

import (
	"encoding/json"
	"fmt"
)

type LP struct {
	ID                  uint64
	CreatorID           string
	TokenID             uint64
	TokenSupply         float64
	TokenPlatformSupply float64
	ExchangeRate        float64 // 1 TokenId * ExchangeRate
}

type User struct {
	ID string
}

type Token struct {
	ID        uint64
	CreatorID string
	Name      string
	State     map[string]float64
}

const PLATFORM_FEE float64 = 1000 // 1000 BUMNPoin exchange fee

var tokens map[uint64]*Token = make(map[uint64]*Token)

func balanceOf(tokenId uint64, userId string) float64 {
	return tokens[tokenId].State[userId]
}

var lps map[uint64]LP = make(map[uint64]LP)

var lastLpId uint64 = 0

const TOKEN_PLATFORM_ID uint64 = 1 // BUMN Poin

const LP_PREFIX_ID = "lp"

func createLP(creatorId string, tokenId uint64, tokenSupply float64, tokenPlatformSupply float64, exchangeRate float64) (*LP, error) {
	// Check if user have enough token to supply
	if balanceOf(tokenId, creatorId) < tokenSupply {
		return nil, fmt.Errorf("User does not have enough %v", tokens[tokenId].Name)
	}
	// Check if user have enough token platform to supply
	if balanceOf(TOKEN_PLATFORM_ID, creatorId) < tokenPlatformSupply {
		return nil, fmt.Errorf("User does not have enough %v", tokens[TOKEN_PLATFORM_ID].Name)
	}

	lpId := lastLpId + 1

	lp := LP{
		ID:                  lpId,
		CreatorID:           creatorId,
		TokenID:             tokenId,
		TokenSupply:         tokenSupply,
		TokenPlatformSupply: tokenPlatformSupply,
		ExchangeRate:        exchangeRate,
	}

	// transfer tokenId from creator to lp
	// lpIdKey := LP_PREFIX_ID + strconv.FormatUint(uint64(lpId), 10)
	// transferFrom(tokenId, creatorId, lpIdKey, tokenSupply)
	// transfer TOKEN_PLATFORM_ID from creator to lp
	// transferFrom(TOKEN_PLATFORM_ID, creatorId, lpIdKey, tokenPlatformSupply)
	tokens[tokenId].State[creatorId] -= tokenSupply
	tokens[TOKEN_PLATFORM_ID].State[creatorId] -= tokenPlatformSupply

	lps[lpId] = lp
	lastLpId++

	return &lp, nil
}

func pprint(data interface{}) {
	bytes, _ := json.MarshalIndent(data, "", " ")
	fmt.Println(string(bytes))
}

func exchange(userId string, fromTokenId uint64, toTokenId uint64, amount float64) {
	// TODO: Check if amount covers platform fee

	// TODO: Check if user has enough balance of fromTokenId

	// TODO: Check if LP has enough balance of toTokenId

	// Find the LP
	var lpTokenId uint64
	var exchangeRate float64
	if fromTokenId == 1 || toTokenId == 1 {
		// -------------------
		// Token X -> BUMNPoin
		// -------------------
		if toTokenId == 1 {
			lpTokenId = fromTokenId
			lp := findLPByTokenId(lpTokenId)
			exchangeRate = lp.ExchangeRate

			// Send toTokenId to exchanger
			grossExchangeAmount := amount * exchangeRate
			platformFeeAmount := PLATFORM_FEE * 1
			toTokenAmount := grossExchangeAmount - platformFeeAmount

			// send fromToken amount from user to LP
			tokens[fromTokenId].State[userId] -= amount
			lp.TokenSupply += amount

			// send toToken amount from LP to user
			tokens[toTokenId].State[userId] += toTokenAmount
			lp.TokenPlatformSupply -= toTokenAmount

			// send toToken as fees to platformProvider
			platformProviderID := tokens[TOKEN_PLATFORM_ID].CreatorID
			tokens[TOKEN_PLATFORM_ID].State[platformProviderID] += platformFeeAmount

			lps[lp.ID] = *lp
		}

		// --------------------
		// BUMNPoin -> Token X
		// --------------------
		if fromTokenId == 1 {
			lpTokenId = toTokenId
			lp := findLPByTokenId(lpTokenId)
			exchangeRate = 1 / lp.ExchangeRate

			// Calculate amounts of each token to be given to user and platform provider
			grossExchangeAmount := amount * exchangeRate
			platformFeeAmount := PLATFORM_FEE * exchangeRate
			toTokenAmount := grossExchangeAmount - platformFeeAmount

			// User:BUMN (amt) --> LP
			// Add amount to supply from user
			tokens[fromTokenId].State[userId] -= amount
			lp.TokenPlatformSupply += amount

			// LP --> User:TokenX (amt - fee)
			//	\--> Platform:TokenX (fee)
			// Take out exchangeAmount from supply
			lp.TokenSupply -= grossExchangeAmount
			// send exchangeAmount - platformFeeAmount to user
			tokens[toTokenId].State[userId] += toTokenAmount
			// send platformFeeAmount to provider
			platformProviderID := tokens[TOKEN_PLATFORM_ID].CreatorID
			tokens[TOKEN_PLATFORM_ID].State[platformProviderID] += platformFeeAmount

			lps[lp.ID] = *lp
		}

	}

	// ---------------
	// Route to 2 LPs
	// ---------------
	if fromTokenId != 1 && toTokenId != 1 {
		// -----
		// First Exchange
		// ---------
		finalTokenId := toTokenId
		toTokenId = 1 // Set toTokenId to platformToken for the first exchange
		lpTokenId = fromTokenId
		lp := findLPByTokenId(lpTokenId)
		exchangeRate = lp.ExchangeRate

		// No Fees for first exchange
		grossExchangeAmount := amount * exchangeRate
		platformFeeAmount := PLATFORM_FEE * 0
		toTokenAmount := grossExchangeAmount - platformFeeAmount

		// send fromToken amount from user to LP
		tokens[fromTokenId].State[userId] -= amount
		lp.TokenSupply += amount

		// send toToken amount from LP to user
		tokens[toTokenId].State[userId] += toTokenAmount
		lp.TokenPlatformSupply -= toTokenAmount

		// send toToken as fees to platformProvider
		platformProviderID := tokens[TOKEN_PLATFORM_ID].CreatorID
		tokens[TOKEN_PLATFORM_ID].State[platformProviderID] += platformFeeAmount

		lps[lp.ID] = *lp

		// ----
		// Second Exchange
		// ----------
		// amount now the result of the first exchange
		amount = toTokenAmount

		// Switch tokenIds
		// temp := fromTokenId
		fromTokenId = toTokenId
		toTokenId = finalTokenId

		lpTokenId = toTokenId
		lp = findLPByTokenId(lpTokenId)
		exchangeRate = 1 / lp.ExchangeRate

		// User:BUMN (amt) --> LP
		// Add amount to supply from user
		tokens[fromTokenId].State[userId] -= amount
		lp.TokenPlatformSupply += amount

		// Calculate amounts of each token to be given to user and platform provider
		grossExchangeAmount = amount * exchangeRate
		platformFeeAmount = PLATFORM_FEE * exchangeRate
		toTokenAmount = grossExchangeAmount - platformFeeAmount

		// LP --> User:TokenX (amt - fee)
		//	\--> Platform:TokenX (fee)
		// Take out exchangeAmount from supply
		lp.TokenSupply -= grossExchangeAmount
		// send exchangeAmount - platformFeeAmount to user
		tokens[toTokenId].State[userId] += toTokenAmount
		// send platformFeeAmount to provider
		platformProviderID = tokens[TOKEN_PLATFORM_ID].CreatorID
		tokens[toTokenId].State[platformProviderID] += platformFeeAmount

		lps[lp.ID] = *lp
	}
}

func findLPByTokenId(tokenId uint64) *LP {
	for _, lp := range lps {
		if lp.TokenID == tokenId {
			return &lp
		}
	}
	return nil
}

func main() {
	adminBUMN := &User{ID: "adminBUMN"}
	adminLivin := &User{ID: "adminLivin"}
	adminMiles := &User{ID: "adminMiles"}

	user1 := &User{ID: "user1"}
	user2 := &User{ID: "user2"}

	bumnToken := &Token{
		ID:        1,
		CreatorID: adminBUMN.ID,
		Name:      "BUMNPoin",
		State:     make(map[string]float64),
	}
	bumnToken.State[adminBUMN.ID] = 5000000  // 5 juta
	bumnToken.State[adminLivin.ID] = 2000000 // 2 juta
	bumnToken.State[adminMiles.ID] = 3000000 // 3 juta
	bumnToken.State[user1.ID] = 0
	bumnToken.State[user2.ID] = 0

	livinToken := &Token{
		ID:        2,
		CreatorID: adminLivin.ID,
		Name:      "LivinPoin",
		State:     make(map[string]float64),
	}
	livinToken.State[adminBUMN.ID] = 0
	livinToken.State[adminLivin.ID] = 1000000 // 1 juta
	livinToken.State[adminMiles.ID] = 0       // 3 juta
	livinToken.State[user1.ID] = 10000        // 10 ribu
	livinToken.State[user2.ID] = 10000        // 10 ribu

	milesToken := &Token{
		ID:        3,
		CreatorID: adminMiles.ID,
		Name:      "MilesPoin",
		State:     make(map[string]float64),
	}
	milesToken.State[adminBUMN.ID] = 0
	milesToken.State[adminLivin.ID] = 0
	milesToken.State[adminMiles.ID] = 200000 // 200 ribu
	milesToken.State[user1.ID] = 10000       // 10 ribu
	milesToken.State[user2.ID] = 10000       // 10 ribu

	tokens[1] = bumnToken
	tokens[2] = livinToken
	tokens[3] = milesToken

	// Create liquidity pool
	// 200 ribu LivinPoin
	// 2 juta BUMNPoin
	// Rate 1 Livin = 10 BUMN
	createLP(adminLivin.ID, livinToken.ID, 200000, 2000000, 10)

	// Create liquidity pool
	// 150 ribu MilesPoin
	// 3 juta BUMNPoin
	// Rate 1 Miles = 200 BUMN
	createLP(adminMiles.ID, milesToken.ID, 150000, 3000000, 200)

	fmt.Println("LP Original State")
	pprint(lps[1]) // LP Livin -> BUMN
	pprint(lps[2]) // LP Miles -> BUMN

	exchange(user1.ID, livinToken.ID, bumnToken.ID, 3000)
	pprint(lps[1])
	fmt.Println("LP Livin TokenSupply harus nya nambah 3000 => 203000 == ", lps[1].TokenSupply)
	fmt.Printf("LP Livin TokenPlatformSupply harusnya berkurang (3000 * 10) - 1000 = 29000 => 1971000 == %.0f\n", lps[1].TokenPlatformSupply)

	fmt.Println("user1 Livin balance harusnya berkurang 3000 => 7000 == ", livinToken.State[user1.ID])
	fmt.Println("user1 BUMN balance harusnya bertambah 29000 => 29000 == ", bumnToken.State[user1.ID])

	fmt.Printf("adminBUMN BUMN balance harusnya bertambah 1000 => 5001000 == %.0f\n", bumnToken.State[adminBUMN.ID])

	exchange(user1.ID, bumnToken.ID, livinToken.ID, 3000)
	fmt.Println("sesudah")
	pprint(lps[1])
	fmt.Println("User:BUMN -> LP")
	fmt.Println("user1 BUMN balance harusnya berkurang 3000 => 26000 == ", bumnToken.State[user1.ID])
	fmt.Printf("LP Livin BUMN Supply harusnya bertambah 3000 => 1974000 == %.0f\n", lps[1].TokenPlatformSupply)

	fmt.Println("LP --> USER:Livin")
	fmt.Println("  \\--> adminBUMN:Livin")
	fmt.Println("LP Livin TokenSupply harus nya berkurang 3000 * (1/10) = 300 => 202700 ==", lps[1].TokenSupply)
	fmt.Println("user1 Livin balance harusnya bertambah 200 => 7200 == ", livinToken.State[user1.ID])
	fmt.Printf("adminBUMN Livin balance harusnya bertambah 100 => 5001100 == %.0f\n", bumnToken.State[adminBUMN.ID])

	fmt.Printf("user1 Livin balance %.0f\n", livinToken.State[user1.ID])
	fmt.Printf("LP1 Livin supply  %.0f\n", lps[1].TokenSupply)
	fmt.Printf("LP1 BUMN supply  %0.f\n", lps[1].TokenPlatformSupply)
	fmt.Printf("LP2 Miles supply %0.f\n", lps[2].TokenSupply)
	fmt.Printf("LP2 BUMN supply %0.f\n", lps[2].TokenPlatformSupply)
	fmt.Printf("user1 Miles balance %0.f\n", milesToken.State[user1.ID])
	fmt.Printf("bumnAdmin Miles balance %0.f\n", milesToken.State[adminBUMN.ID])

	exchange(user1.ID, livinToken.ID, milesToken.ID, 3000)
	fmt.Println("User:Livin -> LP -> User:BUMN -> LP -> User:Miles")
	fmt.Println("                                     \\-> adminBumn:Miles")
	fmt.Printf("user1 Livin balance berkurang 3000 -> 7200 - 3000 = 4200 == %.0f\n", livinToken.State[user1.ID])
	fmt.Printf("LP1 Livin supply bertambah 3000 -> 202700 + 3000 = 205700 == %.0f\n", lps[1].TokenSupply)
	fmt.Printf("LP2 BUMN supply berkurang 3000 * 10 -> 1974000 - 30000 = 1944000 == %0.f\n", lps[1].TokenPlatformSupply)
	fmt.Printf("LP2 Miles supply berkurang 30000 * 1/200 -> 150000 - 150 = 149850 == %0.f\n", lps[2].TokenSupply)
	fmt.Printf("LP2 BUMN supply bertambah 30000 -> 3000000 + 30000 = 3030000 == %0.f\n", lps[2].TokenPlatformSupply)
	fmt.Printf("user1 Miles balance bertambah (150 - (1000 * 1/200)) = 10000 + 145 = 10145 == %0.f\n", milesToken.State[user1.ID])
	fmt.Printf("bumnAdmin Miles balance bertambah (1000 * 1/200) = 0 + 5 = 5 == %0.f\n", milesToken.State[adminBUMN.ID])

}
