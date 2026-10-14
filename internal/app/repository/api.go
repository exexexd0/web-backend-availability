package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/one-compressive/web-backend-availability/internal/app/ds"
	"gorm.io/gorm"
)

func (r *Repository) GetPublishedWithCreator(id uint) (*ds.Component, error) {
	var component ds.Component
	err := r.db.Where("id = ? AND status = ?", id, ds.StatusPublished).First(&component).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &component, nil
}

func (r *Repository) GetFirstPublishedComponent() (*ds.Component, error) {
	var component ds.Component
	err := r.db.Where("status = ?", ds.StatusPublished).Order("id ASC").First(&component).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &component, nil
}

func (r *Repository) GetOwnDraft(creatorID uint) (*ds.Component, error) {
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

func (r *Repository) GetComponentWithCreator(id uint) (*ds.Component, error) {
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

func (r *Repository) PublishOwnComponent(id, creatorID uint, uptimePercent, systemImpact float32) error {
	var component ds.Component
	err := r.db.Where("id = ? AND status <> ?", id, ds.StatusDeleted).First(&component).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if component.CreatorID != creatorID {
		return ErrNotFound
	}
	if component.Status != ds.StatusDraft {
		return ErrNotDraft
	}

	now := time.Now()
	result := r.db.Model(&ds.Component{}).
		Where("id = ? AND creator_id = ? AND status = ?", id, creatorID, ds.StatusDraft).
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
		return ErrNotDraft
	}
	return nil
}

func (r *Repository) SoftDeleteOwnComponent(id, creatorID uint) error {
	result := r.db.Model(&ds.Component{}).
		Where("id = ? AND creator_id = ? AND status <> ?", id, creatorID, ds.StatusDeleted).
		Update("status", ds.StatusDeleted)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) UserLiked(userID, componentID uint) (bool, error) {
	var count int64
	err := r.db.Model(&ds.ComponentLike{}).
		Where("user_id = ? AND component_id = ?", userID, componentID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) SetComponentLike(userID, componentID uint, value int) error {
	var component ds.Component
	err := r.db.Where("id = ? AND status = ?", componentID, ds.StatusPublished).First(&component).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	if value == 0 {
		return r.db.Where("user_id = ? AND component_id = ?", userID, componentID).Delete(&ds.ComponentLike{}).Error
	}

	like := ds.ComponentLike{UserID: userID, ComponentID: componentID}
	return r.db.Where("user_id = ? AND component_id = ?", userID, componentID).FirstOrCreate(&like).Error
}

func (r *Repository) RegisterUser(login, password string) (*ds.User, error) {
	user := ds.User{
		Login:       login,
		Password:    password,
		IsModerator: false,
	}
	if err := r.db.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrLoginTaken
		}
		return nil, err
	}
	return &user, nil
}
