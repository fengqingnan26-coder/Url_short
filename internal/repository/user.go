package repository

import (
	"context"
	"errors"

	"github.com/jekyulll/url_shortener/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error // GORM 会自动将数据库生成的自增 ID 赋值给传入的结构体对象
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdatePasswordByEmail(ctx context.Context, passwordHash string, email string) (uint64, error)
	IsEmailAvailable(ctx context.Context, email string) (bool, error)
}

type userRepositoryIMpl struct {
	db *gorm.DB
}

// CreateUser implements UserRepository.
func (u *userRepositoryIMpl) CreateUser(ctx context.Context, user *model.User) error {
	return u.db.WithContext(ctx).Create(&user).Error
}

// GetUserByEmail implements UserRepository.
func (u *userRepositoryIMpl) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := u.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

// IsEmailAvailable implements UserRepository.
func (u *userRepositoryIMpl) IsEmailAvailable(ctx context.Context, email string) (bool, error) {
	var count int64
	err := u.db.WithContext(ctx).
		Model(&model.User{}).
		Where("email = ?", email).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// UpdatePasswordByEmail implements UserRepository.
func (r *userRepositoryIMpl) UpdatePasswordByEmail(ctx context.Context, passwordHash string, email string) (uint64, error) {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("email = ?", email).
		Update("password_hash", passwordHash)

	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	// 查询用户ID
	var user model.User
	err := r.db.WithContext(ctx).
		Select("id").
		Where("email = ?", email).
		First(&user).Error
	return user.ID, err
}

func NewUserRepo(db *gorm.DB) *userRepositoryIMpl {
	return &userRepositoryIMpl{
		db: db,
	}
}

var _ UserRepository = (*userRepositoryIMpl)(nil)
