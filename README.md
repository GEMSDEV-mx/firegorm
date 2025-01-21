# Firegorm: A Firestore ORM for Go

Firegorm is a lightweight Object-Relational Mapping (ORM) library designed to simplify interactions with Google Firestore in Go projects. It provides a structured approach for managing Firestore collections and documents with features such as validation, logging, and query handling.

---

## Features

- **Simple ORM Abstraction**: Map Firestore collections to Go structs.
- **Validation**: Use struct tags to enforce field requirements.
- **Logging**: Configurable log levels for debugging and production.
- **Soft Deletes**: Mark documents as deleted without removing them from the database.
- **Pagination**: Easily list and paginate through Firestore documents.
- **Model Registry**: Register and manage models for collections dynamically.

---

## Installation

Install Firegorm via `go get`:

bash

```console
go get github.com/GEMSDEV-mx/firegorm
```

---

## Initialization

To use Firegorm, initialize the Firestore client and configure logging:

```go
package main

import (
	"log"
	"github.com/GEMSDEV-mx/firegorm"
)

func main() {
	credentials := loadCredentials() // Load your Firebase service account key
	if err := firegorm.Init(credentials); err != nil {
		log.Fatalf("Failed to initialize Firegorm: %v", err)
	}
}
```

Set the log level via the environment variable `FIREGORM_LOG_LEVEL`. Supported levels are `DEBUG`, `INFO`, `WARN`, and `ERROR`. Default is `INFO`.

---

## Usage

### 1\. Define Your Model

Define your model struct by embedding `BaseModel` and adding your own fields:

```go
type Task struct {
	firegorm.BaseModel
	Title       string `firestore:"title" json:"title" validate:"required"`
	Description string `firestore:"description" json:"description" validate:"required"`
	Done        bool   `firestore:"done" json:"done"`
}
```

### 2\. Register the Model

Register your model with a Firestore collection name. This ensures the model is tied to the appropriate collection:

```go
instance, err := firegorm.RegisterModel(&Task{}, "tasks")
if err != nil {
	log.Fatalf("Failed to register model: %v", err)
}

task := instance.(*Task) // Cast the registered instance
```

### 3\. Perform CRUD Operations

#### Create a Document

```go
ctx := context.Background()
taskData := &Task{
	Title:       "Buy Groceries",
	Description: "Milk, Eggs, Bread, Butter",
	Done:        false,
}
if err := task.Create(ctx, taskData); err != nil {
	log.Fatalf("Failed to create task: %v", err)
}
log.Printf("Task created with ID: %s", taskData.ID)
```

#### Fetch a Document

```go
fetchedTask := &Task{}
if err := task.Get(ctx, taskData.ID, fetchedTask); err != nil {
	log.Fatalf("Failed to fetch task: %v", err)
}
log.Printf("Fetched Task: %+v", fetchedTask)
```

#### Update a Document

```go
updates := map[string]interface{}{
	"title":       "New Title",
	"description": "Updated description",
}
if err := task.Update(ctx, taskData.ID, updates); err != nil {
	log.Fatalf("Failed to update task: %v", err)
} else {
	log.Println("Task updated successfully.")
}
```

#### Soft Delete a Document

```go
if err := task.Delete(ctx, taskData.ID); err != nil {
	log.Fatalf("Failed to delete task: %v", err)
} else {
	log.Println("Task soft deleted.")
}
```

#### List Documents

```go
results := []*Task{}
nextPageToken, err := task.List(ctx, map[string]interface{}{
	"done": false,
}, 10, "", &results)
if err != nil {
	log.Fatalf("Failed to list tasks: %v", err)
}
log.Printf("Fetched Tasks: %+v", results)
log.Printf("Next Page Token: %s", nextPageToken)
```

---

## Logging

Firegorm uses a centralized logger that supports multiple log levels. Configure the logging level by setting the `FIREGORM_LOG_LEVEL` environment variable.

### Example

```console
export FIREGORM_LOG_LEVEL=DEBUG
```

---

## Soft Deletes

Soft deletes mark a document as deleted without removing it from the collection. This is achieved using the `Deleted` and `DeletedAt` fields in the `BaseModel`.

---

## Advanced Usage

### Custom Validation

Use the `validate` struct tag to enforce field requirements:

```go
type Task struct {
	Title string `firestore:"title" json:"title" validate:"required"`
}
```

Fields marked as `required` will throw an error if not set.

---

## Contributing

Contributions are welcome! Please open issues or submit pull requests to improve Firegorm.

---

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

---

Start using Firegorm today to simplify your Firestore interactions in Go! 🚀
