package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/one-compressive/web-backend-availability/internal/app/ds"
	"gorm.io/gorm"
)

func (r *Repository) GetPublishedComponents(minUptime *float64) ([]ds.Component, error) {
	var components []ds.Component
	query := r.db.Where("status = ?", ds.StatusPublished)
	if minUptime != nil {
		query = query.Where("uptime_percent >= ?", *minUptime)
	}
	if err := query.Order("id ASC").Find(&components).Error; err != nil {
		return nil, err
	}
	return components, nil
}

func (r *Repository) CountLikes(componentID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.ComponentLike{}).Where("component_id = ?", componentID).Count(&count).Error
	return count, err
}

func (r *Repository) GetComponentByID(id uint) (*ds.Component, error) {
	var component ds.Component
	err := r.db.Where("id = ? AND status <> ?", id, ds.StatusDeleted).First(&component).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &component, nil
}

func (r *Repository) GetDraftByCreator(creatorID uint) (*ds.Component, error) {
	var component ds.Component
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&component).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &component, nil
}

func (r *Repository) GetFirstPublishedID() (uint, error) {
	var component ds.Component
	err := r.db.Where("status = ?", ds.StatusPublished).Order("id ASC").First(&component).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return component.ID, nil
}

func (r *Repository) CreateDraft(name string, creatorID uint) (*ds.Component, error) {
	existing, err := r.GetDraftByCreator(creatorID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	component := ds.Component{
		Name:      name,
		Status:    ds.StatusDraft,
		CreatorID: creatorID,
		CreatedAt: time.Now(),
	}
	if err := r.db.Create(&component).Error; err != nil {
		return nil, err
	}
	return &component, nil
}

func (r *Repository) PublishComponent(id uint, shortDescription, description string, configType ds.ConfigType, uptimePercent, systemImpact float32) error {
	now := time.Now()
	result := r.db.Model(&ds.Component{}).
		Where("id = ? AND status = ?", id, ds.StatusDraft).
		Updates(map[string]interface{}{
			"short_description": shortDescription,
			"description":       description,
			"config_type":       configType,
			"uptime_percent":    uptimePercent,
			"system_impact":     systemImpact,
			"status":            ds.StatusPublished,
			"formed_at":         now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("черновик не найден")
	}
	return nil
}

func (r *Repository) DeleteComponent(id uint) error {
	query := "UPDATE components SET status = $1 WHERE id = $2 AND status <> $1 RETURNING id"
	row := r.db.Raw(query, ds.StatusDeleted, id).Row()
	var updatedID uint
	if err := row.Scan(&updatedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("компонент не найден")
		}
		return err
	}
	return nil
}
