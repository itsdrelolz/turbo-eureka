package queue

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func New(ctx context.Context, connectionString string) (*RedisClient, error) { 


	opt, err := redis.ParseURL(connectionString)
	if err != nil {
		return nil, fmt.Errorf("could not parse redis connection string: %w", err)
	}



	rdb := redis.NewClient(opt)


	
	if err := rdb.Ping(ctx).Err(); err != nil { 	
		rdb.Close()
		return nil, fmt.Errorf("unable to establish a connection to the database %w", err)
	}
	
	return &RedisClient{Client: rdb}, nil
}


func (c *RedisClient) Close() error  { 
 	return c.Client.Close()
}
