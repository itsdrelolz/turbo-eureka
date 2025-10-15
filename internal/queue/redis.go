package pkg

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

func example() { 

	rdb := redis.NewClient(&redis.Options{
	Addr:  "localhost:6739",
	Password: "", 
	DB: 0,
})


	err := rdb.Set(ctx, "key", "value", 0).Err()

	if err != nul {
		panic(err)
	}



