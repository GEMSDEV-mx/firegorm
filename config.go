package firegorm

import (
	"context"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var Client *firestore.Client

// Init initializes the Firestore client and logger.
// The credentials should be passed as a JSON string.
func Init(credentialsJSON string) error {
	InitializeLogger() // Set up logging

	ctx := context.Background()
	sa := option.WithCredentialsJSON([]byte(credentialsJSON))

	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		Log(ERROR, "Failed to initialize Firebase App: %v", err)
		return err
	}

	Client, err = app.Firestore(ctx)
	if err != nil {
		Log(ERROR, "Failed to initialize Firestore client: %v", err)
		return err
	}

	Log(INFO, "Firestore client successfully initialized")
	return nil
}
