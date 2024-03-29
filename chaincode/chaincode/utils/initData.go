package utils

import (
	"chaincode/model"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func InitializeData() ([]model.Bank, []model.User, []model.BankAccount) {
	banks := []model.Bank{
		{ID: "b1", Name: "UniCredit", Headquarters: "Linz, Austria", Since: 1969, PIB: 138429230},
		{ID: "b2", Name: "Raiffeisen Bank", Headquarters: "Vienna, Austria", Since: 1927, PIB: 537891234},
		{ID: "b3", Name: "Erste Group", Headquarters: "Vienna, Austria", Since: 1819, PIB: 987654321},
		{ID: "b4", Name: "OTP Bank", Headquarters: "Budapest, Hungary", Since: 1949, PIB: 654321789},
	}

	users := []model.User{
		{ID: "u1", Name: "John", Surname: "Doe", Email: "john.doe@gmail.com"},
		{ID: "u2", Name: "Alice", Surname: "Smith", Email: "alice.smith@gmail.com"},
		{ID: "u3", Name: "Bob", Surname: "Johnson", Email: "bob.johnson@gmail.com"},
		{ID: "u4", Name: "Eva", Surname: "Williams", Email: "eva.williams@gmail.com"},
		{ID: "u5", Name: "Daniel", Surname: "Miller", Email: "daniel.miller@gmail.com"},
		{ID: "u6", Name: "Sophia", Surname: "Brown", Email: "sophia.brown@gmail.com"},
		{ID: "u7", Name: "John", Surname: "Davis", Email: "matthew.davis@gmail.com"},
		{ID: "u8", Name: "Olivia", Surname: "Jones", Email: "olivia.jones@gmail.com"},
		{ID: "u9", Name: "Michael", Surname: "Smith", Email: "michael.clark@gmail.com"},
		{ID: "u10", Name: "Emma", Surname: "Garcia", Email: "emma.garcia@gmail.com"},
		{ID: "u11", Name: "William", Surname: "Hill", Email: "william.hill@gmail.com"},
		{ID: "u12", Name: "Ava", Surname: "Martinez", Email: "ava.martinez@gmail.com"},
	}

	bankAccounts := []model.BankAccount{
		{ID: "a1", Balance: 1500, Currency: model.RSD, Cards: []string{"Visa"}, Bank: banks[0], UserID: users[0].ID},
		{ID: "a2", Balance: 80000, Currency: model.EUR, Cards: []string{"MasterCard", "American Express"}, Bank: banks[1], UserID: users[1].ID},
		{ID: "a3", Balance: 300, Currency: model.RSD, Cards: []string{"Dina"}, Bank: banks[2], UserID: users[2].ID},
		{ID: "a4", Balance: 4500, Currency: model.EUR, Cards: []string{"Visa"}, Bank: banks[3], UserID: users[3].ID},
		{ID: "a5", Balance: 1200, Currency: model.EUR, Cards: []string{"MasterCard"}, Bank: banks[0], UserID: users[4].ID},
		{ID: "a6", Balance: 60000, Currency: model.RSD, Cards: []string{"Dina", "Visa"}, Bank: banks[1], UserID: users[5].ID},
		{ID: "a7", Balance: 900, Currency: model.RSD, Cards: []string{"American Express"}, Bank: banks[2], UserID: users[6].ID},
		{ID: "a8", Balance: 20000, Currency: model.EUR, Cards: []string{"Visa", "MasterCard"}, Bank: banks[3], UserID: users[7].ID},
		{ID: "a9", Balance: 700, Currency: model.RSD, Cards: []string{"Dina"}, Bank: banks[0], UserID: users[8].ID},
		{ID: "a10", Balance: 3500, Currency: model.EUR, Cards: []string{"MasterCard"}, Bank: banks[1], UserID: users[9].ID},
		{ID: "a11", Balance: 800, Currency: model.EUR, Cards: []string{"Visa"}, Bank: banks[2], UserID: users[10].ID},
		{ID: "a12", Balance: 40000, Currency: model.RSD, Cards: []string{"Dina"}, Bank: banks[3], UserID: users[11].ID},
		{ID: "a13", Balance: 1100, Currency: model.RSD, Cards: []string{"American Express"}, Bank: banks[0], UserID: users[0].ID},
		{ID: "a14", Balance: 55000, Currency: model.EUR, Cards: []string{"MasterCard"}, Bank: banks[1], UserID: users[1].ID},
		{ID: "a15", Balance: 750, Currency: model.RSD, Cards: []string{"Dina", "Visa"}, Bank: banks[2], UserID: users[2].ID},
		{ID: "a16", Balance: 6000, Currency: model.EUR, Cards: []string{"American Express", "MasterCard"}, Bank: banks[3], UserID: users[3].ID},
		{ID: "a17", Balance: 950, Currency: model.EUR, Cards: []string{"Visa"}, Bank: banks[0], UserID: users[4].ID},
		{ID: "a18", Balance: 30000, Currency: model.RSD, Cards: []string{"Dina", "MasterCard"}, Bank: banks[1], UserID: users[5].ID},
	}
	return banks, users, bankAccounts
}

func PutDataToState(ctx contractapi.TransactionContextInterface, data interface{}, id string) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState(id, dataJSON); err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
}
