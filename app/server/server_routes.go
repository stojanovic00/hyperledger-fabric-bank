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

	router.POST("/initLedger", handler.InitLedger)
	//curl -X POST http://localhost:8080/login/someUserId
	router.POST("/login/:userID", handler.Login)

	s.Router = router
	return nil
}

func (s *Server) SetupUsers() map[string]string {
	users := map[string]string{
		"u1": "ORG1",
		"u2": "ORG2",
		// Add more users as needed
	}

	return users
}
