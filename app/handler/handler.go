package handler

import (
	jwtUtil "app/jwt"
	"app/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"log"
)

type Handler struct {
	Users map[string]string
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
	token, err := jwtUtil.ExtractAndValidateToken(ctx)
	if err != nil {
		ctx.JSON(401, gin.H{"error": "Unauthorized - Invalid token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		ctx.JSON(401, gin.H{"error": "Unauthorized - Invalid token claims"})
		return
	}

	userId := claims["userId"].(string)
	userOrg := h.Users[userId]
	channel := "channel1"
	chaincodeId := "bankchaincode1"
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
	log.Println("--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger")
	result, err := h.submitTransaction(contract, "InitLedger")
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to submit transaction"})
		return
	}

	log.Println(string(result))
}

func (h *Handler) submitTransaction(contract *gateway.Contract, transaction string) ([]byte, error) {
	return contract.SubmitTransaction(transaction)
}
