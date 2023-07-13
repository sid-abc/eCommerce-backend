package main

import (
	"example/ecommerce/database"
	"example/ecommerce/server"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := database.ConnectAndMigrate(
		os.Getenv("host"),
		os.Getenv("port"),
		os.Getenv("databaseName"),
		os.Getenv("user"),
		os.Getenv("password"),
		database.SSLModeDisable); err != nil {
		logrus.Fatalf("Failed to initialize and migrate database with error: %+v", err)
	}

	r := server.SetUpRoutes()

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
