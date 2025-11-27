package redisdb

import (
	"DelayedNotifier/internal/models"
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

type RedisConnection struct {
	rdb *redis.Client
}

// Close redis connection
func (rc *RedisConnection) Close() {
	err := rc.rdb.Close()
	if err != nil {
		panic(err)
	}
}

func DeclareRedisDataBase(options redis.Options) *RedisConnection {
	rdb := redis.NewClient(&options)
	return &RedisConnection{rdb: rdb}
}

func (rc *RedisConnection) SaveMessage(ctx context.Context, notif models.Notification) error {
	n, err := rc.rdb.HSet(ctx, notif.UUID, notif).Result()
	if err != nil {
		return errors.New("Failed to save message into Redis DB")
	} else if n == 0 {
		return errors.New("got message is empty")
	}

	return nil
}

func (rc *RedisConnection) GetStatus(ctx context.Context, uuid string) (string, error) {
	isExists, err := rc.rdb.HExists(ctx, uuid, "status").Result()
	if !isExists || err != nil {
		return "", errors.New("Failed to get status from Redis DB; err: " + err.Error())
	}
	status, err := rc.rdb.HGet(ctx, uuid, "status").Result()
	if err != nil {
		return "", errors.New("Failed to get status from Redis DB")
	}

	return status, nil
}

func (rc *RedisConnection) SaveStatus(ctx context.Context, uuid string, status string) error {
	_, err := rc.rdb.HSet(ctx, uuid, "status", status).Result()
	if err != nil {
		return errors.New("Failed to save status into Redis DB")
	}
	return nil
}

func (rc *RedisConnection) DeleteMessage(ctx context.Context, uuid string) error {
	_, err := rc.rdb.HDel(ctx, uuid, "message").Result()
	if err != nil {
		return errors.New("Failed to delete message from Redis DB")
	}
	return nil
}
