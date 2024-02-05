package handler

import (
	jwtUtil "app/jwt"
	"app/utils"
	"log"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"

	"path/filepath"
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

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = utils.PopulateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract := network.GetContract("basic")
	println(contract)
	log.Println("--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger")
	// result, err := contract.SubmitTransaction("InitLedger")
	// if err != nil {
	// 	log.Fatalf("Failed to Submit transaction: %v", err)
	// }
	// log.Println(string(result))
}
