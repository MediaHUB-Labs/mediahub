package handler

import (
	"fmt"
	"mediahub/dto"
	"mediahub/media/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	service *service.MediaService
}

func NewMediaHandler(service *service.MediaService) *MediaHandler {
	return &MediaHandler{service: service}
}

// getUserID extracts the authenticated user ID from the Gin context.
func getUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint)
}

// Testing Handler
func (h *MediaHandler) MediaHealth(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthReponse{
		Message: "Media Health: OK",
	})
}

// UploadMedia handles multipart file upload with metadata.
func (h *MediaHandler) UploadMedia(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "file is required",
		})
		return
	}

	var req dto.MediaCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid form data",
		})
		return
	}

	userID := getUserID(c)
	media, err := h.service.CreateMedia(c.Request.Context(), file, &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ApiResponse{
		Success: true,
		Message: "Media uploaded successfully",
		Data:    service.MediaToResponse(media),
	})
}

// ListMedia returns paginated media list with visibility enforcement.
func (h *MediaHandler) ListMedia(c *gin.Context) {
	var req dto.MediaListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid query parameters",
		})
		return
	}

	userID := getUserID(c)
	items, totalCount, err := h.service.ListMedia(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Convert to response DTOs
	var responseItems []dto.MediaResponse
	for _, item := range items {
		responseItems = append(responseItems, service.MediaToResponse(&item))
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Media list retrieved",
		Data: dto.MediaListResponse{
			Items:      responseItems,
			TotalCount: totalCount,
			Limit:      limit,
			Offset:     req.Offset,
		},
	})
}

// SearchMedia searches media by title/description.
func (h *MediaHandler) SearchMedia(c *gin.Context) {
	var req dto.MediaSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "query parameter 'q' is required",
		})
		return
	}

	userID := getUserID(c)
	items, totalCount, err := h.service.SearchMedia(c.Request.Context(), req.Query, userID, req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var responseItems []dto.MediaResponse
	for _, item := range items {
		responseItems = append(responseItems, service.MediaToResponse(&item))
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Search results",
		Data: dto.MediaListResponse{
			Items:      responseItems,
			TotalCount: totalCount,
			Limit:      limit,
			Offset:     req.Offset,
		},
	})
}

// GetCategories returns all distinct media categories.
func (h *MediaHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Categories retrieved",
		Data:    dto.CategoryResponse{Categories: categories},
	})
}

// GetMediaDetails returns full details for a single media item.
func (h *MediaHandler) GetMediaDetails(c *gin.Context) {
	var req dto.MediaIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "media id is required",
		})
		return
	}

	userID := getUserID(c)
	media, err := h.service.GetMedia(c.Request.Context(), req.ID, userID)
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
		Message: "Media details retrieved",
		Data:    service.MediaToResponse(media),
	})
}

// UpdateMediaMetadata updates metadata for a media item (owner only).
func (h *MediaHandler) UpdateMediaMetadata(c *gin.Context) {
	var req dto.MediaUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	userID := getUserID(c)
	media, err := h.service.UpdateMedia(c.Request.Context(), &req, userID)
	if err != nil {
		if err.Error() == "access denied: only the uploader can update this media" {
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
		Message: "Media updated",
		Data:    service.MediaToResponse(media),
	})
}

// DeleteMedia deletes a media item (owner only).
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	var req dto.MediaIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "media id is required",
		})
		return
	}

	userID := getUserID(c)
	err := h.service.DeleteMedia(c.Request.Context(), req.ID, userID)
	if err != nil {
		if err.Error() == "access denied: only the uploader can delete this media" {
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
		Message: "Media deleted",
	})
}

// GetUserVault returns private media items for the authenticated user (photos, documents).
func (h *MediaHandler) GetUserVault(c *gin.Context) {
	mediaType := c.Query("type")     // "image" or "document"
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	var limit, offset int
	fmt.Sscanf(limitStr, "%d", &limit)
	fmt.Sscanf(offsetStr, "%d", &offset)

	userID := getUserID(c)
	items, totalCount, err := h.service.GetUserVault(c.Request.Context(), userID, mediaType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var responseItems []dto.MediaResponse
	for _, item := range items {
		responseItems = append(responseItems, service.MediaToResponse(&item))
	}

	if limit <= 0 {
		limit = 20
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Vault items retrieved",
		Data: dto.MediaListResponse{
			Items:      responseItems,
			TotalCount: totalCount,
			Limit:      limit,
			Offset:     offset,
		},
	})
}

// RegenerateThumbnail triggers manual thumbnail generation.
func (h *MediaHandler) RegenerateThumbnail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid media id",
		})
		return
	}

	userID := getUserID(c)
	err = h.service.RegenerateThumbnail(c.Request.Context(), uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Thumbnail regeneration successful",
	})
}
