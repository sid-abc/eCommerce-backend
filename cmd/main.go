package main

import (
	"example/ecommerce/database"
	"example/ecommerce/server"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

func main() {
	if err := database.ConnectAndMigrate(
		"localhost",
		"5433",
		"ecom",
		"local",
		"local",
		database.SSLModeDisable); err != nil {
		logrus.Panicf("Failed to initialize and migrate database with error: %+v", err)
	}

	r := server.SetUpRoutes()

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
