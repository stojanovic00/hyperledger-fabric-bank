package server

import (
	"app/handler"
	"app/jwt"
	"app/utils"
	"github.com/gin-gonic/gin"
)

func (s *Server) CreateRoutersAndSetRoutes() error {
	handler := handler.Handler{}
	handler.Users = utils.SetupUsers()
	handler.ChainCodes = map[string]string{
		"channel1": "bankchaincode1",
		"channel2": "bankchaincode2",
	}

	// ROUTES
	gin.SetMode(gin.ReleaseMode)
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
	router.GET("/max-account/:channel/:bank-id/:currency", jwt.AuthorizationMiddleware("ADMIN"), handler.GetAccountByBankDesiredCurrencyAndMaxBalance)

	s.Router = router
	return nil
}
