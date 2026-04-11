package handler

import (
	"mediahub/dto"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

// StreamMedia serves media files with HTTP 206 Range Request support.
// This enables native <video> and <audio> seeking without downloading the full file.
func (h *MediaHandler) StreamMedia(c *gin.Context) {
	// Get media ID from URL parameter
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
	media, err := h.service.GetMedia(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, dto.ApiResponse{
				Success: false,
				Error:   "access denied",
			})
			return
		}
		c.JSON(http.StatusNotFound, dto.ApiResponse{
			Success: false,
			Error:   "media not found",
		})
		return
	}

	// Use the original file path for direct streaming.
	// We DO NOT hijack this to the transcoded HLS manifest here, as that is handled by /api/media/hls/:id
	filePath := media.FilePath

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ApiResponse{
			Success: false,
			Error:   "file not found on disk",
		})
		return
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   "failed to open file",
		})
		return
	}
	defer file.Close()

	// Set content type - trust database but fallback to detection if empty
	contentType := media.MimeType
	if contentType == "" {
		buffer := make([]byte, 512)
		n, _ := file.ReadAt(buffer, 0)
		if n > 0 {
			contentType = http.DetectContentType(buffer[:n])
		} else {
			contentType = "application/octet-stream"
		}
	}

	// Set required headers for streaming and inline display
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline")
	c.Header("Accept-Ranges", "bytes")
	// Note: http.ServeContent handles Range requests, ETag, and Content-Length automatically.
	http.ServeContent(c.Writer, c.Request, fileInfo.Name(), fileInfo.ModTime(), file)
}

// ServeThumbnail serves a media item's thumbnail image.
func (h *MediaHandler) ServeThumbnail(c *gin.Context) {
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
	media, err := h.service.GetMedia(c.Request.Context(), uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ApiResponse{
			Success: false,
			Error:   "media not found",
		})
		return
	}

	if media.ThumbnailPath == "" {
		c.JSON(http.StatusNotFound, dto.ApiResponse{
			Success: false,
			Error:   "no thumbnail available",
		})
		return
	}

	// Verify thumbnail file exists
	if _, err := os.Stat(media.ThumbnailPath); err != nil {
		c.JSON(http.StatusNotFound, dto.ApiResponse{
			Success: false,
			Error:   "thumbnail file not found",
		})
		return
	}

	c.Header("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	c.File(media.ThumbnailPath)
}

// ServeHLSManifest serves the HLS .m3u8 manifest file for transcoded media.
func (h *MediaHandler) ServeHLSManifest(c *gin.Context) {
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
	media, err := h.service.GetMedia(c.Request.Context(), uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ApiResponse{
			Success: false,
			Error:   "media not found",
		})
		return
	}

	if !media.IsTranscoded || media.TranscodedPath == "" {
		c.JSON(http.StatusNotFound, dto.ApiResponse{
			Success: false,
			Error:   "media has not been transcoded",
		})
		return
	}

	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.File(media.TranscodedPath)
}
