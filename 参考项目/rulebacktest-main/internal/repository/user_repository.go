package repository

import (
	"sync"

	"gorm.io/gorm"
	"rulebacktest/internal/model"
	"rulebacktest/pkg/database"
)

var (
	userRepoInstance *UserRepository
	userRepoOnce     sync.Once
)

// UserRepository 用户数据访问层
type UserRepository struct {
	*BaseRepository
}

// NewUserRepository 创建UserRepository实例
func NewUserRepository(base *BaseRepository) *UserRepository {
	return &UserRepository{BaseRepository: base}
}

// GetUserRepository 获取UserRepository单例
func GetUserRepository() *UserRepository {
	userRepoOnce.Do(func() {
		userRepoInstance = &UserRepository{
			BaseRepository: &BaseRepository{db: database.GetDB()},
		}
	})
	return userRepoInstance
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// List 获取用户列表
func (r *UserRepository) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Scopes(r.Paginate(page, pageSize)).
		Order("id DESC").
		Find(&users).Error

	return users, total, err
}

// UpdatePassword 更新用户密码
func (r *UserRepository) UpdatePassword(userID uint, password string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("password", password).Error
}

// UpdateStatus 更新用户状态
func (r *UserRepository) UpdateStatus(userID uint, status model.Status) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("status", status).Error
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update 更新用户
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete 删除用户
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

// Transaction 事务
func (r *UserRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
