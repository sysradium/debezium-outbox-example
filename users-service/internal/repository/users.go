package repository

import (
	"fmt"

	"github.com/sysradium/debezium-outbox-example/users-service/internal/domain"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox/debezium"
	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"index"`
	FirstName string `gorm:"index"`
	LastName  string `gorm:"index"`
}

func (u User) ToDomain() domain.User {
	return domain.User{
		ID:        u.ID,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}

type TxFn func(repo Repository) (domain.User, error)

type UserRepository struct {
	db            *gorm.DB
	outboxFactory func(*gorm.DB) outbox.Storer
}

func newFromDomainUser(u domain.User) User {
	return User{
		Username:  u.Username,
		LastName:  u.LastName,
		FirstName: u.FirstName,
	}
}

func NewUserRepository(
	db *gorm.DB,
) *UserRepository {
	return &UserRepository{
		db: db,
		outboxFactory: func(db *gorm.DB) outbox.Storer {
			return debezium.NewOutboxPublisher(db)
		},
	}
}

func (r *UserRepository) Create(user domain.User) (domain.User, error) {
	u := newFromDomainUser(user)
	result := r.db.Create(&u)
	if result.Error != nil {
		return domain.User{}, result.Error
	}
	return u.ToDomain(), nil
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

func (r *UserRepository) Outbox() outbox.Storer {
	return r.outboxFactory(r.db)
}

func (r *UserRepository) Atomic(fn TxFn) (rUser domain.User, rErr error) {
	tx := r.db.Begin()

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

	registry := r
	r.db = tx

	return fn(registry)
}
