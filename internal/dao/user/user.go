package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cy77cc/k8s-manage/internal/constants"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserDAO struct {
	db    *gorm.DB
	cache *expirable.LRU[string, any]
	rdb   redis.UniversalClient
}

func NewUserDAO(db *gorm.DB, cache *expirable.LRU[string, any], rdb redis.UniversalClient) *UserDAO {
	return &UserDAO{db: db, cache: cache, rdb: rdb}
}

func (d *UserDAO) Create(ctx context.Context, user *model.User) error {
	if err := d.db.WithContext(ctx).Create(user).Error; err != nil {
		return err
	}
	key := fmt.Sprintf("%s%d", constants.UserIdKey, user.ID)
	if bs, err := json.Marshal(&user); err == nil {
		d.rdb.SetEx(ctx, key, bs, constants.RdbTTL)
	}
	return nil
}

func (d *UserDAO) Update(ctx context.Context, user *model.User) error {
	// 先删除redis，再写数据库

	key := fmt.Sprintf("%s%d", constants.UserIdKey, user.ID)
	if err := d.rdb.Del(ctx, key).Err(); err != nil {
		return err
	}

	if err := d.db.WithContext(ctx).Save(user).Error; err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)
	// 延迟双删
	if err := d.rdb.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}

func (d *UserDAO) Delete(ctx context.Context, id model.UserID) error {
	key := fmt.Sprintf("%s%d", constants.UserIdKey, id)
	d.rdb.Del(ctx, key)
	return d.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (d *UserDAO) FindOneById(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	// 先从redis获取数据
	key := fmt.Sprintf("%s%d", constants.UserIdKey, id)
	buf, err := d.rdb.Get(ctx, key).Bytes()
	if err == nil {
		if err := json.Unmarshal(buf, &user); err == nil {
			// 续约，加时间，方式缓存雪崩，穿透
			utils.ExtendTTL(ctx, d.rdb, key)
			return &user, nil
		}
	}
	err = d.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(&user)
	if err == nil {
		d.rdb.Set(ctx, key, b, constants.RdbTTL)
	}

	return &user, nil
}

func (d *UserDAO) FindOneByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	key := fmt.Sprintf("%s%s", constants.UserNameKey, username)
	buf, err := d.rdb.Get(ctx, key).Bytes()
	if err == nil {
		if err := json.Unmarshal(buf, &user); err == nil {
			// 不处理error，可以容忍失败
			utils.ExtendTTL(ctx, d.rdb, key)
			return &user, nil
		}
	}
	err = d.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(&user)
	if err == nil {
		d.rdb.Set(ctx, key, b, constants.RdbTTL)
	}

	return &user, nil
}
