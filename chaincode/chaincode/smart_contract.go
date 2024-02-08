package chaincode

import (
	"chaincode/chaincode/utils"
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
	banks, users, bankAccounts := utils.InitializeData()

	for _, bank := range banks {
		if err := utils.PutDataToState(ctx, bank, bank.ID); err != nil {
			return err
		}
	}

	for _, user := range users {
		if err := utils.PutDataToState(ctx, user, user.ID); err != nil {
			return err
		}
	}

	for _, bankAcc := range bankAccounts {
		if err := utils.PutDataToState(ctx, bankAcc, bankAcc.ID); err != nil {
			return err
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
			convertedAmount = utils.EurToDin(amount)
		case model.RSD:
			convertedAmount = utils.DinToEur(amount)
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

func (s *SmartContract) MoneyWithdrawal(ctx contractapi.TransactionContextInterface, usrID string, bankAccount string, amount float64) (bool, error) {
	account, err := s.ReadBankAccount(ctx, bankAccount)
	if account.UserID != usrID {
		return false, fmt.Errorf("bank account with ID %s not found for user %s", bankAccount, usrID)
	}
	if err != nil {
		return false, err
	}

	if account.Balance < amount {
		return false, fmt.Errorf("Insufficient funds")
	}

	account.Balance = account.Balance - amount

	accountJSON, err := json.Marshal(account)
	if err != nil {
		return false, err
	}

	ctx.GetStub().PutState(account.ID, accountJSON)

	return true, nil
}

func (s *SmartContract) MoneyDepositToAccount(ctx contractapi.TransactionContextInterface, usrID string, bankAccountID string, amount float64) (bool, error) {
	account, err := s.ReadBankAccount(ctx, bankAccountID)
	if account.UserID != usrID {
		return false, fmt.Errorf("bank account with ID %s not found for user %s", bankAccountID, usrID)
	}
	account.Balance = account.Balance + amount

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

func (s *SmartContract) GetUserByBankAccountId(ctx contractapi.TransactionContextInterface, accId string) (model.User, error) {
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

func (s *SmartContract) GetAccountsByBankDesiredCurrencyAndBalance(ctx contractapi.TransactionContextInterface, bankId, currency, balanceThreshold string) ([]model.BankAccount, error) {
	var currencyEnum model.Currency
	switch strings.ToUpper(currency) {
	case "EUR":
		currencyEnum = model.EUR
	case "RSD":
		currencyEnum = model.RSD
	}

	balanceThresh, err := strconv.ParseFloat(balanceThreshold, 64)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	queryString := fmt.Sprintf(`{
			"selector":{
			  "bank" : {
					"ID" : "%s"
					 },
			  "currency":%d,
			  "balance": {"$gte": %f}
		   }
		 }`, bankId, currencyEnum, balanceThresh)

	queryResults, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer queryResults.Close()

	var bankAccounts []model.BankAccount
	for queryResults.HasNext() {
		queryResult, err := queryResults.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate query results: %v", err)
		}

		var user model.BankAccount
		if err := json.Unmarshal(queryResult.Value, &user); err != nil {
			return nil, fmt.Errorf("failed to bank account user: %v", err)
		}

		bankAccounts = append(bankAccounts, user)
	}

	return bankAccounts, nil
}

func (s *SmartContract) GetAccountByBankDesiredCurrencyAndMaxBalance(ctx contractapi.TransactionContextInterface, bankId, currency string) (model.BankAccount, error) {
	var currencyEnum model.Currency
	switch strings.ToUpper(currency) {
	case "EUR":
		currencyEnum = model.EUR
	case "RSD":
		currencyEnum = model.RSD
	}

	queryString := fmt.Sprintf(`{
       "selector": {
          "bank" : {
            "ID" : "%s"
          },
          "currency": %d
       },
       "sort": [{"balance": "desc"}],
       "limit": 1
     }`, bankId, currencyEnum)

	queryResults, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return model.BankAccount{}, fmt.Errorf("failed to execute query: %v", err)
	}
	defer queryResults.Close()

	if !queryResults.HasNext() {
		return model.BankAccount{}, fmt.Errorf("No accounts found")
	}

	queryResult, err := queryResults.Next()
	if err != nil {
		return model.BankAccount{}, fmt.Errorf("failed to get query result: %v", err)
	}

	var bankAccount model.BankAccount
	if err := json.Unmarshal(queryResult.Value, &bankAccount); err != nil {
		return model.BankAccount{}, fmt.Errorf("failed to unmarshal bank account: %v", err)
	}

	return bankAccount, nil
}
