package dao

import (
	"context"
	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/gorm"
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (d *UserDAO) Create(ctx context.Context, user *model.User) error {
	return d.db.WithContext(ctx).Create(user).Error
}

func (d *UserDAO) Update(ctx context.Context, user *model.User) error {
	return d.db.WithContext(ctx).Save(user).Error
}

func (d *UserDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (d *UserDAO) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := d.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *UserDAO) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := d.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
