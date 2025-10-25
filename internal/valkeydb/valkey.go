package valkeydb

import (
	"context"
	"fmt"

	"github.com/valkey-io/valkey-go"
)

type ValkeyClient struct {
	Client valkey.Client
}

func New(ctx context.Context, connectionString []string) (*ValkeyClient, error) {
    client, err := valkey.NewClient(valkey.ClientOption{
        InitAddress: connectionString,
    })
    if err != nil {
        return nil, fmt.Errorf("unable to create Valkey client: %w", err)
    }
    
    if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
        client.Close() // Clean up the client if ping fails
        return nil, fmt.Errorf("unable to ping Valkey: %w", err)
    }
    
    return &ValkeyClient{Client: client}, nil
}







func (v *ValkeyClient) Close()  {
	 v.Client.Close()
}



// This method should insert a job, and then update the job as pending in the postgres db 
func (v *ValkeyClient) InsertJob(ctx context.Context, jobID string) error { 

	
	if jobID == "" { 
		return fmt.Errorf("Job Id not found")
	}

	
	cmd := v.Client.B().Lpush().
        Key("job-queue").
        Element(jobID).
	Build()
	
	if _, err := v.Client.Do(ctx, cmd).AsInt64(); err != nil {
       		return fmt.Errorf("unable to add job (%s) to the queue: %w", jobID, err)
    	}
	
	return nil
}


