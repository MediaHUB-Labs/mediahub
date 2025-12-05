package handler

import (
	"mediahub/dto"
	"mediahub/media/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	service *service.MediaService
}

func NewMediaHandler(service *service.MediaService) *MediaHandler {
	return &MediaHandler{service: service}
}

// Testing Handler
func (h *MediaHandler) MediaHealth(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthReponse{
		Message: "Media Health: OK",
	})
}
