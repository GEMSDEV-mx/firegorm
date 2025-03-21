package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

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
	if err != nil {
		log.Fatalf("Failed to register model: %v", err)
	}
	taskModel := instance.(*Task) // Type assertion

	ctx := context.Background()

	// Create multiple tasks to test pagination (e.g., 15 tasks)
	totalTasks := 15
	for i := 1; i <= totalTasks; i++ {
		taskData := &Task{
			Title:       "Task " + strconv.Itoa(i),
			Description: fmt.Sprintf("Description for task %d", i),
			Done:        false,
		}
		if err := taskModel.Create(ctx, taskData); err != nil {
			log.Fatalf("Failed to create task %d: %v", i, err)
		}
		log.Printf("Created Task %d with ID: %s", i, taskData.ID)
	}

	// Test listing tasks with pagination
	log.Println("Testing pagination for listing tasks...")
	var allTasks []Task
	limit := 10
	startAfter := ""
	filters := map[string]interface{}{
		"done": false, // only list tasks that are not done
	}

	page := 1
	for {
		var tasksPage []Task
		nextPageToken, err := taskModel.List(ctx, filters, limit, startAfter, &tasksPage)
		if err != nil {
			log.Fatalf("Failed to list tasks on page %d: %v", page, err)
		}

		log.Printf("Page %d: Retrieved %d tasks", page, len(tasksPage))
		for _, t := range tasksPage {
			log.Printf("Task ID: %s, Title: %s", t.ID, t.Title)
		}
		allTasks = append(allTasks, tasksPage...)

		// If there's no next page token, we've fetched all tasks.
		if nextPageToken == "" {
			break
		}

		// Set token for next page and increment page count.
		startAfter = nextPageToken
		page++
	}

	log.Printf("Total tasks retrieved via pagination: %d", len(allTasks))
}
