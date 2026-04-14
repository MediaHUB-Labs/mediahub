package transcode

import (
	"bufio"
	"context"
	"fmt"
	"mediahub/media/repository"
	"mediahub/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// TranscodeService handles CPU-only video transcoding to HLS format.
type TranscodeService struct {
	mediaRepo    *repository.MediaRepository
	basePath     string
	maxJobs      int
	activeJobs   int
	JobProgress  map[uint]float64 // Tracks progress (0-100) per mediaID
	mu           sync.Mutex
	ffmpegPreset string
	ffmpegCRF    string
}

// NewTranscodeService creates a new transcoding service.
func NewTranscodeService(mediaRepo *repository.MediaRepository, basePath string) *TranscodeService {
	maxJobs := 1
	if envJobs := os.Getenv("MAX_TRANSCODE_JOBS"); envJobs != "" {
		if parsed, err := strconv.Atoi(envJobs); err == nil && parsed > 0 {
			maxJobs = parsed
		}
	}

	preset := os.Getenv("FFMPEG_PRESET")
	if preset == "" {
		preset = "veryfast"
	}

	crf := os.Getenv("FFMPEG_CRF")
	if crf == "" {
		crf = "23"
	}

	return &TranscodeService{
		mediaRepo:    mediaRepo,
		basePath:     basePath,
		maxJobs:      maxJobs,
		JobProgress:  make(map[uint]float64),
		ffmpegPreset: preset,
		ffmpegCRF:    crf,
	}
}

// IsAvailable checks if FFmpeg is installed for transcoding.
func (s *TranscodeService) IsAvailable() bool {
	return utils.IsFFmpegAvailable()
}

// CanAcceptJob checks if there's capacity for another transcode job.
func (s *TranscodeService) CanAcceptJob() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeJobs < s.maxJobs
}

// TranscodeToHLS transcodes a video file to HLS format (CPU-only, using libx264).
// Runs in a background goroutine and updates the media record when complete.
func (s *TranscodeService) TranscodeToHLS(mediaID uint) error {
	if !s.IsAvailable() {
		return fmt.Errorf("FFmpeg is not installed — transcoding unavailable")
	}

	if !s.CanAcceptJob() {
		return fmt.Errorf("maximum concurrent transcode jobs reached (%d)", s.maxJobs)
	}

	// Get media record
	media, err := s.mediaRepo.FindByID(context.Background(), mediaID)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	if media.IsTranscoded {
		return fmt.Errorf("media is already transcoded")
	}

	// Increment active jobs and init progress
	s.mu.Lock()
	s.activeJobs++
	s.JobProgress[mediaID] = 0
	s.mu.Unlock()

	// Run transcoding in background
	go func() {
		defer func() {
			s.mu.Lock()
			s.activeJobs--
			delete(s.JobProgress, mediaID)
			s.mu.Unlock()
		}()

		utils.LogToFile(fmt.Sprintf("Starting HLS transcode for media ID %d: %s", mediaID, media.FilePath))

		// Create output directory: uploads/transcoded/<media_id>/
		outputDir := filepath.Join(s.basePath, "transcoded", fmt.Sprintf("%d", mediaID))
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			utils.LogToFile(fmt.Sprintf("Failed to create transcode output dir: %v", err))
			return
		}

		manifestPath := filepath.Join(outputDir, "index.m3u8")
		segmentPattern := filepath.Join(outputDir, "segment_%03d.ts")

		// FFmpeg command for HLS transcoding (CPU-only) with progress reporting
		cmd := exec.Command("ffmpeg",
			"-i", media.FilePath,
			"-progress", "pipe:1",
			"-c:v", "libx264",
			"-preset", s.ffmpegPreset,
			"-crf", s.ffmpegCRF,
			"-c:a", "aac",
			"-b:a", "128k",
			"-vf", "scale=-2:720",
			"-hls_time", "10",
			"-hls_list_size", "0",
			"-hls_segment_filename", segmentPattern,
			"-f", "hls",
			"-y",
			manifestPath,
		)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			utils.LogToFile(fmt.Sprintf("Failed to get stdout pipe for ffmpeg: %v", err))
			return
		}

		if err := cmd.Start(); err != nil {
			utils.LogToFile(fmt.Sprintf("Failed to start ffmpeg: %v", err))
			return
		}

		// Read progress from stdout
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "out_time_us=") {
				timeUsStr := strings.TrimPrefix(line, "out_time_us=")
				timeUs, err := strconv.ParseInt(timeUsStr, 10, 64)
				if err == nil && media.DurationSec > 0 {
					progress := (float64(timeUs) / 1000000.0) / float64(media.DurationSec) * 100
					if progress > 100 {
						progress = 100
					}
					s.mu.Lock()
					s.JobProgress[mediaID] = progress
					s.mu.Unlock()
				}
			}
		}

		if err := cmd.Wait(); err != nil {
			utils.LogToFile(fmt.Sprintf("HLS transcode failed for media ID %d: %v", mediaID, err))
			// Clean up failed output
			_ = os.RemoveAll(outputDir)
			return
		}

		// Update media record
		media.IsTranscoded = true
		media.TranscodedPath = manifestPath
		if err := s.mediaRepo.Update(context.Background(), media); err != nil {
			utils.LogToFile(fmt.Sprintf("Failed to update media record after transcode: %v", err))
			return
		}

		utils.LogToFile(fmt.Sprintf("HLS transcode completed for media ID %d → %s", mediaID, manifestPath))
	}()

	return nil
}

// GetTranscodeStatus returns the current transcoding status.
type TranscodeStatus struct {
	Available   bool             `json:"available"`
	ActiveJobs  int              `json:"active_jobs"`
	MaxJobs     int              `json:"max_jobs"`
	JobProgress map[uint]float64 `json:"job_progress"`
}

func (s *TranscodeService) GetStatus() TranscodeStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a copy of the progress map to avoid concurrent access issues after returning
	progressCopy := make(map[uint]float64)
	for k, v := range s.JobProgress {
		progressCopy[k] = v
	}

	return TranscodeStatus{
		Available:   s.IsAvailable(),
		ActiveJobs:  s.activeJobs,
		MaxJobs:     s.maxJobs,
		JobProgress: progressCopy,
	}
}
