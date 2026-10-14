package ds

import "time"

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusDeleted   Status = "deleted"
)

func (s Status) Label() string {
	switch s {
	case StatusPublished:
		return "Опубликован"
	case StatusDraft:
		return "Черновик"
	case StatusDeleted:
		return "Удалён"
	default:
		return string(s)
	}
}

type ConfigType string

const (
	ConfigTypeSingle      ConfigType = "Single Node"
	ConfigTypeClustering  ConfigType = "Clustering"
	ConfigTypeReplication ConfigType = "Replication"
)

func (c ConfigType) Label() string {
	switch c {
	case ConfigTypeSingle:
		return "Один узел"
	case ConfigTypeClustering:
		return "Кластеризация"
	case ConfigTypeReplication:
		return "Репликация"
	default:
		return string(c)
	}
}

func (c ConfigType) IsValid() bool {
	switch c {
	case ConfigTypeSingle, ConfigTypeClustering, ConfigTypeReplication:
		return true
	default:
		return false
	}
}

type Component struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Name          string     `gorm:"type:varchar(100);not null" json:"name"`
	Description   string     `gorm:"type:varchar(1000);not null;default:''" json:"description"`
	Status        Status     `gorm:"type:varchar(20);not null" json:"-"`
	ImageURL      string     `gorm:"type:varchar(500);not null;default:''" json:"image_url"`
	VideoURL      string     `gorm:"type:varchar(500);not null;default:''" json:"video_url"`
	ConfigType    ConfigType `gorm:"type:varchar(30);not null;default:''" json:"config_type"`
	UptimePercent *float32   `gorm:"type:real" json:"uptime_percent"`
	SystemImpact  *float32   `gorm:"type:real" json:"system_impact"`
	CreatedAt     time.Time  `gorm:"not null" json:"created_at"`
	FormedAt      *time.Time `gorm:"type:timestamptz" json:"formed_at"`
	CreatorID     uint       `gorm:"not null" json:"-"`
	Creator       User       `gorm:"foreignKey:CreatorID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT" json:"-"`
}

// ComponentListSerializer — краткий ответ списка опубликованных услуг.
// IsOwner равен 1, если создатель услуги совпадает с текущим пользователем, иначе 0.
type ComponentListSerializer struct {
	ID            uint       `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	ImageURL      string     `json:"image_url"`
	VideoURL      string     `json:"video_url"`
	ConfigType    ConfigType `json:"config_type"`
	UptimePercent *float32   `json:"uptime_percent"`
	SystemImpact  *float32   `json:"system_impact"`
	LikesCount    int64      `json:"likes_count"`
	IsOwner       int        `json:"is_owner"`
}

// ComponentDetailSerializer — подробная услуга без статуса и без вложенного пользователя.
type ComponentDetailSerializer struct {
	Component
	LikesCount int64 `json:"likes_count"`
}

// ComponentFeedSerializer — лента: подробная услуга и признак, лайкнул ли её текущий пользователь.
// IsLiked равен 1, если лайк есть, иначе 0.
type ComponentFeedSerializer struct {
	ComponentDetailSerializer
	IsLiked int `json:"is_liked"`
}

func (Component) TableName() string {
	return "components"
}
