package repository

import (
	"aquaculture-water-feeding-control/backend/internal/model"
	"fmt"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(username string) (model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return user, fmt.Errorf("find user by username %s: %v", username, err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(id uint) (model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return user, fmt.Errorf("find user by id %d: %v", id, err)
	}
	return user, nil
}
