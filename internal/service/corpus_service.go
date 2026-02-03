package service

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
)

type CorpusService interface {
	CreateStory(story *model.CorpusStory) error
	GetStory(id uint) (*model.CorpusStory, error)
	UpdateStory(story *model.CorpusStory) error
	DeleteStory(id uint) error
	ListStories(page, pageSize int) ([]*model.CorpusStory, int64, error)
	ListStoriesByGenre(genre string, page, pageSize int) ([]*model.CorpusStory, int64, error)
	SearchStories(keyword string, page, pageSize int) ([]*model.CorpusStory, int64, error)
}

type corpusService struct {
	corpusRepo repository.CorpusRepository
}

func NewCorpusService(corpusRepo repository.CorpusRepository) CorpusService {
	return &corpusService{
		corpusRepo: corpusRepo,
	}
}

func (s *corpusService) CreateStory(story *model.CorpusStory) error {
	return s.corpusRepo.Create(story)
}

func (s *corpusService) GetStory(id uint) (*model.CorpusStory, error) {
	return s.corpusRepo.GetByID(id)
}

func (s *corpusService) UpdateStory(story *model.CorpusStory) error {
	return s.corpusRepo.Update(story)
}

func (s *corpusService) DeleteStory(id uint) error {
	return s.corpusRepo.Delete(id)
}

func (s *corpusService) ListStories(page, pageSize int) ([]*model.CorpusStory, int64, error) {
	return s.corpusRepo.List(page, pageSize)
}

func (s *corpusService) ListStoriesByGenre(genre string, page, pageSize int) ([]*model.CorpusStory, int64, error) {
	return s.corpusRepo.ListByGenre(genre, page, pageSize)
}

func (s *corpusService) SearchStories(keyword string, page, pageSize int) ([]*model.CorpusStory, int64, error) {
	return s.corpusRepo.Search(keyword, page, pageSize)
}
