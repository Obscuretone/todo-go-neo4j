package controllers

import (
	"encoding/json"
	"net/http"
	"todo-go/app/models"
	"todo-go/app/services"

	"github.com/gorilla/mux"
)

// TaskController handles HTTP requests for tasks.
type TaskController struct {
	Service *services.TaskService
}

// NewTaskController creates a new TaskController.
func NewTaskController(service *services.TaskService) *TaskController {
	return &TaskController{Service: service}
}

// GetTasks handles GET /tasks.
func (c *TaskController) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := c.Service.GetTasks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// CreateTask handles POST /tasks.
func (c *TaskController) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	newTask, err := c.Service.CreateTask(r.Context(), &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

// GetTaskByID handles GET /tasks/{taskID}.
func (c *TaskController) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	taskID := mux.Vars(r)["taskID"]
	task, err := c.Service.GetTaskByID(r.Context(), taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if task == nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// UpdateTask handles PUT /tasks/{taskID}.
func (c *TaskController) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID := mux.Vars(r)["taskID"]
	var updates struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Now, pass the title and completed fields to the service
	err := c.Service.UpdateTask(r.Context(), taskID, updates.Title, updates.Completed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task updated successfully"))
}

// DeleteTask handles DELETE /tasks/{taskID}.
func (c *TaskController) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := mux.Vars(r)["taskID"]
	err := c.Service.DeleteTask(r.Context(), taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
