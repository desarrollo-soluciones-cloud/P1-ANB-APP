package storage

import (
	"io"
	"mime/multipart"
	"os"
)

type StorageService interface {
	Upload(file multipart.File, destinationPath string) error
}

type localStorageService struct{}

func NewLocalStorageService() StorageService {
	return &localStorageService{}
}

func (s *localStorageService) Upload(file multipart.File, destinationPath string) error {
	dst, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return err
	}

	return nil
}
