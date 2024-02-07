package handler

import (
	jwtUtil "app/jwt"
	"app/utils"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"log"
	"net/http"
)

type Handler struct {
	Users      map[string]string
	ChainCodes map[string]string
}

func (h *Handler) Login(ctx *gin.Context) {
	userID := ctx.Param("userID")
	if userID == "" {
		ctx.JSON(400, gin.H{"error": "UserID is required"})
		return
	}

	token, err := jwtUtil.GenerateJWT(userID)
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

	userOrg := h.Users[userId]

	channel := ctx.Param("channel")
	if channel == "" {
		ctx.JSON(400, gin.H{"error": "channel is required"})
		return
	}

	chaincodeId := h.ChainCodes[channel]

	wallet, err := utils.CreateWallet(userId, userOrg)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to create or populate wallet"})
		return
	}

	gw, err := utils.ConnectToGateway(wallet, userOrg)
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
