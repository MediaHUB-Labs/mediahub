package utils

import (
	"encoding/json"
	"fmt"
	"mediahub/dto"
	"os/exec"
	"strconv"
	"strings"
)

// IsFFmpegAvailable checks if ffmpeg is installed and accessible in PATH.
func IsFFmpegAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// IsFFprobeAvailable checks if ffprobe is installed and accessible in PATH.
func IsFFprobeAvailable() bool {
	_, err := exec.LookPath("ffprobe")
	return err == nil
}

// ffprobeOutput represents the JSON output structure from ffprobe.
type ffprobeOutput struct {
	Streams []struct {
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Duration  string `json:"duration"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
}

// ProbeMediaInfo uses ffprobe to extract media metadata (duration, resolution, codec).
// Returns nil info and nil error if ffprobe is not installed (graceful degradation).
func ProbeMediaInfo(filePath string) (*dto.MediaInfo, error) {
	if !IsFFprobeAvailable() {
		LogToFile("FFprobe not available, skipping media probe for: " + filePath)
		return nil, nil
	}

	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed for %s: %w", filePath, err)
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	info := &dto.MediaInfo{}

	// Extract duration from format (more reliable)
	if probe.Format.Duration != "" {
		if dur, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
			info.DurationSec = uint(dur)
		}
	}

	// Extract bitrate
	if probe.Format.BitRate != "" {
		bitrate, err := strconv.ParseInt(probe.Format.BitRate, 10, 64)
		if err == nil {
			info.Bitrate = fmt.Sprintf("%d kbps", bitrate/1000)
		}
	}

	// Extract resolution and codec from the first video stream
	for _, stream := range probe.Streams {
		if stream.Width > 0 && stream.Height > 0 {
			info.Resolution = fmt.Sprintf("%dx%d", stream.Width, stream.Height)
			info.Codec = stream.CodecName
			break
		}
	}

	// If no video stream found, try to get codec from first audio stream
	if info.Codec == "" && len(probe.Streams) > 0 {
		info.Codec = probe.Streams[0].CodecName
		// Try to get duration from stream if format didn't have it
		if info.DurationSec == 0 && probe.Streams[0].Duration != "" {
			if dur, err := strconv.ParseFloat(probe.Streams[0].Duration, 64); err == nil {
				info.DurationSec = uint(dur)
			}
		}
	}

	LogToFile(fmt.Sprintf("Probed media info | File: %s | Duration: %ds | Resolution: %s | Codec: %s",
		filePath, info.DurationSec, info.Resolution, info.Codec))

	return info, nil
}

// GetMediaType returns a simplified type string from a MIME type.
func GetMediaType(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	default:
		return "document"
	}
}
