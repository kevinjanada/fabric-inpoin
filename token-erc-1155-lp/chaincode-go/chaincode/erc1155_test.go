package chaincode_test

import (
	"os"
	"testing"

	"erc1155/chaincode"
	"erc1155/chaincode/mocks"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/require"
)

type transactionContextInterface interface {
	contractapi.TransactionContextInterface
}

const assetCollectionName = "assetCollection"
const transferAgreementObjectType = "transferAgreement"
const myOrg1Msp = "Org1Testmsp"
const myOrg1Clientid = "myOrg1Userid"
const myOrg1PrivCollection = "Org1TestmspPrivateCollection"
const myOrg2Msp = "Org2Testmsp"
const myOrg2Clientid = "myOrg2Userid"
const myOrg2PrivCollection = "Org2TestmspPrivateCollection"

const minterMSPID = "Org1MSP"
const minterClientId = "OrgClientId"

func prepMocksAsOrg1() (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	return prepMocks(myOrg1Msp, myOrg1Clientid)
}
func prepMocksAsOrg2() (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	return prepMocks(myOrg2Msp, myOrg2Clientid)
}
func prepMocksAsMinterMSP() (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	return prepMocks(minterMSPID, minterClientId)
}
func prepMocks(orgMSP, clientId string) (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	clientIdentity := &mocks.ClientIdentity{}
	clientIdentity.GetMSPIDReturns(orgMSP, nil)
	clientIdentity.GetIDReturns(clientId, nil)
	//set matching msp ID using peer shim env variable
	os.Setenv("CORE_PEER_LOCALMSPID", orgMSP)
	transactionContext.GetClientIdentityReturns(clientIdentity)
	return transactionContext, chaincodeStub
}

func TestCreateToken(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsMinterMSP()
	chaincode := chaincode.SmartContract{}

	transactionContext.GetStubReturns(chaincodeStub)

	expectedCreator := minterClientId

	err := chaincode.CreateToken(transactionContext, 1, "token1Name")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns([]byte(expectedCreator), nil)
	creator, err := chaincode.GetTokenCreator(transactionContext, 1)
	require.NoError(t, err)
	require.Equal(t, creator, expectedCreator)

	expectedTokenName := "token1Name"
	chaincodeStub.GetStateReturns([]byte(expectedTokenName), nil)
	tokenName, err := chaincode.GetTokenName(transactionContext, 1)
	require.NoError(t, err)
	require.Equal(t, tokenName, expectedTokenName)
}
