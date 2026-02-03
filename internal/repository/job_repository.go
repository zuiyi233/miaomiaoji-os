package repository

import (
	"novel-agent-os-backend/internal/model"

	"gorm.io/gorm"
)

type JobRepository interface {
	Create(job *model.Job) error
	GetByUUID(jobUUID string) (*model.Job, error)
	Update(job *model.Job) error
}

type jobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) Create(job *model.Job) error {
	return r.db.Create(job).Error
}

func (r *jobRepository) GetByUUID(jobUUID string) (*model.Job, error) {
	var job model.Job
	if err := r.db.Where("job_uuid = ?", jobUUID).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *jobRepository) Update(job *model.Job) error {
	return r.db.Save(job).Error
}
