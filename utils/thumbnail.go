package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GenerateThumbnail extracts a frame from a video file and saves it as a JPEG thumbnail.
// Uses FFmpeg's CPU-only decoder. Extracts a frame at ~10% of the duration or at 5 seconds.
// Returns nil error if FFmpeg is not installed (graceful skip).
func GenerateThumbnail(videoPath string, outputPath string) error {
	if !IsFFmpegAvailable() {
		LogToFile("FFmpeg not available, skipping thumbnail generation for: " + videoPath)
		return nil
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	// First, try to get the duration to seek to 10%
	seekTime := "5" // Default: grab frame at 5 seconds
	info, err := ProbeMediaInfo(videoPath)
	if err == nil && info != nil && info.DurationSec > 10 {
		// Seek to 10% of the video
		seekTime = fmt.Sprintf("%d", info.DurationSec/10)
	}

	// Generate thumbnail using FFmpeg
	// -ss: seek to position
	// -vframes 1: extract 1 frame
	// -vf scale: scale to max 480px width, keep aspect ratio
	// -q:v 2: high quality JPEG
	cmd := exec.Command("ffmpeg",
		"-ss", seekTime,
		"-i", videoPath,
		"-vframes", "1",
		"-vf", "scale=480:-1",
		"-q:v", "2",
		"-y", // Overwrite output
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		LogToFile(fmt.Sprintf("Thumbnail generation failed for %s: %s | Output: %s", videoPath, err.Error(), string(output)))
		return fmt.Errorf("thumbnail generation failed: %w", err)
	}

	LogToFile("Thumbnail generated: " + outputPath)
	return nil
}

// GetThumbnailPath returns the expected thumbnail path for a given media ID.
func GetThumbnailPath(basePath string, mediaID uint) string {
	return filepath.Join(basePath, "thumbnails", fmt.Sprintf("%d.jpg", mediaID))
}
