package server

import (
	"app/handler"
	"app/jwt"
	"app/model"

	"github.com/gin-gonic/gin"
)

func (s *Server) CreateRoutersAndSetRoutes() error {
	handler := handler.Handler{}
	handler.Users = s.SetupUsers()
	handler.ChainCodes = map[string]string{
		"channel1": "bankchaincode1",
		"channel2": "bankchaincode2",
	}

	// ROUTES
	gin.SetMode(gin.DebugMode)
	router := gin.Default()

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Endpoint doesn't exist"})
	})

	router.POST("/login/:userID", handler.Login)

	router.Use(jwt.AuthenticationMiddleware())
	router.POST("/add-user/:channel", jwt.AuthorizationMiddleware("ADMIN"), handler.AddUser)
	router.POST("/create-bank-account/:channel", jwt.AuthorizationMiddleware("USER"), handler.CreateBankAccount)
	router.POST("/transfer-money/:channel", jwt.AuthorizationMiddleware("USER"), handler.TransferMoney)
	router.POST("/money-withdrawal/:channel", jwt.AuthorizationMiddleware("USER"), handler.MoneyWithdrawal)
	router.POST("/money-deposit/:channel", jwt.AuthorizationMiddleware("USER"), handler.MoneyDepositToAccount)
	router.GET("/search/:channel/:by/:param1/:param2", jwt.AuthorizationMiddleware("ADMIN"), handler.Query)
	router.GET("/search-accounts/:channel/:bank-id/:currency/:balance-thresh", jwt.AuthorizationMiddleware("ADMIN"), handler.GetAccountsByBankDesiredCurrencyAndBalance)

	s.Router = router
	return nil
}

func (s *Server) SetupUsers() map[string]model.UserInfo {
	users := map[string]model.UserInfo{
		"u1":  {UserId: "u1", Organization: "org1", Admin: false},
		"u2":  {UserId: "u2", Organization: "org2", Admin: false},
		"u3":  {UserId: "u3", Organization: "org3", Admin: false},
		"u4":  {UserId: "u4", Organization: "org4", Admin: false},
		"u5":  {UserId: "u5", Organization: "org1", Admin: false},
		"u6":  {UserId: "u6", Organization: "org2", Admin: false},
		"u7":  {UserId: "u7", Organization: "org3", Admin: false},
		"u8":  {UserId: "u8", Organization: "org4", Admin: false},
		"u9":  {UserId: "u9", Organization: "org1", Admin: false},
		"u10": {UserId: "u10", Organization: "org2", Admin: false},
		"u11": {UserId: "u11", Organization: "org3", Admin: false},
		"u12": {UserId: "u12", Organization: "org4", Admin: false},
		//Admins
		"s1": {UserId: "s1", Organization: "org1", Admin: true},
		"s2": {UserId: "s2", Organization: "org2", Admin: true},
		"s3": {UserId: "s3", Organization: "org3", Admin: true},
		"s4": {UserId: "s4", Organization: "org4", Admin: true},
	}
	return users
}
