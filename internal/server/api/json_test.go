package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m1khal3v/gometheus/pkg/response"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string `json:"name" valid:"required"`
	Email string `json:"email" valid:"email,required"`
}

func TestDecodeAndValidateJSONRequest(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedResult *TestStruct
		expectedValid  bool
		expectedStatus int
	}{
		{
			name:           "Valid JSON request",
			requestBody:    `{"name":"John Doe", "email":"john.doe@example.com"}`,
			expectedResult: &TestStruct{Name: "John Doe", Email: "john.doe@example.com"},
			expectedValid:  true,
		},
		{
			name:           "Invalid JSON request",
			requestBody:    `{"name":"John Doe"}`,
			expectedResult: nil,
			expectedValid:  false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty request body",
			requestBody:    ``,
			expectedResult: nil,
			expectedValid:  false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tc.requestBody))
			writer := httptest.NewRecorder()

			result, isValid := DecodeAndValidateJSONRequest[TestStruct](request, writer)

			if tc.expectedValid {
				assert.True(t, isValid)
				assert.Equal(t, tc.expectedResult, result)
				res := writer.Result()
				defer res.Body.Close()
				assert.Equal(t, http.StatusOK, res.StatusCode)
			} else {
				assert.False(t, isValid)
				res := writer.Result()
				defer res.Body.Close()
				assert.Equal(t, tc.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestDecodeAndValidateJSONRequests(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedResult []*TestStruct
		expectedValid  bool
		expectedStatus int
	}{
		{
			name:           "Valid JSON array request",
			requestBody:    `[{"name":"John Doe", "email":"john.doe@example.com"}, {"name":"Jane Roe", "email":"jane.roe@example.com"}]`,
			expectedResult: []*TestStruct{{Name: "John Doe", Email: "john.doe@example.com"}, {Name: "Jane Roe", Email: "jane.roe@example.com"}},
			expectedValid:  true,
		},
		{
			name:           "Invalid JSON array request",
			requestBody:    `[{"name":"John Doe"}, {"name":"Jane Roe", "email":"jane.roe@example.com"}]`,
			expectedResult: nil,
			expectedValid:  false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty request body",
			requestBody:    `[]`,
			expectedResult: nil,
			expectedValid:  false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tc.requestBody))
			writer := httptest.NewRecorder()

			result, isValid := DecodeAndValidateJSONRequests[TestStruct](request, writer)

			if tc.expectedValid {
				assert.True(t, isValid)
				assert.Equal(t, tc.expectedResult, result)
				res := writer.Result()
				defer res.Body.Close()
				assert.Equal(t, http.StatusOK, res.StatusCode)
			} else {
				assert.False(t, isValid)
				res := writer.Result()
				defer res.Body.Close()
				assert.Equal(t, tc.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestWriteJSONResponse(t *testing.T) {
	writer := httptest.NewRecorder()
	responseData := map[string]string{"message": "success"}
	WriteJSONResponse(responseData, writer)

	res := writer.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", writer.Header().Get("Content-Type"))

	var actualResponse map[string]string
	err := json.NewDecoder(writer.Body).Decode(&actualResponse)
	assert.NoError(t, err)
	assert.Equal(t, responseData, actualResponse)
}

func TestWriteJSONErrorResponse(t *testing.T) {
	writer := httptest.NewRecorder()

	WriteJSONErrorResponse(http.StatusBadRequest, writer, "Error occurred", errors.New("some error"))

	res := writer.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", writer.Header().Get("Content-Type"))

	var actualResponse response.APIError
	err := json.NewDecoder(writer.Body).Decode(&actualResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Error occurred", actualResponse.Message)
}
