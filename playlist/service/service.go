package service

import (
	"context"
	"errors"
	"mediahub/dto"
	mediarepository "mediahub/media/repository"
	"mediahub/models"
	"mediahub/playlist/repository"
	"strings"
)

type PlaylistService struct {
	repo      *repository.PlaylistRepository
	mediaRepo *mediarepository.MediaRepository
}

func NewPlaylistService(repo *repository.PlaylistRepository, mediaRepo *mediarepository.MediaRepository) *PlaylistService {
	return &PlaylistService{
		repo:      repo,
		mediaRepo: mediaRepo,
	}
}

// CreatePlaylist creates a new playlist for a user.
func (s *PlaylistService) CreatePlaylist(ctx context.Context, userID uint, req *dto.CreatePlaylistRequest) (*dto.PlaylistResponse, error) {
	playlist := &models.Playlist{
		Name:        req.Name,
		Description: req.Description,
		UserID:      userID,
		IsPublic:    req.IsPublic,
	}

	if err := s.repo.Create(ctx, playlist); err != nil {
		return nil, err
	}

	return s.playlistToResponse(ctx, playlist), nil
}

// GetPlaylist returns a playlist by ID. Only the owner or public playlists are accessible.
func (s *PlaylistService) GetPlaylist(ctx context.Context, id uint, userID uint) (*dto.PlaylistResponse, error) {
	playlist, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Access control: only owner or public
	if playlist.UserID != userID && !playlist.IsPublic {
		return nil, errors.New("access denied")
	}

	return s.playlistToResponse(ctx, playlist), nil
}

// ListUserPlaylists returns all playlists for a user.
func (s *PlaylistService) ListUserPlaylists(ctx context.Context, userID uint) ([]dto.PlaylistResponse, error) {
	playlists, err := s.repo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var responses []dto.PlaylistResponse
	for _, p := range playlists {
		responses = append(responses, *s.playlistToResponse(ctx, &p))
	}

	return responses, nil
}

// UpdatePlaylist updates playlist metadata (owner only).
func (s *PlaylistService) UpdatePlaylist(ctx context.Context, id uint, userID uint, req *dto.UpdatePlaylistRequest) (*dto.PlaylistResponse, error) {
	playlist, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if playlist.UserID != userID {
		return nil, errors.New("access denied: only the owner can update this playlist")
	}

	if req.Name != "" {
		playlist.Name = req.Name
	}
	if req.Description != "" {
		playlist.Description = req.Description
	}
	if req.IsPublic != nil {
		playlist.IsPublic = *req.IsPublic
	}

	if err := s.repo.Update(ctx, playlist); err != nil {
		return nil, err
	}

	return s.playlistToResponse(ctx, playlist), nil
}

// DeletePlaylist removes a playlist (owner only).
func (s *PlaylistService) DeletePlaylist(ctx context.Context, id uint, userID uint) error {
	playlist, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if playlist.UserID != userID {
		return errors.New("access denied: only the owner can delete this playlist")
	}

	return s.repo.Delete(ctx, id)
}

// AddItem adds a media item to a playlist. Only audio/music can be added.
func (s *PlaylistService) AddItem(ctx context.Context, playlistID uint, mediaID uint, userID uint) (*dto.PlaylistItemResponse, error) {
	// Check playlist ownership
	playlist, err := s.repo.FindByID(ctx, playlistID)
	if err != nil {
		return nil, err
	}
	if playlist.UserID != userID {
		return nil, errors.New("access denied: only the owner can modify this playlist")
	}

	// Verify media exists and is audio
	media, err := s.mediaRepo.FindByID(ctx, mediaID)
	if err != nil {
		return nil, errors.New("media not found")
	}
	if !strings.HasPrefix(media.MimeType, "audio/") {
		return nil, errors.New("only audio/music can be added to playlists")
	}

	item, err := s.repo.AddItem(ctx, playlistID, mediaID, 0)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, errors.New("this song is already in the playlist")
		}
		return nil, err
	}

	return &dto.PlaylistItemResponse{
		ID:       item.ID,
		MediaID:  item.MediaID,
		Position: item.Position,
		Media: dto.MediaResponse{
			ID:          media.ID,
			Title:       media.Title,
			Description: media.Description,
			Category:    media.Category,
			Genres:      media.Genres,
			MimeType:    media.MimeType,
			DurationSec: media.DurationSec,
			CreatedAt:   media.CreatedAt,
		},
	}, nil
}

// RemoveItem removes a media item from a playlist (owner only).
func (s *PlaylistService) RemoveItem(ctx context.Context, playlistID uint, mediaID uint, userID uint) error {
	// Check playlist ownership
	playlist, err := s.repo.FindByID(ctx, playlistID)
	if err != nil {
		return err
	}
	if playlist.UserID != userID {
		return errors.New("access denied: only the owner can modify this playlist")
	}

	return s.repo.RemoveItem(ctx, playlistID, mediaID)
}

// playlistToResponse converts a Playlist model to a PlaylistResponse DTO.
func (s *PlaylistService) playlistToResponse(ctx context.Context, playlist *models.Playlist) *dto.PlaylistResponse {
	itemCount, _ := s.repo.GetItemCount(ctx, playlist.ID)

	var items []dto.PlaylistItemResponse
	for _, item := range playlist.Items {
		items = append(items, dto.PlaylistItemResponse{
			ID:       item.ID,
			MediaID:  item.MediaID,
			Position: item.Position,
			Media: dto.MediaResponse{
				ID:          item.Media.ID,
				Title:       item.Media.Title,
				Description: item.Media.Description,
				Category:    item.Media.Category,
				Genres:      item.Media.Genres,
				MimeType:    item.Media.MimeType,
				DurationSec: item.Media.DurationSec,
				CreatedAt:   item.Media.CreatedAt,
			},
		})
	}

	return &dto.PlaylistResponse{
		ID:          playlist.ID,
		Name:        playlist.Name,
		Description: playlist.Description,
		IsPublic:    playlist.IsPublic,
		UserID:      playlist.UserID,
		ItemCount:   int(itemCount),
		Items:       items,
		CreatedAt:   playlist.CreatedAt,
		UpdatedAt:   playlist.UpdatedAt,
	}
}
