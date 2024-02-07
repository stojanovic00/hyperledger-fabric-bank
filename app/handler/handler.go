package handler

import (
	"app/dto"
	jwtUtil "app/jwt"
	"app/model"
	"app/utils"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"log"
	"net/http"
)

type Handler struct {
	Users      map[string]model.UserInfo
	ChainCodes map[string]string
}

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
