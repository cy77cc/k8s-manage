package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/consts"
	"github.com/redis/go-redis/v9"
)

func SlicesToString[T any](s []T, sep string) string {
	if len(s) == 0 {
		return ""
	}
	var b strings.Builder
	for i, v := range s {
		if i > 0 {
			b.WriteString(sep)
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "%v", v)
	}
	return b.String()
}

func GetTimestamp() int64 {
	return time.Now().Unix()
}


func ExtendTTL(ctx context.Context, rdb redis.UniversalClient, key string) error {
    ttl, err := rdb.TTL(ctx, key).Result()
    if err != nil {
        return err
    }
    if ttl < 0 {
        // key 不存在或无过期时间，可以直接设置 add 作为 TTL
        ttl = 0
    }
    return rdb.Expire(ctx, key, ttl+consts.RdbAddTTL).Err()
}

