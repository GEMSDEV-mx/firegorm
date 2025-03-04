package main

import (
	"context"
	"log"

	"github.com/GEMSDEV-mx/firegorm"
)

type Task struct {
	firegorm.BaseModel
	Title       string `firestore:"title" json:"title" validate:"required"`
	Description string `firestore:"description" json:"description" validate:"required"`
	Done        bool   `firestore:"done" json:"done"`
}

func main() {
	credentials := loadCredentials()
	if err := firegorm.Init(credentials); err != nil {
		log.Fatalf("Failed to initialize Firegorm: %v", err)
	}

	// Register the Task model
	instance, err := firegorm.RegisterModel(&Task{}, "tasks")
	task := instance.(*Task) // Type assertion

	if err != nil {
		log.Fatalf("Failed to register model: %v", err)
	}

	ctx := context.Background()

	// Create Task
	taskData := &Task{
		Title:       "Buy Groceries",
		Description: "Milk, Eggs, Bread, Butter",
		Done:        false,
	}
	if err := task.Create(ctx, taskData); err != nil {
		log.Fatalf("Failed to create task: %v", err)
	}
	log.Printf("Task created with ID: %s", taskData.ID)

	// Fetch Task
	fetchedTask := &Task{}
	if err := task.Get(ctx, taskData.ID, fetchedTask); err != nil {
		log.Fatalf("Failed to fetch task: %v", err)
	}
	log.Printf("Fetched Task: %+v", fetchedTask)


	// Test listing tasks
	log.Println("Listing tasks...")
	var tasks []Task
	filters := map[string]interface{}{
		"done": false, // List only tasks where "done" is false
	}
	limit := 10
	startAfter := ""

	nextPageToken, err := task.List(ctx, filters, limit, startAfter, &tasks)
	if err != nil {
		log.Fatalf("Failed to list tasks: %v", err)
	}

	log.Printf("Listed tasks: %+v", tasks)
	log.Printf("Next page token: %s", nextPageToken)

	// Validate listing results
	if len(tasks) == 0 {
		log.Println("No tasks found matching the filters.")
	} else {
		for _, t := range tasks {
			log.Printf("Task ID: %s, Title: %s, Done: %v", t.ID, t.Title, t.Done)
		}
	}

	// Valid Update
	updates := map[string]interface{}{
		"title":       "New Title",
		"description": "Updated description",
	}
	if err := task.Update(ctx, taskData.ID, updates); err != nil {
		log.Fatalf("Case 1 - Valid Update failed: %v", err)
	} else {
		log.Println("Case 1 - Valid Update successful.")
	}

	// Invalid Update (Empty Field)
	invalidUpdatesEmpty := map[string]interface{}{
		"title": "",
	}
	if err := task.Update(ctx, taskData.ID, invalidUpdatesEmpty); err != nil {
		log.Printf("Case 2 - Invalid Update (Empty Field) failed as expected: %v", err)
	} else {
		log.Println("Case 2 - Invalid Update (Empty Field) unexpectedly succeeded.")
	}

	// Invalid Update (Nil Field)
	invalidUpdatesNil := map[string]interface{}{
		"description": nil,
	}
	if err := task.Update(ctx, taskData.ID, invalidUpdatesNil); err != nil {
		log.Printf("Case 3 - Invalid Update (Nil Field) failed as expected: %v", err)
	} else {
		log.Println("Case 3 - Invalid Update (Nil Field) unexpectedly succeeded.")
	}

	// Invalid Update (NonExistent Field)
	invalidUpdatesNonExistent := map[string]interface{}{
		"nonexistentField": "Some Value",
	}
	if err := task.Update(ctx, taskData.ID, invalidUpdatesNonExistent); err != nil {
		log.Printf("Case 4 - Invalid Update (NonExistent Field) failed as expected: %v", err)
	} else {
		log.Println("Case 4 - Invalid Update (NonExistent Field) unexpectedly succeeded.")
	}

	// Valid Partial Update
	validPartialUpdate := map[string]interface{}{
		"done": true,
	}
	if err := task.Update(ctx, taskData.ID, validPartialUpdate); err != nil {
		log.Fatalf("Case 5 - Valid Partial Update failed: %v", err)
	} else {
		log.Println("Case 5 - Valid Partial Update successful.")
	}
}
