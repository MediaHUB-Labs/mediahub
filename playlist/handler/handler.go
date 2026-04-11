package handler

import (
	"mediahub/dto"
	"mediahub/playlist/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PlaylistHandler struct {
	service *service.PlaylistService
}

func NewPlaylistHandler(service *service.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{service: service}
}

// getUserID extracts the authenticated user ID from the Gin context.
func getUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint)
}

// CreatePlaylist handles POST /api/playlist/create
func (h *PlaylistHandler) CreatePlaylist(c *gin.Context) {
	var req dto.CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "name is required",
		})
		return
	}

	userID := getUserID(c)
	playlist, err := h.service.CreatePlaylist(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ApiResponse{
		Success: true,
		Message: "Playlist created",
		Data:    playlist,
	})
}

// ListPlaylists handles GET /api/playlist/list
func (h *PlaylistHandler) ListPlaylists(c *gin.Context) {
	userID := getUserID(c)
	playlists, err := h.service.ListUserPlaylists(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Playlists retrieved",
		Data:    playlists,
	})
}

// GetPlaylist handles GET /api/playlist/:id
func (h *PlaylistHandler) GetPlaylist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid playlist id",
		})
		return
	}

	userID := getUserID(c)
	playlist, err := h.service.GetPlaylist(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, dto.ApiResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusNotFound, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Playlist retrieved",
		Data:    playlist,
	})
}

// UpdatePlaylist handles PUT /api/playlist/:id
func (h *PlaylistHandler) UpdatePlaylist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid playlist id",
		})
		return
	}

	var req dto.UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	userID := getUserID(c)
	playlist, err := h.service.UpdatePlaylist(c.Request.Context(), uint(id), userID, &req)
	if err != nil {
		if err.Error() == "access denied: only the owner can update this playlist" {
			c.JSON(http.StatusForbidden, dto.ApiResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Playlist updated",
		Data:    playlist,
	})
}

// DeletePlaylist handles DELETE /api/playlist/:id
func (h *PlaylistHandler) DeletePlaylist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid playlist id",
		})
		return
	}

	userID := getUserID(c)
	err = h.service.DeletePlaylist(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err.Error() == "access denied: only the owner can delete this playlist" {
			c.JSON(http.StatusForbidden, dto.ApiResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Playlist deleted",
	})
}

// AddItem handles POST /api/playlist/:id/add
func (h *PlaylistHandler) AddItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid playlist id",
		})
		return
	}

	var req dto.PlaylistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "media_id is required",
		})
		return
	}

	userID := getUserID(c)
	item, err := h.service.AddItem(c.Request.Context(), uint(id), req.MediaID, userID)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "access denied: only the owner can modify this playlist" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ApiResponse{
		Success: true,
		Message: "Item added to playlist",
		Data:    item,
	})
}

// RemoveItem handles DELETE /api/playlist/:id/remove
func (h *PlaylistHandler) RemoveItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid playlist id",
		})
		return
	}

	var req dto.PlaylistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "media_id is required",
		})
		return
	}

	userID := getUserID(c)
	err = h.service.RemoveItem(c.Request.Context(), uint(id), req.MediaID, userID)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "access denied: only the owner can modify this playlist" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Item removed from playlist",
	})
}
