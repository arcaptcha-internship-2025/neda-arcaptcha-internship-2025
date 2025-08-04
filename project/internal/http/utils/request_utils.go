package utils

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(&dst); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return err
	}
	return nil
}

func ValidateRequiredFields(w http.ResponseWriter, fields map[string]string) bool {
	for field, value := range fields {
		if value == "" {
			WriteErrorResponse(w, http.StatusBadRequest, field+" is required")
			return false
		}
	}
	return true
}

func ParseFileUpload(w http.ResponseWriter, r *http.Request, fieldName string, maxSize int64) (io.ReadCloser, *multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(maxSize); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "failed to parse form")
		return nil, nil, err
	}

	file, header, err := r.FormFile(fieldName)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "no file uploaded")
		return nil, nil, err
	}
	return file, header, nil
}

func ValidateFileType(w http.ResponseWriter, filename string, allowedTypes []string) bool {
	fileExt := strings.ToLower(filepath.Ext(filename))
	for _, allowedType := range allowedTypes {
		if fileExt == allowedType {
			return true
		}
	}

	WriteErrorResponse(w, http.StatusBadRequest,
		"invalid file type. allowed types: "+strings.Join(allowedTypes, ", "))
	return false
}
