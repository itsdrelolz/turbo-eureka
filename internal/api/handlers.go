package api 

import (
	"net/http"
	"strings"
	"encoding/json"
	"os"
	"path/filepath"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	
)


// TODO 
// Add s3 
// Add job to postgres with status "queued"
// get jobid back from postgres 
// add job to redis 


type handler struct { 
	db *pgxpool.Pool
	s3 *s3.Client
	redis *rClient
	s3Bucket string 
}


func handleUploadResume(w http.ResponseWriter, r *http.Request) {
	// limit file to size of 10MB
	
	r.Body = http.MaxBytesReader(w, r.Body, 10 << 20)




	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"jobId" : "1"})
}




func isValidFileType(file []byte) bool {
	fileType := http.DetectContentType(file)
	return strings.HasPrefix(fileType, "application/pdf")
}


func handleGetResumeStatus(w http.ResponseWriter, r *http.Request)  {
	
	
}
