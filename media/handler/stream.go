package handler

import (
	"fmt"
	"mediahub/dto"
	"net/http"
	"os"
	"strconv"
	"strings"

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

	// Verify file exists
	filePath := media.FilePath

	// If transcoded version exists and is an HLS manifest, redirect to that
	if media.IsTranscoded && media.TranscodedPath != "" {
		filePath = media.TranscodedPath
	}

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

	fileSize := fileInfo.Size()

	// Set content type
	contentType := media.MimeType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Check for Range header
	rangeHeader := c.GetHeader("Range")

	if rangeHeader == "" {
		// No range request — serve the full file
		c.Header("Content-Length", fmt.Sprintf("%d", fileSize))
		c.Header("Content-Type", contentType)
		c.Header("Accept-Ranges", "bytes")
		c.File(filePath)
		return
	}

	// Parse Range header: "bytes=start-end"
	rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.SplitN(rangeHeader, "-", 2)

	var start, end int64

	if parts[0] != "" {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			c.JSON(http.StatusRequestedRangeNotSatisfiable, dto.ApiResponse{
				Success: false,
				Error:   "invalid range",
			})
			return
		}
	}

	if len(parts) > 1 && parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			c.JSON(http.StatusRequestedRangeNotSatisfiable, dto.ApiResponse{
				Success: false,
				Error:   "invalid range",
			})
			return
		}
	} else {
		// If no end specified, serve up to 2MB chunk or end of file
		end = start + 2*1024*1024 - 1
		if end >= fileSize {
			end = fileSize - 1
		}
	}

	// Validate range
	if start < 0 || start >= fileSize || end >= fileSize || start > end {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		c.JSON(http.StatusRequestedRangeNotSatisfiable, dto.ApiResponse{
			Success: false,
			Error:   "range not satisfiable",
		})
		return
	}

	contentLength := end - start + 1

	// Seek to start position
	if _, err := file.Seek(start, 0); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Success: false,
			Error:   "failed to seek file",
		})
		return
	}

	// Set response headers for partial content
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Length", fmt.Sprintf("%d", contentLength))
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-cache")

	// Send 206 Partial Content
	c.DataFromReader(http.StatusPartialContent, contentLength, contentType, file, nil)
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
