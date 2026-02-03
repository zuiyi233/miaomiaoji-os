package service

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"rulebacktest/internal/config"
	"rulebacktest/internal/model"
	"rulebacktest/internal/repository"
	apperrors "rulebacktest/pkg/errors"
	"rulebacktest/pkg/logger"
)

var (
	userServiceInstance *UserService
	userServiceOnce     sync.Once
)

// UserService 用户业务逻辑层
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService 创建UserService实例
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetUserService 获取UserService单例
func GetUserService() *UserService {
	userServiceOnce.Do(func() {
		userServiceInstance = &UserService{
			repo: repository.GetUserRepository(),
		}
	})
	return userServiceInstance
}

// Register 用户注册
func (s *UserService) Register(req *model.UserRegisterReq) (*model.User, error) {
	exists, err := s.repo.ExistsByUsername(req.Username)
	if err != nil {
		logger.Error("检查用户名失败", logger.Err(err))
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	if exists {
		return nil, apperrors.ErrUserExists
	}

	if req.Email != "" {
		exists, err = s.repo.ExistsByEmail(req.Email)
		if err != nil {
			logger.Error("检查邮箱失败", logger.Err(err))
			return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
		}
		if exists {
			return nil, apperrors.New(apperrors.CodeUserExists, "邮箱已被使用")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("密码加密失败", logger.Err(err))
		return nil, apperrors.WrapWithCode(apperrors.CodeInternalError, err)
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   model.StatusEnabled,
	}

	if err := s.repo.Create(user); err != nil {
		logger.Error("创建用户失败", logger.Err(err))
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(req *model.UserLoginReq) (*model.UserLoginResp, error) {
	user, err := s.repo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		logger.Error("查询用户失败", logger.Err(err))
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if !user.Status.IsEnabled() {
		return nil, apperrors.NewWithCode(apperrors.CodeUserDisabled)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperrors.ErrPasswordIncorrect
	}

	token, err := s.generateToken(user.ID, user.Role)
	if err != nil {
		logger.Error("生成Token失败", logger.Err(err))
		return nil, apperrors.WrapWithCode(apperrors.CodeInternalError, err)
	}

	return &model.UserLoginResp{
		Token: token,
		User:  user,
	}, nil
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(id uint) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return user, nil
}

// Update 更新用户信息
func (s *UserService) Update(id uint, req *model.UserUpdateReq) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if req.Email != "" && req.Email != user.Email {
		exists, err := s.repo.ExistsByEmail(req.Email)
		if err != nil {
			return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
		}
		if exists {
			return nil, apperrors.New(apperrors.CodeUserExists, "邮箱已被使用")
		}
		user.Email = req.Email
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if err := s.repo.Update(user); err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return user, nil
}

// generateToken 生成JWT Token
func (s *UserService) generateToken(userID uint, role string) (string, error) {
	cfg := config.Get()
	secret := cfg.JWT.Secret
	expireHours := cfg.JWT.ExpireHours

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Duration(expireHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// TokenClaims Token解析结果
type TokenClaims struct {
	UserID uint
	Role   string
}

// ParseToken 解析JWT Token
func (s *UserService) ParseToken(tokenString string) (*TokenClaims, error) {
	cfg := config.Get()
	secret := cfg.JWT.Secret

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.ErrTokenInvalid
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrTokenExpired
		}
		return nil, apperrors.ErrTokenInvalid
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint(claims["user_id"].(float64))
		role := ""
		if r, ok := claims["role"].(string); ok {
			role = r
		}
		return &TokenClaims{
			UserID: userID,
			Role:   role,
		}, nil
	}

	return nil, apperrors.ErrTokenInvalid
}
