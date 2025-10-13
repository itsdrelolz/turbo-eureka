package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func fileUploadHandler(w http.ResponseWriter, r *http.Request) {
	// limit file to size of 10MB

	r.Body = http.MaxBytesReader(w, r.Body, 10<<30)

	file, handler, err := r.FormFile("myFile")

	if err != nil {
		http.Error(w, "error retrieving the file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fmt.Println(w, "Uploaded file: %s\n", handler.Filename)
	fmt.Println(w, "Uploaded file of size: %s\n", handler.Size)
	fmt.Println(w, "MIME Header %v\n", handler.Header)

	// save file locally
	dst, err := createFile(handler.Filename)

	if err != nil {
		http.Error(w, "error saving the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// copy the uploaded file to the destination file
	if _, err := dst.ReadFrom(file); err != nil {
		http.Error(w, "error saving file", http.StatusInternalServerError)
	}

	fileBytes, err := io.ReadAll(file)

	if err != nil {
		http.Error(w, "Invalid file ", http.StatusBadRequest)
		return
	}

	if !isValidFileType(fileBytes) {
		http.Error(w, "Invalid file type, only pdfs are allowed", http.StatusInternalServerError)
		return
	}

	if _, err := dst.Write(fileBytes); err != nil {
		http.Error(w, "error saving file", http.StatusInternalServerError)
	}

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
	return strings.HasSuffix(fileType, ".pdf/")
}

func main() {
	http.HandleFunc("/upload", fileUploadHandler)
	fmt.Println("Starting server on port 8080:")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
