package main

import (
	"context"
	"fmt"
	"log"

	"github.com/GEMSDEV-mx/firegorm"
)

// Task model with BaseModel embedded
// Hooks will run around Create/Update/Delete for this model.
type Task struct {
    firegorm.BaseModel
    Title       string `firestore:"title" json:"title" validate:"required"`
    Description string `firestore:"description" json:"description" validate:"required"`
    Done        bool   `firestore:"done" json:"done"`
}

func main() {
    // Load credentials and initialize Firegorm
    creds := loadCredentials()
    if err := firegorm.Init(creds); err != nil {
        log.Fatalf("Failed to initialize Firegorm: %v", err)
    }

    // Register Task model under "tasks" collection
    inst, err := firegorm.RegisterModel(&Task{}, "tasks")
    if err != nil {
        log.Fatalf("Failed to register Task model: %v", err)
    }
    taskModel := inst.(*Task)
    ctx := context.Background()

    // ----------------------
    // Register Hooks
    // ----------------------

    // PreCreate: prefix Title
    firegorm.DefaultRegistry.RegisterHook("tasks", firegorm.PreCreate, func(ctx context.Context, data interface{}) error {
        t := data.(*Task)
        t.Title = "[PRECREATE] " + t.Title
        return nil
    })

    // PostCreate: print info
    firegorm.DefaultRegistry.RegisterHook("tasks", firegorm.PostCreate, func(ctx context.Context, data interface{}) error {
        t := data.(*Task)
        fmt.Printf("PostCreate Hook fired: ID=%s, Title=%s\n", t.ID, t.Title)
        return nil
    })

    // ----------------------
    // Create tasks with hooks enabled
    // ----------------------
    fmt.Println("Creating task 1 with hooks enabled...")
    t1 := &Task{
        Title:       "Task One",
        Description: "First task",
        Done:        false,
    }
    if err := taskModel.Create(ctx, t1); err != nil {
        log.Fatalf("Failed to create task1: %v", err)
    }

    // ----------------------
    // Disable PreCreate globally
    // ----------------------
    fmt.Println("Disabling PreCreate hooks...")
    firegorm.DefaultRegistry.EnableType(firegorm.PreCreate, false)

    // Create another task (no prefix)
    fmt.Println("Creating task 2 with PreCreate disabled...")
    t2 := &Task{
        Title:       "Task Two",
        Description: "Second task",
        Done:        false,
    }
    if err := taskModel.Create(ctx, t2); err != nil {
        log.Fatalf("Failed to create task2: %v", err)
    }

    // ----------------------
    // Disable all hooks for "tasks" on PostCreate only
    // ----------------------
    fmt.Println("Disabling PostCreate for tasks only...")
    firegorm.DefaultRegistry.EnableScope("tasks", firegorm.PostCreate, false)

    // Create a third task (no prefix, no post-print)
    fmt.Println("Creating task 3 with all hooks off for this scope...")
    t3 := &Task{
        Title:       "Task Three",
        Description: "Third task",
        Done:        false,
    }
    if err := taskModel.Create(ctx, t3); err != nil {
        log.Fatalf("Failed to create task3: %v", err)
    }

    // ----------------------
    // Summary of created IDs
    // ----------------------
    fmt.Printf("Tasks created: IDs = [%s, %s, %s]\n", t1.ID, t2.ID, t3.ID)
}

