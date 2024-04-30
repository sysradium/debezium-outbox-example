package repository

import (
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"index"`
	FirstName string `gorm:"index"`
	LastName  string `gorm:"index"`
}

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(user User) (User, error) {
	result := r.DB.Create(&user)
	if result.Error != nil {
		return User{}, result.Error
	}
	return user, nil
}

func (r *UserRepository) Delete(id uint) error {
	result := r.DB.Delete(&User{}, id)
	return result.Error
}

type TxFn func(repo *UserRepository) (User, error)

func (r *UserRepository) Atomic(fn TxFn) (rUser User, rErr error) {
	tx := r.DB.Begin()

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if rErr != nil {
			xerr := tx.Rollback().Error
			if xerr != nil {
				rErr = fmt.Errorf("%s: %w", xerr.Error(), rErr)
			}
			return
		}
		rErr = tx.Commit().Error
	}()

	registry := &UserRepository{
		DB: tx,
	}

	return fn(registry)

}
