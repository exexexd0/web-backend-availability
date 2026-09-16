package ds

type ComponentLike struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null;uniqueIndex:idx_component_like"`
	ComponentID uint      `gorm:"not null;uniqueIndex:idx_component_like"`
	User        User      `gorm:"foreignKey:UserID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	Component   Component `gorm:"foreignKey:ComponentID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}

func (ComponentLike) TableName() string {
	return "component_likes"
}
