package handler

import (
	"app/dto"
	jwtUtil "app/jwt"
	"app/model"
	"app/utils"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

type Handler struct {
	Users      map[string]model.UserInfo
	ChainCodes map[string]string
}

type Currency int

func (h *Handler) Login(ctx *gin.Context) {
	userID := ctx.Param("userID")
	if userID == "" {
		ctx.JSON(400, gin.H{"error": "UserID is required"})
		return
	}

	userInfo, exists := h.Users[userID]
	if !exists {
		ctx.JSON(404, gin.H{"error": "user not found"})
		return
	}
	var role string
	if userInfo.Admin {
		role = "ADMIN"
	} else {
		role = "USER"
	}

	token, err := jwtUtil.GenerateJWT(userInfo.UserId, role)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to generate JWT token"})
		return
	}

	ctx.JSON(200, gin.H{"token": token})
}

func (h *Handler) InitLedger(ctx *gin.Context) {
	userIdEntry, ok := ctx.Get("userId")
	if !ok {
		ctx.JSON(400, gin.H{"error": "no auth parameters provided"})
		return
	}
	userId := userIdEntry.(string)

	userInfo := h.Users[userId]

	channel := ctx.Param("channel")
	if channel == "" {
		ctx.JSON(400, gin.H{"error": "channel is required"})
		return
	}

	chaincodeId := h.ChainCodes[channel]

	wallet, err := utils.CreateWallet(userId, userInfo.Organization)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to create or populate wallet"})
		return
	}

	gw, err := utils.ConnectToGateway(wallet, userInfo.Organization)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to connect to gateway"})
		return
	}
	defer gw.Close()

	network, err := gw.GetNetwork(channel)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to get network"})
		return
	}

	contract := network.GetContract(chaincodeId)
	log.Println("Submit Transaction: InitLedger")
	_, err = h.submitTransaction(contract, "InitLedger")
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to submit transaction"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "ledger initialized"})
}

func (h *Handler) submitTransaction(contract *gateway.Contract, transaction string) ([]byte, error) {
	return contract.SubmitTransaction(transaction)
}

func (h *Handler) AddUser(ctx *gin.Context) {
	//Admins info
	userIdEntry, ok := ctx.Get("userId")
	if !ok {
		ctx.JSON(400, gin.H{"error": "no auth parameters provided"})
		return
	}
	adminId := userIdEntry.(string)
	adminUserInfo := h.Users[adminId]

	//Chaincode info
	var user dto.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "couldn't resolve body"})
		return
	}

	channel := ctx.Param("channel")
	if channel == "" {
		ctx.JSON(400, gin.H{"error": "channel is required"})
		return
	}
	chaincodeId := h.ChainCodes[channel]

	wallet, err := utils.CreateWallet(adminId, adminUserInfo.Organization)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to create or populate wallet"})
		return
	}

	gw, err := utils.ConnectToGateway(wallet, adminUserInfo.Organization)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to connect to gateway"})
		return
	}
	defer gw.Close()

	network, err := gw.GetNetwork(channel)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to get network"})
		return
	}

	contract := network.GetContract(chaincodeId)
	log.Println("Submit Transaction: AddUser")
	_, err = contract.SubmitTransaction("AddUser", user.Id, user.Name, user.Surname, user.Email)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "user already exists"})
		return
	}

	//Register it also in SDK app
	newUserInfo := model.UserInfo{
		UserId:       user.Id,
		Organization: adminUserInfo.Organization,
		Admin:        false,
	}
	h.Users[user.Id] = newUserInfo

	ctx.JSON(http.StatusOK, gin.H{"message": "added user"})
}

