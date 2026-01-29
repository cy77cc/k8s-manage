package node

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cy77cc/k8s-manage/internal/constants"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type NodeDao struct {
	db    *gorm.DB
	cache *expirable.LRU[string, any]
	rdb   redis.UniversalClient
}

func NewNodeDao(db *gorm.DB, cache *expirable.LRU[string, any], rdb redis.UniversalClient) *NodeDao {
	return &NodeDao{
		db:    db,
		cache: cache,
		rdb:   rdb,
	}
}

func (d *NodeDao) Create(ctx context.Context, node *model.Node) error {
	if err := d.db.WithContext(ctx).Create(node).Error; err != nil {
		return err
	}

	key := fmt.Sprintf("%s%d", constants.NodeKey, node.ID)

	if bs, err := json.Marshal(node); err == nil {
		d.rdb.SetEx(ctx, key, bs, constants.RdbTTL)
	}

	return nil
}

func (d *NodeDao) Update(ctx context.Context, node *model.Node) error {
	// 双删策略
	key := fmt.Sprintf("%s%d", constants.NodeKey, node.ID)
	if err := d.rdb.Del(ctx, key).Err(); err != nil {
		return nil
	}

	if err := d.db.WithContext(ctx).Save(node).Error; err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)
	if err := d.rdb.Del(ctx, key).Err(); err != nil {
		return nil
	}
	return nil
}

func (d *NodeDao) Delete(ctx context.Context, id model.NodeID) error {
	key := fmt.Sprintf("%s%d", constants.NodeKey, id)
	if err := d.rdb.Del(ctx, key).Err(); err != nil {
		return nil
	}

	if err := d.db.WithContext(ctx).Delete(&model.Node{}, id).Error; err != nil {
		return err
	}
	return nil
}
