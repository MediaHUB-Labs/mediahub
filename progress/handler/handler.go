package handler

import (
	"mediahub/dto"
	"mediahub/progress/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProgressHandler struct {
	service *service.ProgressService
}

func NewProgressHandler(service *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{service: service}
}

// getUserID extracts the authenticated user ID from the Gin context.
func getUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint)
}

// SaveProgress handles POST /api/progress/save
func (h *ProgressHandler) SaveProgress(c *gin.Context) {
	var req dto.SaveProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid request: media_id and position_sec are required",
		})
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ApiResponse{
			Success: false,
			Error:   "authentication required",
		})
		return
	}

	progress, err := h.service.SaveProgress(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Progress saved",
		Data:    progress,
	})
}

// ListContinueWatching handles GET /api/progress/continue
func (h *ProgressHandler) ListContinueWatching(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ApiResponse{
			Success: false,
			Error:   "authentication required",
		})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	items, err := h.service.GetContinueWatching(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Continue watching list",
		Data:    items,
	})
}

// RemoveProgress handles DELETE /api/progress/clear
func (h *ProgressHandler) RemoveProgress(c *gin.Context) {
	var req dto.ClearProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "media_id is required",
		})
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ApiResponse{
			Success: false,
			Error:   "authentication required",
		})
		return
	}

	err := h.service.ClearProgress(c.Request.Context(), userID, req.MediaID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Progress cleared",
	})
}
