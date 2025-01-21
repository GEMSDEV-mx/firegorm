package main

import (
	"context"
	"log"

	"github.com/GEMSDEV-mx/firegorm"
)

type Task struct {
	firegorm.BaseModel
	Title       string `firestore:"title" json:"title"`
	Description string `firestore:"description" json:"description"`
	Done        bool   `firestore:"done" json:"done"`
}

func main() {
	// Initialize Firegorm
	credentials := loadCredentials()
	if err := firegorm.Init(credentials); err != nil {
		log.Fatalf("Failed to initialize Firegorm: %v", err)
	}

	// Initialize Task Model
	task := &Task{}
	task.SetCollectionName("tasks")

	ctx := context.Background()

	// Create
	taskData := &Task{
		Title:       "Buy Groceries",
		Description: "Milk, Eggs, Bread, Butter",
		Done:        false,
	}
	if err := task.Create(ctx, taskData); err != nil {
		log.Fatalf("Failed to create task: %v", err)
	}
	log.Printf("Task created with ID: %s", taskData.ID)

	// Fetch
	fetchedTask := &Task{}
	if err := task.Get(ctx, taskData.ID, fetchedTask); err != nil {
		log.Fatalf("Failed to fetch task: %v", err)
	}
	log.Printf("Fetched Task: %+v", fetchedTask)
}


