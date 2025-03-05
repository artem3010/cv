package main

import (
	"fmt"
	"net/http"

	"github.com/artem3010/cv/handler"
)

func main() {
	http.HandleFunc("/", handler.HomeHandler)
	fmt.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
