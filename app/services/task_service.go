package services

import (
	"context"
	"errors"
	"todo-go/app/models"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// TaskService handles task-related operations.
type TaskService struct {
	driver neo4j.DriverWithContext
}

// NewTaskService creates a new instance of TaskService.
func NewTaskService(driver neo4j.DriverWithContext) *TaskService {
	return &TaskService{driver: driver}
}

// GetTasks retrieves all tasks from the database.
func (s *TaskService) GetTasks(ctx context.Context) ([]models.Task, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Run the Cypher query
		res, err := tx.Run(ctx,
			"MATCH (t:Task) "+
				"OPTIONAL MATCH (t)-[:HAS_PARENT]->(p:Task) "+
				"RETURN t.id AS id, t.title AS title, t.completed AS completed, p.id AS parent_id",
			nil,
		)
		if err != nil {
			return nil, err
		}

		var tasks []models.Task
		// Iterate over the results
		for res.Next(ctx) {
			record := res.Record()
			var parentID *string
			if record.Values[3] != nil {
				id := record.Values[3].(string)
				parentID = &id
			}

			tasks = append(tasks, models.Task{
				ID:        record.Values[0].(string),
				Title:     record.Values[1].(string),
				Completed: record.Values[2].(bool),
				ParentID:  parentID,
			})
		}

		// Check for any errors during iteration
		if err := res.Err(); err != nil {
			return nil, err
		}

		return tasks, nil
	})

	if err != nil {
		return nil, err
	}

	// Assert result to []models.Task
	return result.([]models.Task), nil
}

// GetTaskByID retrieves a single task by its ID.
func (s *TaskService) GetTaskByID(ctx context.Context, taskID string) (*models.Task, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// Type-assert the result returned from ExecuteRead to neo4j.Result
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(ctx,
			"MATCH (t:Task {id: $id}) "+
				"OPTIONAL MATCH (t)-[:HAS_PARENT]->(p:Task) "+
				"RETURN t.id AS id, t.title AS title, t.completed AS completed, p.id AS parent_id",
			map[string]any{"id": taskID},
		)
	})

	if err != nil {
		return nil, err
	}

	// Assert the result type to neo4j.Result
	res, ok := result.(neo4j.Result)
	if !ok {
		return nil, errors.New("failed to assert result to neo4j.Result")
	}

	if res.Next() { // Removed the ctx argument here
		record := res.Record()
		var parentID *string
		if record.Values[3] != nil {
			id := record.Values[3].(string)
			parentID = &id
		}

		return &models.Task{
			ID:        record.Values[0].(string),
			Title:     record.Values[1].(string),
			Completed: record.Values[2].(bool),
			ParentID:  parentID,
		}, nil
	}

	if err := res.Err(); err != nil {
		return nil, err
	}

	return nil, errors.New("task not found")
}

// CreateTask adds a new task to the database.
func (s *TaskService) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	// If the task ID is not set, generate a new UUID
	if task.ID == "" {
		task.ID = uuid.New().String() // Generate a new UUID for the task
	}

	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Create the task in Neo4j with the generated or provided ID
		_, err := tx.Run(ctx,
			"CREATE (t:Task {id: $id, title: $title, completed: $completed}) RETURN t",
			map[string]any{
				"id":        task.ID,
				"title":     task.Title,
				"completed": task.Completed,
			},
		)
		return nil, err
	})

	if err != nil {
		return nil, err
	}

	// If ParentID is provided, create a relationship with the parent task
	if task.ParentID != nil && *task.ParentID != "" {
		session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			_, err := tx.Run(ctx,
				"MATCH (child:Task {id: $childID}), (parent:Task {id: $parentID}) "+
					"CREATE (child)-[:HAS_PARENT]->(parent)",
				map[string]any{
					"childID":  task.ID,
					"parentID": *task.ParentID,
				},
			)
			return nil, err
		})
	}

	return task, nil
}

// UpdateTask updates an existing task's information.
func (s *TaskService) UpdateTask(ctx context.Context, taskID string, title string, completed bool) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// Retrieve the task from the database
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Match the task by ID and update the title and completion status
		_, err := tx.Run(ctx,
			"MATCH (t:Task {id: $id}) "+
				"SET t.title = $title, t.completed = $completed "+
				"RETURN t",
			map[string]any{
				"id":        taskID,
				"title":     title,
				"completed": completed,
			},
		)
		return nil, err
	})
	return err
}

// DeleteTask deletes a task and its relationships.
func (s *TaskService) DeleteTask(ctx context.Context, taskID string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// First, remove child tasks
		_, err := tx.Run(ctx,
			"MATCH (t:Task {id: $id})-[:HAS_PARENT]->(child:Task) "+
				"DETACH DELETE child",
			map[string]any{"id": taskID},
		)
		if err != nil {
			return nil, err
		}

		// Now, delete the task itself
		_, err = tx.Run(ctx,
			"MATCH (t:Task {id: $id}) "+
				"DETACH DELETE t",
			map[string]any{"id": taskID},
		)
		return nil, err
	})
	return err
}
