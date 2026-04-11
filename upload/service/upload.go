package service

import "mediahub/upload/repository"

type UploadService struct {
	repo *repository.UploadRepository
}

func NewUploadService(repo *repository.UploadRepository) *UploadService {
	return &UploadService{repo: repo}
}
