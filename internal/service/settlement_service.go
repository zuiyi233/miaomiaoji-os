package service

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
)

type SettlementService interface {
	CreateEntry(entry *model.SettlementEntry) error
	GetEntry(id uint) (*model.SettlementEntry, error)
	UpdateEntry(entry *model.SettlementEntry) error
	DeleteEntry(id uint) error
	ListEntries(userID uint, page, pageSize int) ([]*model.SettlementEntry, int64, error)
	ListEntriesByWorld(worldID string, page, pageSize int) ([]*model.SettlementEntry, int64, error)
	ListEntriesByChapter(chapterID string, page, pageSize int) ([]*model.SettlementEntry, int64, error)
	ListEntriesByLoopStage(loopStage string, page, pageSize int) ([]*model.SettlementEntry, int64, error)
	FilterEntries(worldID, chapterID, loopStage string, page, pageSize int) ([]*model.SettlementEntry, int64, error)
	GetTotalPoints(userID uint) (int, error)
}

type settlementService struct {
	settlementRepo repository.SettlementRepository
}

func NewSettlementService(settlementRepo repository.SettlementRepository) SettlementService {
	return &settlementService{
		settlementRepo: settlementRepo,
	}
}

func (s *settlementService) CreateEntry(entry *model.SettlementEntry) error {
	return s.settlementRepo.Create(entry)
}

func (s *settlementService) GetEntry(id uint) (*model.SettlementEntry, error) {
	return s.settlementRepo.GetByID(id)
}

func (s *settlementService) UpdateEntry(entry *model.SettlementEntry) error {
	return s.settlementRepo.Update(entry)
}

func (s *settlementService) DeleteEntry(id uint) error {
	return s.settlementRepo.Delete(id)
}

func (s *settlementService) ListEntries(userID uint, page, pageSize int) ([]*model.SettlementEntry, int64, error) {
	return s.settlementRepo.ListByUserID(userID, page, pageSize)
}

func (s *settlementService) ListEntriesByWorld(worldID string, page, pageSize int) ([]*model.SettlementEntry, int64, error) {
	return s.settlementRepo.ListByWorldID(worldID, page, pageSize)
}

func (s *settlementService) ListEntriesByChapter(chapterID string, page, pageSize int) ([]*model.SettlementEntry, int64, error) {
	return s.settlementRepo.ListByChapterID(chapterID, page, pageSize)
}

func (s *settlementService) ListEntriesByLoopStage(loopStage string, page, pageSize int) ([]*model.SettlementEntry, int64, error) {
	return s.settlementRepo.ListByLoopStage(loopStage, page, pageSize)
}

func (s *settlementService) FilterEntries(worldID, chapterID, loopStage string, page, pageSize int) ([]*model.SettlementEntry, int64, error) {
	return s.settlementRepo.Filter(worldID, chapterID, loopStage, page, pageSize)
}

func (s *settlementService) GetTotalPoints(userID uint) (int, error) {
	return s.settlementRepo.GetTotalPointsByUserID(userID)
}
