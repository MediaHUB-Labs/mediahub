package service

import "mediahub/media/repository"

type MediaService struct {
	repo *repository.MediaRepository
}

func NewMediaService(repo *repository.MediaRepository) *MediaService {
	return &MediaService{repo: repo}
}
