package repository

import (
	"encoding/json"
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/pkg/database"

	"gorm.io/datatypes"
)

// DocumentRepository 文档数据访问接口
type DocumentRepository interface {
	Create(document *model.Document) error
	FindByID(id uint) (*model.Document, error)
	FindByProjectID(projectID uint, page, size int) ([]*model.Document, int64, error)
	FindByVolumeID(volumeID uint, page, size int) ([]*model.Document, int64, error)
	Update(document *model.Document) error
	Delete(id uint) error
	AddBookmark(documentID uint, bookmark model.Bookmark) error
	RemoveBookmark(documentID uint, index int) error
	LinkEntity(documentID, entityID uint, refType string, metadata map[string]interface{}) error
	UnlinkEntity(documentID, entityID uint) error
	GetEntityRefs(documentID uint) ([]*model.DocumentEntityRef, error)
}

// documentRepository 文档数据访问实现
type documentRepository struct{}

// NewDocumentRepository 创建文档仓库实例
func NewDocumentRepository() DocumentRepository {
	return &documentRepository{}
}

// Create 创建文档
func (r *documentRepository) Create(document *model.Document) error {
	return database.GetDB().Create(document).Error
}

// FindByID 根据ID查找文档
func (r *documentRepository) FindByID(id uint) (*model.Document, error) {
	var document model.Document
	err := database.GetDB().First(&document, id).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

// FindByProjectID 根据项目ID查找文档列表
func (r *documentRepository) FindByProjectID(projectID uint, page, size int) ([]*model.Document, int64, error) {
	var documents []*model.Document
	var total int64

	db := database.GetDB().Model(&model.Document{}).Where("project_id = ?", projectID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("order_index ASC, id ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&documents).Error; err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

// FindByVolumeID 根据卷ID查找文档列表
func (r *documentRepository) FindByVolumeID(volumeID uint, page, size int) ([]*model.Document, int64, error) {
	var documents []*model.Document
	var total int64

	db := database.GetDB().Model(&model.Document{}).Where("volume_id = ?", volumeID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("order_index ASC, id ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&documents).Error; err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

// Update 更新文档
func (r *documentRepository) Update(document *model.Document) error {
	return database.GetDB().Save(document).Error
}

// Delete 删除文档（软删除）
func (r *documentRepository) Delete(id uint) error {
	return database.GetDB().Delete(&model.Document{}, id).Error
}

// AddBookmark 添加书签
func (r *documentRepository) AddBookmark(documentID uint, bookmark model.Bookmark) error {
	var document model.Document
	if err := database.GetDB().First(&document, documentID).Error; err != nil {
		return err
	}

	var bookmarks []model.Bookmark
	if document.Bookmarks != nil && len(document.Bookmarks) > 0 {
		if err := json.Unmarshal(document.Bookmarks, &bookmarks); err != nil {
			bookmarks = []model.Bookmark{}
		}
	}

	bookmarks = append(bookmarks, model.Bookmark{
		Title:    bookmark.Title,
		Position: bookmark.Position,
		Note:     bookmark.Note,
	})

	bookmarksJSON, err := json.Marshal(bookmarks)
	if err != nil {
		return err
	}

	document.Bookmarks = datatypes.JSON(bookmarksJSON)
	return database.GetDB().Save(&document).Error
}

// RemoveBookmark 移除书签
func (r *documentRepository) RemoveBookmark(documentID uint, index int) error {
	var document model.Document
	if err := database.GetDB().First(&document, documentID).Error; err != nil {
		return err
	}

	var bookmarks []model.Bookmark
	if document.Bookmarks != nil && len(document.Bookmarks) > 0 {
		if err := json.Unmarshal(document.Bookmarks, &bookmarks); err != nil {
			return err
		}
	}

	if index < 0 || index >= len(bookmarks) {
		return nil // 索引无效，不做操作
	}

	bookmarks = append(bookmarks[:index], bookmarks[index+1:]...)

	bookmarksJSON, err := json.Marshal(bookmarks)
	if err != nil {
		return err
	}

	document.Bookmarks = datatypes.JSON(bookmarksJSON)
	return database.GetDB().Save(&document).Error
}

// LinkEntity 关联实体
func (r *documentRepository) LinkEntity(documentID, entityID uint, refType string, metadata map[string]interface{}) error {
	metadataJSON, _ := json.Marshal(metadata)
	ref := &model.DocumentEntityRef{
		DocumentID: documentID,
		EntityID:   entityID,
		RefType:    refType,
		Metadata:   datatypes.JSON(metadataJSON),
	}
	return database.GetDB().Create(ref).Error
}

// UnlinkEntity 取消关联实体
func (r *documentRepository) UnlinkEntity(documentID, entityID uint) error {
	return database.GetDB().
		Where("document_id = ? AND entity_id = ?", documentID, entityID).
		Delete(&model.DocumentEntityRef{}).Error
}

// GetEntityRefs 获取文档的实体关联
func (r *documentRepository) GetEntityRefs(documentID uint) ([]*model.DocumentEntityRef, error) {
	var refs []*model.DocumentEntityRef
	err := database.GetDB().
		Where("document_id = ?", documentID).
		Preload("Entity").
		Find(&refs).Error
	return refs, err
}
