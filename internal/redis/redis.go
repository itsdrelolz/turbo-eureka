package redis

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func New(ctx context.Context, address string, password string) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0,
	})

	return &RedisClient{Client: rdb}, nil
}

func (v *RedisClient) HealthCheck(ctx context.Context) error {

	err := v.Client.Ping(ctx).Err()

	if err != nil {
		return fmt.Errorf("ERROR: failed to ping memory store with: %w", err)
	}

	return nil
}


func (v *RedisClient) Produce(ctx context.Context, jobID uuid.UUID, s3Key string) error {

	err := v.Client.LPush(ctx, "queue", jobID, s3Key).Err()

	if err != nil { 
		return fmt.Errorf("failed to push job onto queue: %w", err)
	}

	return nil

}

func (v *RedisClient) Consume(ctx context.Context) (uuid.UUID, string, error) {


	result, err := v.Client.BRPop(ctx, 0, "queue").Result()



	if err != nil { 
		return uuid.Nil, "", fmt.Errorf("redis BRPop failed: %w", err)
	}

	jobIdStr, s3Key := result[1], result[2] 


	
	jobID, parseErr := uuid.Parse(jobIdStr)

	if parseErr != nil { 
		return uuid.Nil, "", fmt.Errorf("invalid ID in the queue: %q, Errof: %w", jobIdStr, parseErr)
	}
		return jobID, s3Key, nil
}
