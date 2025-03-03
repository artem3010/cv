package main

import (
	"fmt"
	"github.com/artem3010/cv/handler"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler.HomeHandler)
	fmt.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
