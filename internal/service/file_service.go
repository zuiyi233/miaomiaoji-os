package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/internal/storage"
)

type FileService interface {
	CreateFile(file *model.File, reader io.Reader) error
	GetFile(id uint) (*model.File, error)
	GetFileByStorageKey(key string) (*model.File, error)
	UpdateFile(file *model.File) error
	DeleteFile(id uint) error
	ListFiles(userID uint, page, pageSize int) ([]*model.File, int64, error)
	ListFilesByProject(projectID uint, page, pageSize int) ([]*model.File, int64, error)
	ListFilesByType(fileType string, page, pageSize int) ([]*model.File, int64, error)
	DownloadFile(id uint) (io.ReadCloser, error)
	GetLatestBackupByProject(projectID uint) (*model.File, error)
}

type fileService struct {
	fileRepo repository.FileRepository
	storage  storage.Storage
}

func NewFileService(fileRepo repository.FileRepository, storage storage.Storage) FileService {
	return &fileService{
		fileRepo: fileRepo,
		storage:  storage,
	}
}

func (s *fileService) CreateFile(file *model.File, reader io.Reader) error {
	hash := sha256.New()
	tee := io.TeeReader(reader, hash)

	storageKey := file.StorageKey
	if storageKey == "" {
		storageKey = generateStorageKey(file.FileName, file.UserID)
		file.StorageKey = storageKey
	}

	if err := s.storage.Put(storageKey, tee, file.ContentType); err != nil {
		return err
	}

	file.SHA256 = hex.EncodeToString(hash.Sum(nil))

	return s.fileRepo.Create(file)
}

func (s *fileService) GetFile(id uint) (*model.File, error) {
	return s.fileRepo.GetByID(id)
}

func (s *fileService) GetFileByStorageKey(key string) (*model.File, error) {
	return s.fileRepo.GetByStorageKey(key)
}

func (s *fileService) UpdateFile(file *model.File) error {
	return s.fileRepo.Update(file)
}

func (s *fileService) DeleteFile(id uint) error {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.storage.Delete(file.StorageKey); err != nil {
		return err
	}

	return s.fileRepo.Delete(id)
}

func (s *fileService) ListFiles(userID uint, page, pageSize int) ([]*model.File, int64, error) {
	return s.fileRepo.ListByUserID(userID, page, pageSize)
}

func (s *fileService) ListFilesByProject(projectID uint, page, pageSize int) ([]*model.File, int64, error) {
	return s.fileRepo.ListByProjectID(projectID, page, pageSize)
}

func (s *fileService) ListFilesByType(fileType string, page, pageSize int) ([]*model.File, int64, error) {
	return s.fileRepo.ListByType(fileType, page, pageSize)
}

func (s *fileService) GetLatestBackupByProject(projectID uint) (*model.File, error) {
	return s.fileRepo.GetLatestByProjectAndType(projectID, "backup")
}

func (s *fileService) DownloadFile(id uint) (io.ReadCloser, error) {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.storage.Open(file.StorageKey)
}

func generateStorageKey(fileName string, userID uint) string {
	return "files/" + fmt.Sprintf("%d", userID) + "/" + fileName
}
