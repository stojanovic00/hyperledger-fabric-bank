package server

import (
	"app/config"
	"github.com/gin-gonic/gin"
	"log"
)

type Server struct {
	Config config.Config
	Router *gin.Engine
}

func NewServer() (*Server, error) {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	server := &Server{
		Config: config,
	}

	err = server.CreateRoutersAndSetRoutes()
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	return server, nil
}
