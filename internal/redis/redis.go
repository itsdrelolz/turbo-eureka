package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

type JobPayload struct {
	JobID string `json:"job_id"`
	S3Key string `json:"s3_key"`
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

	payload := JobPayload{
		JobID: jobID.String(),
		S3Key: s3Key,
	}

	data, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("failed to marshal job payload: %w", err)
	}

	err = v.Client.LPush(ctx, "queue", data).Err()

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

	payloadJSON := result[1]

	var payload JobPayload

	err = json.Unmarshal([]byte(payloadJSON), &payload)

	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to unmarshal job payload: %w", err)
	}

	jobID, parseErr := uuid.Parse(payload.JobID)

	if parseErr != nil {
		return uuid.Nil, "", fmt.Errorf("invalid ID in the queue: %q, Errof: %w", payload.JobID, parseErr)
	}
	return jobID, payload.S3Key, nil
}
