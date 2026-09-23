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
	return InitWithContext(context.Background(), credentialsJSON)
}

// InitWithContext initializes the Firestore client using the supplied context.
func InitWithContext(ctx context.Context, credentialsJSON string) error {
	InitializeLogger() // Set up logging

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

// Close releases resources held by the Firestore client. It is safe to call
// when Firegorm has not been initialized.
func Close() error {
	if Client == nil {
		return nil
	}
	err := Client.Close()
	Client = nil
	return err
}
