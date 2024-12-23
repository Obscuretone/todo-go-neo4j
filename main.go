package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Task struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Completed bool    `json:"completed"`
	ParentID  *string `json:"parent_id"` // Use a pointer to string here
}

var neo4jDriver neo4j.DriverWithContext

func main() {
	// Initialize Neo4j driver
	var err error
	neo4jDriver, err = neo4j.NewDriverWithContext("neo4j://localhost:7687", neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		log.Fatal("Failed to create driver:", err)
	}
	defer neo4jDriver.Close(context.Background())

	// Setup HTTP server
	r := mux.NewRouter()
	r.HandleFunc("/tasks", getTasks).Methods("GET")
	r.HandleFunc("/tasks", createTask).Methods("POST")
	r.HandleFunc("/tasks/{taskID}", getTaskByID).Methods("GET")
	r.HandleFunc("/tasks/{taskID}", updateTask).Methods("PUT")
	r.HandleFunc("/tasks/{taskID}", deleteTask).Methods("DELETE")

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	session := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(context.Background())

	tasks, err := session.ExecuteRead(context.Background(), func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(context.Background(),
			"MATCH (t:Task) "+
				"OPTIONAL MATCH (t)-[:HAS_PARENT]->(p:Task) "+
				"RETURN t.id AS id, t.title AS title, t.completed AS completed, p.id AS parent_id",
			nil,
		)
		if err != nil {
			return nil, err
		}

		var tasks []Task
		for result.Next(context.Background()) {
			record := result.Record()

			// Initialize parentID to nil
			var parentID *string

			// Safely check if the parent_id exists and is not nil
			if record.Values[3] != nil {
				id := record.Values[3].(string) // Ensure type assertion is safe
				parentID = &id                  // Assign parentID pointer
			}

			// Append the task, using nil parentID when no parent is set
			tasks = append(tasks, Task{
				ID:        record.Values[0].(string),
				Title:     record.Values[1].(string),
				Completed: record.Values[2].(bool),
				ParentID:  parentID, // This will be nil if no parent
			})
		}
		return tasks, nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Assign a new unique ID to the task
	task.ID = uuid.New().String()
	task.Completed = false

	// Start a Neo4j session
	session := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(context.Background())

	_, err := session.ExecuteWrite(context.Background(), func(tx neo4j.ManagedTransaction) (any, error) {
		// Create the new task node
		_, err := tx.Run(context.Background(),
			"CREATE (t:Task {id: $id, title: $title, completed: $completed}) RETURN t",
			map[string]any{
				"id":        task.ID,
				"title":     task.Title,
				"completed": task.Completed,
			},
		)
		if err != nil {
			return nil, err
		}

		// If ParentID is provided, create a relationship with the parent task
		if task.ParentID != nil && *task.ParentID != "" {
			_, err := tx.Run(context.Background(),
				"MATCH (child:Task {id: $childID}), (parent:Task {id: $parentID}) "+
					"CREATE (child)-[:HAS_PARENT]->(parent)",
				map[string]any{
					"childID":  task.ID,
					"parentID": *task.ParentID, // Dereference pointer to pass the string
				},
			)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func getTaskByID(w http.ResponseWriter, r *http.Request) {
	taskID := mux.Vars(r)["taskID"]

	session := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(context.Background())

	task, err := session.ExecuteRead(context.Background(), func(tx neo4j.ManagedTransaction) (any, error) {
		// Fetch task and its parent task (if any)
		result, err := tx.Run(context.Background(),
			"MATCH (t:Task {id: $id}) "+
				"OPTIONAL MATCH (t)-[:HAS_PARENT]->(p:Task) "+
				"RETURN t.id AS id, t.title AS title, t.completed AS completed, p.id AS parent_id",
			map[string]any{"id": taskID},
		)
		if err != nil {
			return nil, err
		}

		if result.Next(context.Background()) {
			record := result.Record()
			var parentID *string
			if record.Values[3] != nil {
				id := record.Values[3].(string)
				parentID = &id
			}

			return Task{
				ID:        record.Values[0].(string),
				Title:     record.Values[1].(string),
				Completed: record.Values[2].(bool),
				ParentID:  parentID, // This is a pointer to string
			}, nil
		}
		return nil, nil
	})
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

func updateTask(w http.ResponseWriter, r *http.Request) {
	taskID := mux.Vars(r)["taskID"]
	var updates struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	session := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(context.Background())

	_, err := session.ExecuteWrite(context.Background(), func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(context.Background(),
			"MATCH (t:Task {id: $id}) SET t.title = $title, t.completed = $completed",
			map[string]any{
				"id":        taskID,
				"title":     updates.Title,
				"completed": updates.Completed,
			},
		)
		return nil, err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task updated successfully"))
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := mux.Vars(r)["taskID"]

	session := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(context.Background())

	// Delete child tasks first
	_, err := session.ExecuteWrite(context.Background(), func(tx neo4j.ManagedTransaction) (any, error) {
		// Match and delete child tasks that have a parent relationship with the task
		_, err := tx.Run(context.Background(),
			"MATCH (t:Task {id: $id})-[:HAS_PARENT]->(child:Task) "+
				"DETACH DELETE child",
			map[string]any{"id": taskID},
		)
		if err != nil {
			return nil, err
		}

		// Now delete the parent task itself
		_, err = tx.Run(context.Background(),
			"MATCH (t:Task {id: $id}) "+
				"DETACH DELETE t",
			map[string]any{"id": taskID},
		)
		return nil, err
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // No content response after successful deletion
}