func (h *Handler) CreateBankAccount(ctx *gin.Context) {
	var bankAccount struct {
		Id       string   `json:"id"`
		Currency string   `json:"currency"`
		Cards    []string `json:"cards"`
		BankId   string   `json:"bankId"`
		UserID   string   `json:"userID"`
	}

	if err := ctx.ShouldBindJSON(&bankAccount); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "couldn't resolve body"})
		return
	}

	channel := ctx.Param("channel")
	if channel == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "channel is required"})
		return
	}
	chaincodeId := h.ChainCodes[channel]

	userId := bankAccount.UserID
	userInfo := h.Users[userId]

	wallet, err := utils.CreateWallet(userId, userInfo.Organization)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create or populate wallet"})
		return
	}

	gw, err := utils.ConnectToGateway(wallet, userInfo.Organization)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to gateway"})
		return
	}
	defer gw.Close()

	network, err := gw.GetNetwork(channel)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get network"})
		return
	}

	contract := network.GetContract(chaincodeId)
	log.Println("Submit Transaction: CreateBankAccount")
	_, err = contract.SubmitTransaction("CreateBankAccount", bankAccount.Id, bankAccount.Currency, strings.Join(bankAccount.Cards, ","), bankAccount.BankId, bankAccount.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "bank account created"})
}

func (h *Handler) TransferMoney(ctx *gin.Context) {
	var transfer struct {
		SrcAccount      string `json:"srcAccount"`
		DstAccount      string `json:"dstAccount"`
		AmountStr       string `json:"amountStr"`
		ConfirmationStr string `json:"confirmationStr"`
	}

	if err := ctx.ShouldBindJSON(&transfer); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "couldn't resolve body"})
		return
	}

	channel := ctx.Param("channel")
	if channel == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "channel is required"})
		return
	}
	chaincodeID := h.ChainCodes[channel]

	userIdContext, _ := ctx.Get("userId")
	userId := fmt.Sprintf("%v", userIdContext)

	userInfo := h.Users[userId]

	wallet, err := utils.CreateWallet(userId, userInfo.Organization)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create or populate wallet"})
		return
	}

	gw, err := utils.ConnectToGateway(wallet, userInfo.Organization)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to gateway"})
		return
	}
	defer gw.Close()

	network, err := gw.GetNetwork(channel)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get network"})
		return
	}

	contract := network.GetContract(chaincodeID)
	log.Println("Submit Transaction: TransferMoney")
	response, err := contract.SubmitTransaction("TransferMoney", transfer.SrcAccount, transfer.DstAccount, transfer.AmountStr, transfer.ConfirmationStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	responseStr := string(response)
	var resultMsg string
	if responseStr == "true" {
		resultMsg = "I received true"
	} else {
		resultMsg = "I received false"
	}
	log.Println(resultMsg)

	ctx.JSON(http.StatusOK, gin.H{"message": resultMsg})
}

func (h *Handler) Query(ctx *gin.Context) {
	by := ctx.Param("by")
	param1 := ctx.Param("param1")

	if by == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'by' is required"})
		return
	}

	if param1 == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'param1' is required"})
		return
	}

	channel := ctx.Param("channel")
	if channel == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "channel is required"})
		return
	}

	chaincodeID := h.ChainCodes[channel]

	userID := ctx.Get("userId")
	userInfo := h.Users[userID]
	wallet, err := utils.CreateWallet(userID, userInfo.Organization)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create or populate wallet"})
		return
	}
	defer wallet.Close()

	gw, err := utils.ConnectToGateway(wallet, userInfo.Organization)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to gateway"})
		return
	}
	defer gw.Close()

	network, err := gw.GetNetwork(channel)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get network"})
		return
	}

	contract := network.GetContract(chaincodeID)

	var result interface{}
	switch by {
	case "name":
		result, err = contract.EvaluateTransaction("GetUsersByName", param1)
	case "surname":
		result, err = contract.EvaluateTransaction("GetUsersBySurname", param1)
	case "account":
		result, err = contract.EvaluateTransaction("GetByBankAccountId", param1)
	case "both":
		param2 := ctx.Param("param2")
		result, err = contract.EvaluateTransaction("GetUsersBySurnameAndEmail", param1, param2)
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported query parameter 'by'"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": result})
}
