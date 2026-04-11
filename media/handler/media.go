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

func (h *MediaHandler) UploadMedia(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}

	var req dto.MediaCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid form data"})
		return
	}

	media, err := h.service.CreateMedia(c.Request.Context(), file, &req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, media)
}
