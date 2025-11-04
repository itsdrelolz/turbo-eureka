package valkeydb

import (
	"context"
	"fmt"
	"github.com/valkey-io/valkey-go"
)

type ValkeyClient struct {
	Client valkey.Client
}

func New(ctx context.Context, address string, password string) (*ValkeyClient, error) {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{address},
		Password:    password,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create Valkey client: %w", err)
	}

	if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
		client.Close()
		return nil, fmt.Errorf("unable to ping Valkey: %w", err)
	}

	return &ValkeyClient{Client: client}, nil
}

func (v *ValkeyClient) Close() {
	v.Client.Close()
}

// This method should insert a job, and then update the job as pending in the postgres db
func (v *ValkeyClient) InsertJob(ctx context.Context, jobID string) error {

	cmd := v.Client.B().Lpush().
		Key("job-queue").
		Element(jobID).
		Build()

	if _, err := v.Client.Do(ctx, cmd).AsInt64(); err != nil {

		return fmt.Errorf("unable to add job (%s) to the queue: %w", jobID, err)
	}

	return nil
}

func (v *ValkeyClient) ConsumeJob(ctx context.Context) (string, error) {

	cmd := v.Client.B().Brpop().
		Key("job-queue").
		Timeout(0).
		Build()

	res := v.Client.Do(ctx, cmd)

	arr, err := res.AsStrSlice()


	if err != nil {
		return "", fmt.Errorf("failed to parse blocking right pop respose: %w", err)
	}

	queuedJob := arr[1]

	return queuedJob, nil

}
