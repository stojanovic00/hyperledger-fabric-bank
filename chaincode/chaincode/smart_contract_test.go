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
	chaincodeStub.GetStateReturns(nil, nil)
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"someUserData":"value"}`), nil)
	chaincodeStub.GetStateReturnsOnCall(2, nil, nil)

	err := smartContract.CreateBankAccount(transactionContext, "a3", "RSD", "American Express", "b2", "u3")
	require.EqualError(t, err, "the bank with id b2 does not exist")
}

func TestTransferMoney_EnoughMoney(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{}

	// Test Case: Enough money in the source account
	chaincodeStub.GetStateReturnsOnCall(0, []byte(`{"ID":"srcAccount","Currency":0,"Balance":100}`), nil)
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"ID":"dstAccount","Currency":0,"Balance":0}`), nil)
	chaincodeStub.GetStateReturnsOnCall(2, nil, nil)
	confirmation, err := smartContract.TransferMoney(transactionContext, "srcAccount", "dstAccount", "75.0", "false")
	require.True(t, confirmation)
	require.Nil(t, err)
}

func TestTransferMoney_NotEnoughMoney(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{}

	// Test Case: Not enough money in the source account
	chaincodeStub.GetStateReturnsOnCall(0, []byte(`{"ID":"srcAccount","Currency":0,"Balance":50}`), nil)
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"ID":"dstAccount","Currency":0,"Balance":0}`), nil)
	chaincodeStub.GetStateReturnsOnCall(2, nil, nil)

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
	chaincodeStub.GetStateReturnsOnCall(0, []byte(`{"ID":"srcAccount","Currency":0,"Balance":100}`), nil)
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"ID":"dstAccount","Currency":1,"Balance":0}`), nil)
	chaincodeStub.GetStateReturnsOnCall(2, []byte(`{"someUserData":"value"}`), nil)

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
	chaincodeStub.GetStateReturnsOnCall(1, nil, nil)

	_, nonExistingErr := smartContract.ReadBankAccount(transactionContext, "nonExistingAccount")
	require.Error(t, nonExistingErr)
	require.EqualError(t, nonExistingErr, "the bank account with id nonExistingAccount does not exist")
}

func TestTransferMoney_SameCurrency(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{}

	// Test Case: Same currency with confirmation
	chaincodeStub.GetStateReturnsOnCall(0, []byte(`{"ID":"srcAccount","Currency":0,"Balance":100}`), nil)
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"ID":"dstAccount","Currency":0,"Balance":50}`), nil)
	chaincodeStub.GetStateReturnsOnCall(2, nil, nil)

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
	chaincodeStub.GetStateReturnsOnCall(0, []byte(`{"ID":"srcAccount","Currency":0,"Balance":100}`), nil)
	chaincodeStub.GetStateReturnsOnCall(1, []byte(`{"ID":"dstAccount","Currency":1,"Balance":50}`), nil)
	chaincodeStub.GetStateReturnsOnCall(2, nil, nil)

	confirmation, err := smartContract.TransferMoney(transactionContext, "srcAccount", "dstAccount", "75.0", "true")
	require.True(t, confirmation)
	require.Nil(t, err)
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

func TestMoneyWithdrawal(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{}

	// Test Case: Successful withdrawal
	account := model.BankAccount{
		ID:      "bankAccountID",
		UserID:  "usrID",
		Balance: 100,
	}
	accountJSON, _ := json.Marshal(account)
	chaincodeStub.GetStateReturns(accountJSON, nil)

	chaincodeStub.PutStateStub = func(key string, value []byte) error {
		require.Equal(t, "bankAccountID", key)
		var updatedAccount model.BankAccount
		json.Unmarshal(value, &updatedAccount)
		require.Equal(t, float64(50.0), updatedAccount.Balance)
		return nil
	}

	confirmation, err := smartContract.MoneyWithdrawal(transactionContext, "usrID", "bankAccountID", 50.0)
	require.True(t, confirmation)
	require.NoError(t, err)
}

func TestMoneyWithdrawal_InsufficientFunds(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{}

	// Test Case: Insufficient funds
	account := model.BankAccount{
		ID:      "bankAccountID",
		UserID:  "usrID",
		Balance: 40,
	}
	accountJSON, _ := json.Marshal(account)
	chaincodeStub.GetStateReturns(accountJSON, nil)

	confirmation, err := smartContract.MoneyWithdrawal(transactionContext, "usrID", "bankAccountID", 50.0)
	require.False(t, confirmation)
	require.EqualError(t, err, "Insufficient funds")
}

func TestMoneyWithdrawal_AccountNotFound(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{}

	// Test Case: Account not found
	chaincodeStub.GetStateReturns([]byte(`{"ID":"bankAccountID","UserID":"usrID","Balance":100}`), nil)

	confirmation, err := smartContract.MoneyWithdrawal(transactionContext, "usrID", "bankAccountID", 50.0)
	require.False(t, confirmation)
	require.EqualError(t, err, "bank account with ID bankAccountID not found for user usrID")
}

func TestMoneyDepositToAccount(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{}

	// Test Case: Successful deposit
	account := model.BankAccount{
		ID:      "bankAccountID",
		UserID:  "usrID",
		Balance: 100,
	}
	accountJSON, _ := json.Marshal(account)
	chaincodeStub.GetStateReturns(accountJSON, nil)

	chaincodeStub.PutStateStub = func(key string, value []byte) error {
		require.Equal(t, "bankAccountID", key)
		var updatedAccount model.BankAccount
		json.Unmarshal(value, &updatedAccount)
		require.Equal(t, float64(150.0), updatedAccount.Balance)
		return nil
	}

	confirmation, err := smartContract.MoneyDepositToAccount(transactionContext, "usrID", "bankAccountID", 50.0)
	require.True(t, confirmation)
	require.NoError(t, err)
}

func TestMoneyDepositToAccount_AccountNotFound(t *testing.T) {
	// Setup
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	smartContract := chaincode.SmartContract{}

	// Test Case: Account not found
	chaincodeStub.GetStateReturns([]byte(`{"ID":"bankAccountID","UserID":"usrID","Balance":100}`), nil)

	confirmation, err := smartContract.MoneyDepositToAccount(transactionContext, "usrID", "bankAccountID", 50.0)
	require.False(t, confirmation)
	require.EqualError(t, err, "bank account with ID bankAccountID not found for user usrID")
}
