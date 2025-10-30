package api

import (
	"bytes"
	"encoding/json"
	"job-matcher/mocks" 	
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleUploadResume_Success(t *testing.T) {

	testUUID := uuid.New()

	// The database returns this struct
	expectedDBResult := uuid.NullUUID{
		UUID:  testUUID,
		Valid: true,
	}
	expectedJobID := testUUID.String()
	mockDB := new(mocks.MockJobStore)
	mockQueue := new(mocks.MockJobQueuer)
	mockStore := new(mocks.MockFileStorer)

	expectedFileURL := "http://s3.com/mock-uuid.pdf"

	mockStore.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedFileURL, nil)

	mockDB.On("InsertJobReturnID", mock.Anything, expectedFileURL, mock.Anything).Return(expectedDBResult, nil)

	mockQueue.On("InsertJob", mock.Anything, expectedJobID).Return(nil)
	handler := NewAPIHandler(mockDB, mockQueue, mockStore, "fake-bucket-name")

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("resume", "test-resume.pdf")
	part.Write([]byte("this is a fake PDF"))
	writer.Close()

	req := httptest.NewRequest("POST", "/resumes", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	// Call the handler function that we are testing.
	handler.HandleUploadResume(rr, req)

	// ASSERT
	assert.Equal(t, http.StatusOK, rr.Code)

	var responseBody map[string]string
	json.Unmarshal(rr.Body.Bytes(), &responseBody)

	assert.Equal(t, expectedJobID, responseBody["jobId"])
	mockDB.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
	mockStore.AssertExpectations(t)
}
