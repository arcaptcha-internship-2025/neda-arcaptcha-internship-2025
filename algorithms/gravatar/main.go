package main

import (
	"gravatar/handlers"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/avatar", handlers.HandleGravatarRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
