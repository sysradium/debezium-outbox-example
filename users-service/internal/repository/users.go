package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/domain"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox"
	"gorm.io/gorm"
)

var _ Repository[domain.User] = (*UserRepository)(nil)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	UUID      uuid.UUID `gorm:"column:uuid;type:uuid;primary_key;default:uuid_generate_v4()"`
	Username  string    `gorm:"index"`
	FirstName string    `gorm:"index"`
	LastName  string    `gorm:"index"`
}

func (u User) ToDomain() domain.User {
	return domain.User{
		ID:        u.UUID,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}

type UserRepository struct {
	db            *gorm.DB
	outboxFactory func(*sql.DB) outbox.Storer
	logger        *slog.Logger
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
	opts ...option,
) *UserRepository {
	logger := slog.Default()

	u := &UserRepository{
		db:     db,
		logger: logger,
	}
	//	u.outboxFactory = debezium.NewOutboxPublisher(debezium.WithLogger(
	//		u.logger.With("logger", "debezium.outbox"),
	//	))

	for _, o := range opts {
		o(u)
	}

	return u
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	u := newFromDomainUser(user)
	result := r.db.WithContext(ctx).Create(&u)
	if result.Error != nil {
		return domain.User{}, result.Error
	}
	return u.ToDomain(), nil
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

func (r *UserRepository) Outbox() outbox.Storer {
	db, err := r.db.DB()
	if err != nil {
		panic(err)
	}
	return r.outboxFactory(db)
}

func (r *UserRepository) Atomic(ctx context.Context, fn TxFn[domain.User]) (rUser domain.User, rErr error) {
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

	registry := *r
	registry.db = tx

	return fn(ctx, &registry)
}
