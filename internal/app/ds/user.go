package ds

type User struct {
	ID          uint   `gorm:"primaryKey"`
	Login       string `gorm:"type:varchar(25);unique;not null"`
	Password    string `gorm:"type:varchar(100);not null"`
	IsModerator bool   `gorm:"type:boolean;not null;default:false"`
}

func (User) TableName() string {
	return "users"
}
