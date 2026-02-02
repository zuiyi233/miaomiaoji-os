package repository

import (
	"novel-agent-os-backend/internal/model"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
	List(page, size int) ([]*model.User, int64, error)
}

// userRepository 用户仓储实现
type userRepository struct {
	*BaseRepository
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository() UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository(),
	}
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.DB.Create(user).Error
}

// FindByID 根据ID查找用户
func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(user *model.User) error {
	return r.DB.Save(user).Error
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(id uint) error {
	return r.DB.Delete(&model.User{}, id).Error
}

// List 获取用户列表
func (r *userRepository) List(page, size int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	if err := r.DB.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := r.DB.Offset(offset).Limit(size).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
