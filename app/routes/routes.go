package routes

import (
	"net/http"
	"todo-go/app/controllers"

	"github.com/gorilla/mux"
)

// RegisterRoutes sets up all routes for the application.
func RegisterRoutes(router *mux.Router, taskController *controllers.TaskController) {
	router.HandleFunc("/tasks", taskController.GetTasks).Methods(http.MethodGet)
	router.HandleFunc("/tasks", taskController.CreateTask).Methods(http.MethodPost)
	router.HandleFunc("/tasks/{taskID}", taskController.GetTaskByID).Methods(http.MethodGet)
	router.HandleFunc("/tasks/{taskID}", taskController.UpdateTask).Methods(http.MethodPut)
	router.HandleFunc("/tasks/{taskID}", taskController.DeleteTask).Methods(http.MethodDelete)
}
