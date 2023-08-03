package handler

import (
	"github.com/denizumutdereli/golang-K8-microservice-probs/internal/config"
	"github.com/denizumutdereli/golang-K8-microservice-probs/internal/service"
	"github.com/gin-gonic/gin"
)

type RestHandler struct {
	StreamService service.StreamService
	Config        *config.Config
}

func NewRestHandler(s service.StreamService, cfg *config.Config) *RestHandler {
	return &RestHandler{StreamService: s, Config: cfg}
}

func (s *RestHandler) Live(c *gin.Context) {

	statusCode, ok, err := s.StreamService.Live(c.Request.Context())
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, gin.H{"status": ok})
}

func (s *RestHandler) Read(c *gin.Context) {

	statusCode, ok, err := s.StreamService.Live(c.Request.Context())
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, gin.H{"status": ok})
}
