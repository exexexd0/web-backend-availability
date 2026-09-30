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

func (r *Repository) GetNextComponent(currentID int) (*ds.Component, error) {
	var component ds.Component
	err := r.db.Where("status = ? AND id > ?", ds.StatusPublished, currentID).Order("id ASC").First(&component).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = r.db.Where("status = ?", ds.StatusPublished).Order("id ASC").First(&component).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &component, nil
}

func (r *Repository) CreateDraft(draft ds.Component) (*ds.Component, error) {
	existing, err := r.GetDraftByCreator(draft.CreatorID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	draft.Status = ds.StatusDraft
	draft.CreatedAt = time.Now()
	if err := r.db.Create(&draft).Error; err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *Repository) PublishComponent(id uint, uptimePercent, systemImpact float32) error {
	now := time.Now()
	result := r.db.Model(&ds.Component{}).
		Where("id = ? AND status = ?", id, ds.StatusDraft).
		Updates(map[string]interface{}{
			"uptime_percent": uptimePercent,
			"system_impact":  systemImpact,
			"status":         ds.StatusPublished,
			"formed_at":      now,
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

	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	var updatedID uint
	if err := sqlDB.QueryRow(query, ds.StatusDeleted, id).Scan(&updatedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("компонент не найден")
		}
		return err
	}

	return nil
}
