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
)

func fileUploadHandler(w http.ResponseWriter, r *http.Request) {
	// limit file to size of 10MB
	
	r.Body = http.MaxBytesReader(w, r.Body, 10 << 20)

	file, handler, err := r.FormFile("myFile")

	if err != nil {
		http.Error(w, "error retrieving the file", http.StatusBadRequest)
		return
	}

	defer file.Close()



	fileBytes, err := io.ReadAll(file)

	if err != nil {
		http.Error(w, "Invalid file ", http.StatusBadRequest)
		return
	}

	if !isValidFileType(fileBytes) {
		http.Error(w, "Invalid file type, only PDFs are allowed", http.StatusInternalServerError)
		return
	}

	dst, err := createFile(handler.Filename)

	if err != nil { 
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return 
	}

	defer dst.Close()

	if _, err := dst.Write(fileBytes); err != nil { 
	
		http.Error(w, "error saving file", http.StatusInternalServerError)
		return 
	}

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

func main() {

	ctx := context.Background()
	dbStore , err := database.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil { 
	panic(err)
	}

	defer dbStore.Close()
	
	http.HandleFunc("/upload", fileUploadHandler)
	fmt.Println("Starting server on port 8080:")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
