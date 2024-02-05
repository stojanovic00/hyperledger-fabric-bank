package main

import (
	"chaincode/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

func main() {
	bankChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating bank chaincode: %v", err)
	}

	if err := bankChaincode.Start(); err != nil {
		log.Panicf("Error starting bank chaincode: %v", err)
	}
}
