package main

import (
	_ "download_service/docs"
	"download_service/internal/routes"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Download Service API
// @version 1.0
// @description API для управления задачами загрузки файлов
// @host localhost:8080 (default)
// @BasePath /
// @schemes http
func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//setup port
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080" // default port
	}

	r := mux.NewRouter()
	// Routes
	r.HandleFunc("/createTask", routes.CreateTask).Methods("POST")
	r.HandleFunc("/addTaskItems", routes.AddTaskItems).Methods("POST")
	r.HandleFunc("/taskStatus", routes.GetTaskStatus).Methods("GET")
	r.HandleFunc("/download/{file_id}", routes.Download).Methods("GET")

	// Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Printf("Server started on port :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
