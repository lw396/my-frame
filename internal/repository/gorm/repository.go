package gorm

import (
	"my-frame/internal/repository"

	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) repository.Repository {
	return &gormRepository{
		db: db,
	}
}
