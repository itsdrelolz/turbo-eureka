package api 

import (
	"net/http"
	"strings"
	"encoding/json"
	"os"
	"path/filepath"
	"job-matcher/internal/objectstore/s3"
)


// TODO 
// Add s3 
// Add job to postgres with status "queued"
// get jobid back from postgres 
// add job to redis 



func handleUploadResum(w http.ResponseWriter, r *http.Request) {
	// limit file to size of 10MB
	
	r.Body = http.MaxBytesReader(w, r.Body, 10 << 20)



	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"jobId" : "1"})
}



func createFile(filename string) (*os.File, error) {

	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	dst, err := os.Create(filepath.Join("uploads", filename))

	if err != nil {
		return nil, err
	}

	return dst, nil
}

func isValidFileType(file []byte) bool {
	fileType := http.DetectContentType(file)
	return strings.HasPrefix(fileType, "application/pdf")
}


func handleGetResumeStatus(w http.ResponseWriter, r *http.Request)  {
	
	
}
