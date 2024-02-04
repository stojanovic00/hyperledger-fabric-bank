package chaincode

import (
	"chaincode/model"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

func (s *SmartContract) InitLedger() {
	banks := []model.Bank{
		{ID: "bank1", Name: "UniCredit", Headquarters: "Linz, Austria", Since: 1969, PIB: 138429230},
		{ID: "bank2", Name: "Raiffeisen Bank", Headquarters: "Vienna, Austria", Since: 1927, PIB: 537891234},
		{ID: "bank3", Name: "Erste Group", Headquarters: "Vienna, Austria", Since: 1819, PIB: 987654321},
		{ID: "bank4", Name: "OTP Bank", Headquarters: "Budapest, Hungary", Since: 1949, PIB: 654321789},
	}

	users := []model.User{
		{ID: "user1", Name: "John", Surname: "Doe", Email: "john.doe@gmail.com"},
		{ID: "user2", Name: "Alice", Surname: "Smith", Email: "alice.smith@gmail.com"},
		{ID: "user3", Name: "Bob", Surname: "Johnson", Email: "bob.johnson@gmail.com"},
		{ID: "user4", Name: "Eva", Surname: "Williams", Email: "eva.williams@gmail.com"},
		{ID: "user5", Name: "Daniel", Surname: "Miller", Email: "daniel.miller@gmail.com"},
		{ID: "user6", Name: "Sophia", Surname: "Brown", Email: "sophia.brown@gmail.com"},
		{ID: "user7", Name: "Matthew", Surname: "Davis", Email: "matthew.davis@gmail.com"},
		{ID: "user8", Name: "Olivia", Surname: "Jones", Email: "olivia.jones@gmail.com"},
		{ID: "user9", Name: "Michael", Surname: "Clark", Email: "michael.clark@gmail.com"},
		{ID: "user10", Name: "Emma", Surname: "Garcia", Email: "emma.garcia@gmail.com"},
		{ID: "user11", Name: "William", Surname: "Hill", Email: "william.hill@gmail.com"},
		{ID: "user12", Name: "Ava", Surname: "Martinez", Email: "ava.martinez@gmail.com"},
	}

	bankAccounts := []model.BankAccount{
		{ID: "bc1", Balance: 1500, Currency: model.RSD, Cards: []string{"Visa"}, Bank: banks[0], UserID: users[0].ID},
		{ID: "bc2", Balance: 80000, Currency: model.EUR, Cards: []string{"MasterCard", "American Express"}, Bank: banks[1], UserID: users[1].ID},
		{ID: "bc3", Balance: 300, Currency: model.RSD, Cards: []string{"Dina"}, Bank: banks[2], UserID: users[2].ID},
		{ID: "bc4", Balance: 4500, Currency: model.EUR, Cards: []string{"Visa"}, Bank: banks[3], UserID: users[3].ID},
		{ID: "bc5", Balance: 1200, Currency: model.EUR, Cards: []string{"MasterCard"}, Bank: banks[0], UserID: users[4].ID},
		{ID: "bc6", Balance: 60000, Currency: model.RSD, Cards: []string{"Dina", "Visa"}, Bank: banks[1], UserID: users[5].ID},
		{ID: "bc7", Balance: 900, Currency: model.RSD, Cards: []string{"American Express"}, Bank: banks[2], UserID: users[6].ID},
		{ID: "bc8", Balance: 20000, Currency: model.EUR, Cards: []string{"Visa", "MasterCard"}, Bank: banks[3], UserID: users[7].ID},
		{ID: "bc9", Balance: 700, Currency: model.RSD, Cards: []string{"Dina"}, Bank: banks[0], UserID: users[8].ID},
		{ID: "bc10", Balance: 3500, Currency: model.EUR, Cards: []string{"MasterCard"}, Bank: banks[1], UserID: users[9].ID},
		{ID: "bc11", Balance: 800, Currency: model.EUR, Cards: []string{"Visa"}, Bank: banks[2], UserID: users[10].ID},
		{ID: "bc12", Balance: 40000, Currency: model.RSD, Cards: []string{"Dina"}, Bank: banks[3], UserID: users[11].ID},
		{ID: "bc13", Balance: 1100, Currency: model.RSD, Cards: []string{"American Express"}, Bank: banks[0], UserID: users[0].ID},
		{ID: "bc14", Balance: 55000, Currency: model.EUR, Cards: []string{"MasterCard"}, Bank: banks[1], UserID: users[1].ID},
		{ID: "bc15", Balance: 750, Currency: model.RSD, Cards: []string{"Dina", "Visa"}, Bank: banks[2], UserID: users[2].ID},
		{ID: "bc16", Balance: 6000, Currency: model.EUR, Cards: []string{"American Express", "MasterCard"}, Bank: banks[3], UserID: users[3].ID},
		{ID: "bc17", Balance: 950, Currency: model.EUR, Cards: []string{"Visa"}, Bank: banks[0], UserID: users[4].ID},
		{ID: "bc18", Balance: 30000, Currency: model.RSD, Cards: []string{"Dina", "MasterCard"}, Bank: banks[1], UserID: users[5].ID},
	}
}
