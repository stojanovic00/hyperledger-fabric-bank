package chaincode_test

import (
	"chaincode/chaincode"
	"chaincode/chaincode/mocks"
	"chaincode/model"
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

	err := smartContract.CreateBankAccount(transactionContext, "a1", model.EUR, []string{"Visa"}, "b1", "u1")
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
	err := smartContract.CreateBankAccount(transactionContext, "a1", model.EUR, []string{"Visa"}, "b1", "u1")
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

	err := smartContract.CreateBankAccount(transactionContext, "a2", model.RSD, []string{"MasterCard"}, "b1", "u2")
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

	err := smartContract.CreateBankAccount(transactionContext, "a3", model.RSD, []string{"American Express"}, "b2", "u3")
	require.EqualError(t, err, "the bank with id b2 does not exist")
}
