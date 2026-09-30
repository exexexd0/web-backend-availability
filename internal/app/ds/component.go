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
	ID            uint       `gorm:"primaryKey"`
	Name          string     `gorm:"type:varchar(100);not null"`
	Description   string     `gorm:"type:varchar(1000);not null;default:''"`
	Status        Status     `gorm:"type:varchar(20);not null"`
	ImageURL      string     `gorm:"type:varchar(500);not null;default:''"`
	VideoURL      string     `gorm:"type:varchar(500);not null;default:''"`
	ConfigType    ConfigType `gorm:"type:varchar(30);not null;default:''"`
	UptimePercent *float32   `gorm:"type:real"`
	SystemImpact  *float32   `gorm:"type:real"`
	CreatedAt     time.Time  `gorm:"not null"`
	FormedAt      *time.Time `gorm:"type:timestamptz"`
	CreatorID     uint       `gorm:"not null"`
	Creator       User       `gorm:"foreignKey:CreatorID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}

func (Component) TableName() string {
	return "components"
}
