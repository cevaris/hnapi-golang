package main

import (
	"io"
	"net/http"
	"os"
)

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hellow World!")
}

func main() {
	port := os.Getenv("PORT")
	http.HandleFunc("/", hello)
	http.ListenAndServe(":"+port, nil)
}
