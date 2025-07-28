package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
)

type GravatarResponse struct {
	Ok          bool   `json:"ok"`
	GravatarUrl string `json:"gravatar_url"`
}

type ErrorResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

func HandleGravatarRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	email := r.URL.Query().Get("email")

	if email == "" {
		responce := ErrorResponse{
			Ok:      false,
			Message: "No email address provided",
		}

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responce)
		return
	}

	_, err := mail.ParseAddress(email) //for checking ivalid
	if err != nil {
		responce := ErrorResponse{
			Ok:      false,
			Message: "Invalid email address",
		}

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responce)
		return
	}

	//in test i see that there is a hashed str after avatar/ in url in responce
	//in gravatar requirement i see that we should trims whitespace and converts to lowercase then hash it and add it to the end of the url

	trimmedEmail := strings.TrimSpace(strings.ToLower(email))
	hash := md5.Sum([]byte(trimmedEmail))
	hashStr := hex.EncodeToString(hash[:])

	response := GravatarResponse{
		Ok:          true,
		GravatarUrl: "https://www.gravatar.com/avatar/" + hashStr,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}
