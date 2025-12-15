package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("organisation-service starting on :8080")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"organisation-service"}`))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
