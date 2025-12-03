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

// This method should insert a job, and then update the job as pending in the postgres db
func (v *RedisClient) Produce(ctx context.Context, jobID uuid.UUID) error {

	JobID := jobID.String()

	 err := v.Client.LPush(ctx, "queue:pending", JobID).Err()

	if err != nil {
		return fmt.Errorf("ERROR: Failed to push job into redis queue with: %w", err)
	}

	return nil
}

// moves the job into a processing list, ensuring data safety in the case of a worker crashing
func (v *RedisClient) Consume(ctx context.Context) (uuid.UUID, error) {

	jobID, err := v.Client.BLMove(ctx, "queue:pending", "queue:processing", "RIGHT", "LEFT", 0).Result()

	if err != nil {
		return uuid.Nil, fmt.Errorf("Failed to move job into processing queue with: %w", err)
	}

	JobID, err := uuid.Parse(jobID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("Failed to convert job ID into valid uuid, given: %v, with: %w", jobID, err)
	}
	return JobID, nil
}
