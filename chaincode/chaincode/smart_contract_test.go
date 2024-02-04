package chaincode_test

import (
	"chaincode/chaincode"
	"chaincode/chaincode/mocks"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"

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
