package server

import (
	"app/handler"
	"github.com/gin-gonic/gin"
)

func (s *Server) CreateRoutersAndSetRoutes() error {
	handler := handler.Handler{}
	handler.Users = s.SetupUsers()

	// ROUTES
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Endpoint doesn't exist"})
	})

	//curl -X POST http://localhost:8080/login/someUserId
	//curl -X POST http://localhost:8080/initLedger -H "Authorization:$token"
	router.POST("/initLedger", handler.InitLedger)
	router.POST("/login/:userID", handler.Login)

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
