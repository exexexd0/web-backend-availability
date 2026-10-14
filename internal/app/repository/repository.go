package repository

import (
	"errors"
	"fmt"
	"sync"

	"github.com/one-compressive/web-backend-availability/internal/app/ds"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const DefaultCreatorID = 1

var (
	ErrNotFound   = errors.New("запись не найдена")
	ErrNotDraft   = errors.New("опубликовать можно только свой черновик, вернуть в черновик нельзя")
	ErrLoginTaken = errors.New("пользователь с таким логином уже есть")

	currentUser     *ds.User
	currentUserOnce sync.Once
	currentUserErr  error
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

// CurrentUser — singleton текущего пользователя. Идентификатор зафиксирован константой.
func (r *Repository) CurrentUser() (*ds.User, error) {
	const currentUserID = DefaultCreatorID

	currentUserOnce.Do(func() {
		var user ds.User
		if err := r.db.First(&user, currentUserID).Error; err != nil {
			currentUserErr = fmt.Errorf("текущий пользователь не найден")
			return
		}
		currentUser = &user
	})
	if currentUserErr != nil {
		return nil, currentUserErr
	}
	return currentUser, nil
}
