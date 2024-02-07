package server

import (
	"app/handler"
	"app/jwt"
	"github.com/gin-gonic/gin"
)

func (s *Server) CreateRoutersAndSetRoutes() error {
	handler := handler.Handler{}
	handler.Users = s.SetupUsers()

	// ROUTES
	gin.SetMode(gin.DebugMode)
	router := gin.Default()

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Endpoint doesn't exist"})
	})

	//curl -X POST http://localhost:8080/login/someUserId
	//Unauthorized
	router.POST("/login/:userID", handler.Login)

	//Authorized
	//curl -X POST http://localhost:8080/init-ledger -H "Authorization:$token"
	router.Use(jwt.AuthMiddleware())
	router.POST("/init-ledger", handler.InitLedger)

	s.Router = router
	return nil
}

func (s *Server) SetupUsers() map[string]string {
	users := map[string]string{
		"u1": "org1",
		"u2": "org2",
		// Add more users as needed
	}

	return users
}
