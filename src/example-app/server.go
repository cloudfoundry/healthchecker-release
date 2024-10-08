package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", hello)
	fmt.Println("listening...")

	port := os.Getenv("PORT")
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: nil,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(res, "Hello")
}
