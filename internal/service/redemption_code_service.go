package service

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"time"

	"gorm.io/datatypes"
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
)

// RedemptionCodeService 兑换码服务接口
type RedemptionCodeService interface {
	Redeem(userID uint, deviceID string, code string) (*model.RedemptionCode, int, error)
	Generate(payload GenerateCodesPayload, createdBy uint) ([]*model.RedemptionCode, error)
	List(filter repository.RedemptionCodeFilter) ([]*model.RedemptionCode, int64, error)
	UpdateStatus(codes []string, action string, value int) error
}

// redemptionCodeService 兑换码服务实现
type redemptionCodeService struct {
	codeRepo repository.RedemptionCodeRepository
}

// GenerateCodesPayload 生成兑换码参数
type GenerateCodesPayload struct {
	Prefix       string
	Length       int
	Count        int
	ValidityDays int
	MaxUses      int
	CharType     string
	Tags         []string
	Note         string
	Source       string
}

// NewRedemptionCodeService 创建兑换码服务实例
func NewRedemptionCodeService(codeRepo repository.RedemptionCodeRepository) RedemptionCodeService {
	return &redemptionCodeService{
		codeRepo: codeRepo,
	}
}

// Redeem 兑换码校验并记录使用
func (s *redemptionCodeService) Redeem(userID uint, deviceID string, code string) (*model.RedemptionCode, int, error) {
	cleanCode := strings.TrimSpace(strings.ToUpper(code))
	if cleanCode == "" {
		return nil, 0, errors.New("empty code")
	}

	item, err := s.codeRepo.FindByCode(cleanCode)
	if err != nil {
		return nil, 0, errors.New("code not found")
	}

	if item.Status != "active" {
		return nil, 0, errors.New("code disabled")
	}

	if item.ExpiresAt != nil && item.ExpiresAt.Before(time.Now()) {
		return nil, 0, errors.New("code expired")
	}

	if item.MaxUses > 0 && item.UsedCount >= item.MaxUses {
		return nil, 0, errors.New("code depleted")
	}

	if userID > 0 {
		used, err := s.codeRepo.CountUserUses(item.ID, userID)
		if err != nil {
			return nil, 0, err
		}
		if used > 0 {
			return nil, 0, errors.New("code already used")
		}
	}

	useRecord := &model.RedemptionCodeUse{
		CodeID:   item.ID,
		UserID:   userID,
		DeviceID: deviceID,
	}
	if err := s.codeRepo.RecordUse(useRecord); err != nil {
		return nil, 0, err
	}

	item.UsedCount++
	if item.MaxUses > 0 && item.UsedCount >= item.MaxUses {
		item.Status = "depleted"
	}

	if err := s.codeRepo.Update(item); err != nil {
		return nil, 0, err
	}

	return item, item.DurationDays, nil
}

// Generate 批量生成兑换码
func (s *redemptionCodeService) Generate(payload GenerateCodesPayload, createdBy uint) ([]*model.RedemptionCode, error) {
	if payload.Count <= 0 || payload.Length <= 0 {
		return nil, errors.New("invalid config")
	}

	if payload.MaxUses <= 0 {
		payload.MaxUses = 1
	}

	if payload.ValidityDays <= 0 {
		payload.ValidityDays = 30
	}

	batchID := time.Now().Format("20060102150405")
	result := make([]*model.RedemptionCode, 0, payload.Count)

	for i := 0; i < payload.Count; i++ {
		codeValue := buildCode(payload.Prefix, payload.Length, payload.CharType)
		expiresAt := time.Now().AddDate(0, 0, payload.ValidityDays)
		tagsJSON, _ := json.Marshal(payload.Tags)

		item := &model.RedemptionCode{
			Code:         codeValue,
			Status:       "active",
			ExpiresAt:    &expiresAt,
			MaxUses:      payload.MaxUses,
			UsedCount:    0,
			DurationDays: payload.ValidityDays,
			CreatedBy:    createdBy,
			BatchID:      batchID,
			Prefix:       payload.Prefix,
			Tags:         datatypes.JSON(tagsJSON),
			Note:         payload.Note,
			Source:       payload.Source,
		}

		if err := s.codeRepo.Create(item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

// List 获取兑换码列表
func (s *redemptionCodeService) List(filter repository.RedemptionCodeFilter) ([]*model.RedemptionCode, int64, error) {
	return s.codeRepo.List(filter)
}

// UpdateStatus 批量操作
func (s *redemptionCodeService) UpdateStatus(codes []string, action string, value int) error {
	switch action {
	case "disable":
		return s.codeRepo.UpdateStatus(codes, "disabled")
	case "enable":
		return s.codeRepo.UpdateStatus(codes, "active")
	case "delete":
		return s.codeRepo.DeleteByCodes(codes)
	case "renew":
		return s.codeRepo.RenewByCodes(codes, value)
	default:
		return errors.New("invalid action")
	}
}

func buildCode(prefix string, length int, charType string) string {
	letters := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	if charType == "alpha" {
		letters = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	} else if charType == "num" {
		letters = "23456789"
	}
	if length < 4 {
		length = 4
	}
	if length > 32 {
		length = 32
	}

	var b strings.Builder
	if prefix != "" {
		b.WriteString(prefix)
	}
	for i := 0; i < length; i++ {
		idx := rand.Intn(len(letters))
		b.WriteByte(letters[idx])
	}
	return b.String()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
