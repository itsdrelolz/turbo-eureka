package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"encoding/json"
	"job-matcher/internal/database"
	"job-matcher/internal/objectstore"
)



func main() {

	ctx := context.Background()
	dbStore , err := database.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil { 
	panic(err)
	}

	defer dbStore.Close()




	router := api.NewRouter(dbStore.Pool, redisClient.Client, s3.Client, s3Bucket)
	
	http.ListenAndServe("8080", router)
}
