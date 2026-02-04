package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
)

// UserService 用户服务接口
type UserService interface {
	Register(username, password, email, nickname string) (*model.User, error)
	Login(username, password string) (*model.User, error)
	GetUserByID(id uint) (*model.User, error)
	UpdateUser(user *model.User) error
	ChangePassword(userID uint, newPassword string) error
	EnsureDefaultAdmin() error
	CheckIn(userID uint) error
	ListUsers(page, size int) ([]*model.User, int64, error)
	UpdateAIAccess(userID uint, durationDays int, code string) (*model.User, error)
	UpdateHeartbeat(userID uint, deviceID string) (*model.User, error)
}

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// Register 用户注册
func (s *userService) Register(username, password, email, nickname string) (*model.User, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if email != "" {
		if _, err := s.userRepo.FindByEmail(email); err == nil {
			return nil, errors.New("email already exists")
		}
	}

	if nickname == "" {
		nickname = "用户-" + uuid.New().String()[:8]
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
		Nickname: nickname,
		Role:     "user",
		Status:   model.StatusEnabled,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *userService) Login(username, password string) (*model.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 检查用户状态
	if user.Status != model.StatusEnabled {
		return nil, errors.New("account is disabled")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(user *model.User) error {
	return s.userRepo.Update(user)
}

// ChangePassword 修改用户密码
func (s *userService) ChangePassword(userID uint, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if user.Status != model.StatusEnabled {
		return errors.New("account is disabled")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.MustChangePassword = false
	return s.userRepo.Update(user)
}

// EnsureDefaultAdmin 确保默认管理员账号存在
func (s *userService) EnsureDefaultAdmin() error {
	const defaultUsername = "admin"
	const defaultPassword = "admin"

	if _, err := s.userRepo.FindByUsername(defaultUsername); err == nil {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:           defaultUsername,
		Password:           string(hashedPassword),
		Nickname:           "管理员",
		Role:               "admin",
		Status:             model.StatusEnabled,
		MustChangePassword: true,
	}

	return s.userRepo.Create(user)
}

// CheckIn 每日签到
func (s *userService) CheckIn(userID uint) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 检查今天是否已经签到
	if user.LastCheckIn != nil {
		lastCheckIn := time.Date(user.LastCheckIn.Year(), user.LastCheckIn.Month(), user.LastCheckIn.Day(), 0, 0, 0, 0, user.LastCheckIn.Location())
		if lastCheckIn.Equal(today) {
			return errors.New("already checked in today")
		}

		// 检查是否连续签到
		yesterday := today.AddDate(0, 0, -1)
		if lastCheckIn.Equal(yesterday) {
			user.CheckInStreak++
		} else {
			user.CheckInStreak = 1
		}
	} else {
		user.CheckInStreak = 1
	}

	user.LastCheckIn = &now
	user.Points += 10 // 签到获得10积分

	return s.userRepo.Update(user)
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(page, size int) ([]*model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	return s.userRepo.List(page, size)
}

// UpdateAIAccess 更新用户AI权限有效期
func (s *userService) UpdateAIAccess(userID uint, durationDays int, code string) (*model.User, error) {
	if durationDays <= 0 {
		return nil, errors.New("invalid duration")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	base := time.Now()
	if user.AIAccessUntil != nil && user.AIAccessUntil.After(base) {
		base = *user.AIAccessUntil
	}

	newUntil := base.AddDate(0, 0, durationDays)
	user.AIAccessUntil = &newUntil
	user.InviteCodeUsed = code

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateHeartbeat 更新用户心跳信息
func (s *userService) UpdateHeartbeat(userID uint, deviceID string) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if deviceID != "" {
		user.LastDeviceID = deviceID
	}
	lastActive := time.Now()
	user.LastActiveAt = &lastActive

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}
