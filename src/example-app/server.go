package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", hello)
	fmt.Println("listening...")

	port := os.Getenv("PORT")
	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           nil,
		ReadHeaderTimeout: 5 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(res, "Hello")
}
