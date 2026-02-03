package repository

import (
	"gorm.io/gorm"
	"novel-agent-os-backend/internal/model"
)

type CorpusRepository interface {
	Create(story *model.CorpusStory) error
	GetByID(id uint) (*model.CorpusStory, error)
	Update(story *model.CorpusStory) error
	Delete(id uint) error
	List(page, pageSize int) ([]*model.CorpusStory, int64, error)
	ListByGenre(genre string, page, pageSize int) ([]*model.CorpusStory, int64, error)
	Search(keyword string, page, pageSize int) ([]*model.CorpusStory, int64, error)
}

type corpusRepository struct {
	db *gorm.DB
}

func NewCorpusRepository(db *gorm.DB) CorpusRepository {
	return &corpusRepository{db: db}
}

func (r *corpusRepository) Create(story *model.CorpusStory) error {
	return r.db.Create(story).Error
}

func (r *corpusRepository) GetByID(id uint) (*model.CorpusStory, error) {
	var story model.CorpusStory
	err := r.db.First(&story, id).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

func (r *corpusRepository) Update(story *model.CorpusStory) error {
	return r.db.Save(story).Error
}

func (r *corpusRepository) Delete(id uint) error {
	return r.db.Delete(&model.CorpusStory{}, id).Error
}

func (r *corpusRepository) List(page, pageSize int) ([]*model.CorpusStory, int64, error) {
	var stories []*model.CorpusStory
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.CorpusStory{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&stories).Error

	return stories, total, err
}

func (r *corpusRepository) ListByGenre(genre string, page, pageSize int) ([]*model.CorpusStory, int64, error) {
	var stories []*model.CorpusStory
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.CorpusStory{}).Where("genre = ?", genre).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("genre = ?", genre).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&stories).Error

	return stories, total, err
}

func (r *corpusRepository) Search(keyword string, page, pageSize int) ([]*model.CorpusStory, int64, error) {
	var stories []*model.CorpusStory
	var total int64

	offset := (page - 1) * pageSize

	searchPattern := "%" + keyword + "%"

	err := r.db.Model(&model.CorpusStory{}).
		Where("title LIKE ? OR genre LIKE ?", searchPattern, searchPattern).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("title LIKE ? OR genre LIKE ?", searchPattern, searchPattern).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&stories).Error

	return stories, total, err
}
