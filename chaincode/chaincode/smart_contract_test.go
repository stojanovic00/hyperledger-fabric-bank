package chaincode_test

import (
	"chaincode/chaincode"
	"chaincode/chaincode/mocks"
	"chaincode/model"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// go run github.com/maxbrunsfeld/counterfeiter/v6 -o mocks/transaction.go -fake-name TransactionContext . transactionContext
type transactionContext interface {
	contractapi.TransactionContextInterface
}

// go run github.com/maxbrunsfeld/counterfeiter/v6 -o mocks/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

// go run github.com/maxbrunsfeld/counterfeiter/v6 -o mocks/statequeryiterator.go -fake-name StateQueryIterator . stateQueryIterator
type stateQueryIterator interface {
	shim.StateQueryIteratorInterface
}

//RUN ALL TESTS with go test -v ./chaincode from root dir

func TestInitLedger(t *testing.T) {
	//Arrange
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	contract := chaincode.SmartContract{}

	//Testing happy path
	err := contract.InitLedger(transactionContext)
	require.NoError(t, err)

	//Testing error with insert
	//Mocks return value of "PutState" function to return fmt.Errorf("failed inserting key")
	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = contract.InitLedger(transactionContext)
	require.EqualError(t, err, "failed to put to world state. failed inserting key")
}
func TestCreateBankAccount_BankAccountDoesNotExist(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{} // Correct instantiation

	// Test Case: Bank account doesn't exist, user exists, and bank exists
	chaincodeStub.GetStateReturns(nil, nil)                                                                                                           // Set state to indicate bank account doesn't exist
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"someUserData":"value"}`), nil)                                                                   // Set state to indicate user exists
	chaincodeStub.GetStateReturnsOnCall(2, []byte(`{"ID":"b1","Name":"UniCredit","Headquarters":"Linz, Austria","Since":1969,"PIB":138429230}`), nil) // Set state to indicate bank exists

	err := smartContract.CreateBankAccount(transactionContext, "a1", "EUR", "Visa", "b1", "u1")
	require.NoError(t, err)
}

func TestCreateBankAccount_BankAccountAlreadyExists(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{} // Correct instantiation

	// Test Case: Bank account already exists
	chaincodeStub.GetStateReturns([]byte(`{"ID":"a1","Currency":"EUR","Balance":0.0,"Cards":["Visa"],"Bank":{"ID":"b1","Name":"UniCredit","Headquarters":"Linz, Austria","Since":1969,"PIB":138429230},"UserID":"u1"}`), nil)
	err := smartContract.CreateBankAccount(transactionContext, "a1", "EUR", "Visa", "b1", "u1")
	require.EqualError(t, err, "the bank account with id a1 already exists")
}

func TestCreateBankAccount_UserDoesNotExist(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{} // Correct instantiation

	// Test Case: User doesn't exist, Bank account doesn't exist
	chaincodeStub.GetStateReturnsOnCall(0, nil, nil) // Set state to indicate user doesn't exist
	chaincodeStub.GetStateReturnsOnCall(1, nil, nil) // Set state to indicate bank account doesn't exist

	err := smartContract.CreateBankAccount(transactionContext, "a2", "RSD", "MasterCard", "b1", "u2")
	require.EqualError(t, err, "no registered user with id u2")
}

func TestCreateBankAccount_BankDoesNotExist(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{} // Correct instantiation

	// Test Case: No bank accounts, user exists, and no banks
	chaincodeStub.GetStateReturns(nil, nil)                                         // Set state to indicate no bank accounts exist
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"someUserData":"value"}`), nil) // Set state to indicate user exists
	chaincodeStub.GetStateReturnsOnCall(2, nil, nil)                                // Set state to indicate no banks exist

	err := smartContract.CreateBankAccount(transactionContext, "a3", "RSD", "American Express", "b2", "u3")
	require.EqualError(t, err, "the bank with id b2 does not exist")
}

func TestTransferMoney_EnoughMoney(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{} // Correct instantiation

	// Test Case: Enough money in the source account
	chaincodeStub.GetStateReturnsOnCall(0, []byte(`{"ID":"srcAccount","Currency":0,"Balance":100}`), nil) // Set state to indicate source account exists with EUR currency and balance 100
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"ID":"dstAccount","Currency":0,"Balance":0}`), nil)   // Set state to indicate destination account exists with EUR currency and balance 0
	chaincodeStub.GetStateReturnsOnCall(2, nil, nil)                                                      // Set state to indicate confirmation user exists

	confirmation, err := smartContract.TransferMoney(transactionContext, "srcAccount", "dstAccount", "75.0", "false")
	require.True(t, confirmation)
	require.Nil(t, err)
}

func TestTransferMoney_NotEnoughMoney(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{} // Correct instantiation

	// Test Case: Not enough money in the source account
	chaincodeStub.GetStateReturnsOnCall(0, []byte(`{"ID":"srcAccount","Currency":0,"Balance":50}`), nil) // Set state to indicate source account exists with EUR currency and balance 50
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"ID":"dstAccount","Currency":0,"Balance":0}`), nil)  // Set state to indicate destination account exists with EUR currency and balance 0
	chaincodeStub.GetStateReturnsOnCall(2, nil, nil)                                                     // Set state to indicate confirmation user exists

	_, err := smartContract.TransferMoney(transactionContext, "srcAccount", "dstAccount", "100.0", "false")
	require.EqualError(t, err, "not enough money")
}

