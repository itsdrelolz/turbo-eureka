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

func TestHandleUploadResumeSuccess(t *testing.T) {

	testUUID := uuid.New()

	// The database returns this struct
	expectedDBResult := testUUID

	expectedJobID := testUUID.String()
	mockDB := new(mocks.MockCreator)
	mockQueue := new(mocks.MockProducer)
	mockStore := new(mocks.MockFileUploader)

	mockStore.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockDB.On("InsertJobReturnID", mock.Anything, mock.Anything, mock.Anything).Return(expectedDBResult, nil)

	mockQueue.On("InsertJob", mock.Anything, mock.Anything).Return(nil)
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
