package chaincode

import (
	"chaincode/model"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	banks := []model.Bank{
		{ID: "b1", Name: "UniCredit", Headquarters: "Linz, Austria", Since: 1969, PIB: 138429230},
		{ID: "b_2", Name: "Raiffeisen Bank", Headquarters: "Vienna, Austria", Since: 1927, PIB: 537891234},
		{ID: "b_3", Name: "Erste Group", Headquarters: "Vienna, Austria", Since: 1819, PIB: 987654321},
		{ID: "b_4", Name: "OTP Bank", Headquarters: "Budapest, Hungary", Since: 1949, PIB: 654321789},
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

	for _, bank := range banks {
		bankJSON, err := json.Marshal(bank)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(bank.ID, bankJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	for _, user := range users {
		userJSON, err := json.Marshal(user)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(user.ID, userJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	for _, bankAcc := range bankAccounts {
		bankAccJSON, err := json.Marshal(bankAcc)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(bankAcc.ID, bankAccJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

func (s *SmartContract) CreateBankAccount(ctx contractapi.TransactionContextInterface, id string, currency string, cards string, bankId string, userID string) error {

	accountExists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if accountExists {
		return fmt.Errorf("the bank account with id %s already exists", id)
	}

	userExists, err := s.AssetExists(ctx, userID)
	if err != nil {
		return err
	}
	if !userExists {
		return fmt.Errorf("no registered user with id %s", userID)
	}

	bank, err := s.ReadBank(ctx, bankId)
	if err != nil {
		return err
	}

	currency_converted, _ := StringToCurrency(currency)

	bankAccount := model.BankAccount{
		ID:       id,
		Currency: currency_converted,
		Balance:  0.0,
		Cards:    strings.Split(cards, ","),
		Bank:     *bank,
		UserID:   userID,
	}

	bankAccountJSON, err := json.Marshal(bankAccount)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, bankAccountJSON)
}

func (s *SmartContract) ReadBank(ctx contractapi.TransactionContextInterface, id string) (*model.Bank, error) {
	bankJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if bankJSON == nil {
		return nil, fmt.Errorf("the bank with id %s does not exist", id)
	}

	var bank model.Bank
	err = json.Unmarshal(bankJSON, &bank)

	if err != nil {
		return nil, err
	}
	return &bank, nil
}

func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

func (s *SmartContract) TransferMoney(ctx contractapi.TransactionContextInterface, srcAccount string, dstAccount string, amountStr string, confirmationStr string) (bool, error) {
	sourceAccount, err := s.ReadBankAccount(ctx, srcAccount)
	if err != nil {
		return false, err
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return false, fmt.Errorf("failed to convert amount to float64: %v", err)
	}

	confirmation, err := strconv.ParseBool(confirmationStr)
	if err != nil {
		return false, fmt.Errorf("failed to convert confirmation to boolean: %v", err)
	}

	if sourceAccount.Balance < amount {
		return false, fmt.Errorf("not enough money")
	}

	destAccount, err := s.ReadBankAccount(ctx, dstAccount)
	if err != nil {
		return false, err
	}

	if sourceAccount.Currency != destAccount.Currency && !confirmation {
		return false, nil
	} else if sourceAccount.Currency != destAccount.Currency && confirmation {

		var convertedAmount float64

		switch sourceAccount.Currency {
		case model.EUR:
			convertedAmount = EurToDin(amount)
		case model.RSD:
			convertedAmount = DinToEur(amount)
		}

		sourceAccount.Balance -= amount
		destAccount.Balance += convertedAmount

	} else {
		destAccount.Balance += amount
		sourceAccount.Balance -= amount
	}

	sourceAccountJSON, err := json.Marshal(sourceAccount)
	if err != nil {
		return false, err
	}
	destAccountJSON, err := json.Marshal(destAccount)
	if err != nil {
		return false, err
	}

	ctx.GetStub().PutState(sourceAccount.ID, sourceAccountJSON)
	ctx.GetStub().PutState(destAccount.ID, destAccountJSON)

	return true, nil
}

func (s *SmartContract) MoneyWithdrawal(ctx contractapi.TransactionContextInterface, acc string, amount float64) (bool, error) {
	account, err := s.ReadBankAccount(ctx, acc)
	fmt.Printf("MoneyWithdrawal: ReadBankAccount result - Account: %+v, Error: %+v\n", account, err)

	if err != nil {
		return false, err
	}

	if account.Balance < amount {
		return false, fmt.Errorf("Insufficient funds")
	}

	account.Balance = account.Balance - amount

	fmt.Printf("MoneyWithdrawal: Updated balance: %f\n", account.Balance)

	accountJSON, err := json.Marshal(account)
	if err != nil {
		return false, err
	}

	ctx.GetStub().PutState(account.ID, accountJSON)

	fmt.Printf("MoneyWithdrawal: ReadBankAccount result - Account: %+v, Error: %+v\n", account, err)
	return true, nil
}

func (s *SmartContract) MoneyDepositToAccount(ctx contractapi.TransactionContextInterface, acc string, amount float64) (bool, error) {
	account, err := s.ReadBankAccount(ctx, acc)
	fmt.Printf("MoneyDepositToAccount: ReadBankAccount result - Account: %+v, Error: %+v\n", account, err)

	account.Balance = account.Balance + amount

	fmt.Printf("MoneyDepositToAccount: Updated balance: %f\n", account.Balance)

	if err != nil {
		return false, err
	}

	accountJSON, err := json.Marshal(account)
	if err != nil {
		return false, err
	}

	ctx.GetStub().PutState(account.ID, accountJSON)

	return true, nil
}

func (s *SmartContract) Exists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	someJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return someJSON != nil, nil
}

func (s *SmartContract) ReadBankAccount(ctx contractapi.TransactionContextInterface, id string) (*model.BankAccount, error) {
	accJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if accJSON == nil {
		return nil, fmt.Errorf("the bank account with id %s does not exist", id)
	}

	var bankAccount model.BankAccount
	err = json.Unmarshal(accJSON, &bankAccount)

	if err != nil {
		return nil, err
	}
	return &bankAccount, nil
}

func (s *SmartContract) AddUser(ctx contractapi.TransactionContextInterface, id, name, surname, email string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the user %s already exists", id)
	}
	user := model.User{
		ID:      id,
		Name:    name,
		Surname: surname,
		Email:   email,
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, userJson)
}

func StringToCurrency(currencyStr string) (model.Currency, error) {
	switch currencyStr {
	case "EUR":
		return model.EUR, nil
	case "RSD":
		return model.RSD, nil
	default:
		return 0, fmt.Errorf("invalid currency string: %s", currencyStr)
	}
}
func (s *SmartContract) GetUsersByName(ctx contractapi.TransactionContextInterface, name string) ([]model.User, error) {
	queryString := fmt.Sprintf(`{
		"selector": {
			"name": "%s"
		}
	}`, name)

	queryResults, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer queryResults.Close()

	var users []model.User
	for queryResults.HasNext() {
		queryResult, err := queryResults.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate query results: %v", err)
		}

		var user model.User
		if err := json.Unmarshal(queryResult.Value, &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %v", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (s *SmartContract) GetUsersBySurname(ctx contractapi.TransactionContextInterface, surname string) ([]model.User, error) {
	queryString := fmt.Sprintf(`{
		"selector": {
			"surname": "%s"
		}
	}`, surname)

	queryResults, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer queryResults.Close()

	var users []model.User
	for queryResults.HasNext() {
		queryResult, err := queryResults.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate query results: %v", err)
		}

		var user model.User
		if err := json.Unmarshal(queryResult.Value, &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %v", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (s *SmartContract) GetByBankAccountId(ctx contractapi.TransactionContextInterface, accId string) (model.User, error) {
	accountJSON, err := ctx.GetStub().GetState(accId)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to read from world state: %v", err)
	}

	var account model.BankAccount
	err = json.Unmarshal(accountJSON, &account)

	if err != nil {
		return model.User{}, err
	}

	userJSON, err := ctx.GetStub().GetState(account.UserID)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to read from world state: %v", err)
	}

	var user model.User
	err = json.Unmarshal(userJSON, &user)

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (s *SmartContract) GetUsersBySurnameAndEmail(ctx contractapi.TransactionContextInterface, surname, email string) ([]model.User, error) {
	queryString := fmt.Sprintf(`{
		"selector": {
			"surname": "%s",
			"email": "%s"
		}
	}`, surname, email)

	queryResults, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer queryResults.Close()

	var users []model.User
	for queryResults.HasNext() {
		queryResult, err := queryResults.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate query results: %v", err)
		}

		var user model.User
		if err := json.Unmarshal(queryResult.Value, &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %v", err)
		}

		users = append(users, user)
	}

	return users, nil
}