func TestTransferMoney_DifferentCurrenciesWithoutConfirmation(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{} // Correct instantiation

	// Test Case: Different currencies without confirmation
	chaincodeStub.GetStateReturnsOnCall(0, []byte(`{"ID":"srcAccount","Currency":0,"Balance":100}`), nil) // Set state to indicate source account exists
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"ID":"dstAccount","Currency":1,"Balance":0}`), nil)   // Set state to indicate destination account exists
	chaincodeStub.GetStateReturnsOnCall(2, []byte(`{"someUserData":"value"}`), nil)                       // Set state to indicate confirmation user exists

	confirmation, err := smartContract.TransferMoney(transactionContext, "srcAccount", "dstAccount", "50.0", "false")
	require.False(t, confirmation)
	require.Nil(t, err)
}

func TestReadBankAccount(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{} // Correct instantiation

	// Test Case 1: Bank account exists
	chaincodeStub.GetStateReturnsOnCall(0, []byte(`{"ID":"existingAccount","Currency":0,"Balance":100}`), nil) // Set state to indicate existing account with EUR currency and balance 100

	_, err := smartContract.ReadBankAccount(transactionContext, "existingAccount")
	require.NoError(t, err)

	// Test Case 2: Bank account does not exist
	chaincodeStub.GetStateReturnsOnCall(1, nil, nil) // Set state to indicate non-existing account

	_, nonExistingErr := smartContract.ReadBankAccount(transactionContext, "nonExistingAccount")
	require.Error(t, nonExistingErr)
	require.EqualError(t, nonExistingErr, "the bank account with id nonExistingAccount does not exist")
}

func TestTransferMoney_SameCurrency(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{} // Correct instantiation

	// Test Case: Same currency with confirmation
	chaincodeStub.GetStateReturnsOnCall(0, []byte(`{"ID":"srcAccount","Currency":0,"Balance":100}`), nil) // Set state for source account with EUR currency and balance 100
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"ID":"dstAccount","Currency":0,"Balance":50}`), nil)  // Set state for destination account with EUR currency and balance 50
	chaincodeStub.GetStateReturnsOnCall(2, nil, nil)                                                      // Set state to indicate confirmation user exists

	confirmation, err := smartContract.TransferMoney(transactionContext, "srcAccount", "dstAccount", "75.0", "true")
	require.True(t, confirmation)
	require.Nil(t, err)
}

func TestTransferMoney_DifferentCurrenciesWithConfirmation(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{} // Correct instantiation

	// Test Case: Different currencies with confirmation
	chaincodeStub.GetStateReturnsOnCall(0, []byte(`{"ID":"srcAccount","Currency":0,"Balance":100}`), nil) // Set state for source account with EUR currency and balance 100
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"ID":"dstAccount","Currency":1,"Balance":50}`), nil)  // Set state for destination account with RSD currency and balance 50
	chaincodeStub.GetStateReturnsOnCall(2, nil, nil)                                                      // Set state to indicate confirmation user exists

	confirmation, err := smartContract.TransferMoney(transactionContext, "srcAccount", "dstAccount", "75.0", "true")
	require.True(t, confirmation)
	require.Nil(t, err)
}

func TestMoneyDepositToAccount(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{}

	accountID := "a99"
	initialBalance := 100.0
	depositAmount := 10.0

	chaincodeStub.GetStateReturns([]byte(fmt.Sprintf(`{"ID":"%s","Currency":0,"Balance":%f}`, accountID, initialBalance+depositAmount)), nil)

	result, err := smartContract.MoneyDepositToAccount(transactionContext, accountID, depositAmount)

	require.True(t, result, "Expected successful deposit")
	require.Nil(t, err, "Unexpected error during deposit")

	expectedUpdatedBalance := initialBalance + depositAmount
	accountJSON, err := chaincodeStub.GetState(accountID)
	require.Nil(t, err, "Error retrieving account state after deposit")

	var updatedAccount model.BankAccount
	err = json.Unmarshal(accountJSON, &updatedAccount)
	require.Nil(t, err, "Error unmarshalling updated account JSON")

	require.Equal(t, expectedUpdatedBalance, updatedAccount.Balance, "Account balance not updated correctly after deposit")
}

func TestAddUser(t *testing.T) {
	//Arrange
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	contract := chaincode.SmartContract{}

	//Happy path
	err := contract.AddUser(transactionContext, "u1", "Aleksandar", "Stojanovic", "aleksandar@gmail.com")
	require.NoError(t, err)

	//Already exists
	chaincodeStub.GetStateReturns([]byte{}, nil)
	err = contract.AddUser(transactionContext, "u1", "Aleksandar", "Stojanovic", "aleksandar@gmail.com")
	require.EqualError(t, err, "the user u1 already exists")
}
