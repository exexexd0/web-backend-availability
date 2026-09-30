package main

import (
	"github.com/joho/godotenv"
	"github.com/one-compressive/web-backend-availability/internal/app/ds"
	"github.com/one-compressive/web-backend-availability/internal/app/dsn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.User{},
		&ds.Component{},
		&ds.ComponentLike{},
	)
	if err != nil {
		panic("cant migrate db")
	}

	if err := db.Exec(`
		ALTER TABLE components
			DROP COLUMN IF EXISTS short_description,
			ALTER COLUMN uptime_percent DROP NOT NULL,
			ALTER COLUMN uptime_percent DROP DEFAULT,
			ALTER COLUMN system_impact DROP NOT NULL,
			ALTER COLUMN system_impact DROP DEFAULT
	`).Error; err != nil {
		panic("cant update components columns")
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_components_one_draft_per_user
		ON components (creator_id)
		WHERE status = 'draft'
	`).Error; err != nil {
		panic("cant create draft uniqueness index")
	}
}
