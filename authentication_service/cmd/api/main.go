package main

import (
	"fmt"
	"net/http"

	"github.com/BernardN38/flutter-backend/application"
)

func main() {
	application.New().Run()
	fmt.Println(http.ListenAndServe(":8080", nil))
}
