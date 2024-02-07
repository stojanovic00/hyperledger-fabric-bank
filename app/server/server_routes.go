package server

import (
	"app/handler"
	"app/jwt"
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

	//curl -X POST http://localhost:8080/login/someUserId
	//Unauthorized
	router.POST("/login/:userID", handler.Login)
	router.POST("/add-user/:channel", handler.AddUser)

	//Authorized
	//curl -X POST http://localhost:8080/init-ledger -H "Authorization:$token"
	router.Use(jwt.AuthMiddleware())
	router.POST("/init-ledger/:channel", handler.InitLedger)

	s.Router = router
	return nil
}

func (s *Server) SetupUsers() map[string]string {
	users := map[string]string{
		"u1":  "org1",
		"u2":  "org2",
		"u3":  "org3",
		"u4":  "org4",
		"u5":  "org1",
		"u6":  "org2",
		"u7":  "org3",
		"u8":  "org4",
		"u9":  "org2",
		"u10": "org2",
		"u11": "org2",
		"u12": "org2",
	}
	return users
}
