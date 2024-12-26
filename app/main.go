package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"todo-go/app/config"
	"todo-go/app/controllers"
	"todo-go/app/routes"
	"todo-go/app/services"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize Neo4j connection
	neo4jDriver, err := config.InitNeo4j()
	if err != nil {
		log.Fatal("Failed to initialize Neo4j connection:", err)
	}
	defer neo4jDriver.Close(context.Background())

	// Initialize the service layer
	taskService := services.NewTaskService(neo4jDriver) // Pass Neo4j driver

	// Initialize the controller layer
	taskController := controllers.NewTaskController(taskService) // Pass TaskService instance

	// Setup HTTP server
	router := mux.NewRouter()
	routes.RegisterRoutes(router, taskController)

	fmt.Println("Server is running on http://0.0.0.0:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", router))
}
