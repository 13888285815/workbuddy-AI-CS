package repository

import (
	"github.com/yndxw/workbuddy-ai/backend/models"
	"gorm.io/gorm"
)

// WidgetOpenRepository 访客小窗打开埋点
type WidgetOpenRepository struct {
	db *gorm.DB
}

func NewWidgetOpenRepository(db *gorm.DB) *WidgetOpenRepository {
	return &WidgetOpenRepository{db: db}
}

func (r *WidgetOpenRepository) Create(e *models.WidgetOpenEvent) error {
	return r.db.Create(e).Error
}
