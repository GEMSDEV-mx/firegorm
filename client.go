package firegorm

import (
	"context"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var Client *firestore.Client

// Init initializes the Firestore client with the given credentials.
// The credentials should be passed as a JSON string.
func Init(credentialsJSON string) error {
	ctx := context.Background()
	sa := option.WithCredentialsJSON([]byte(credentialsJSON))

	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		return err // Return error instead of logging and exiting
	}

	Client, err = app.Firestore(ctx)
	if err != nil {
		return err // Return error instead of logging and exiting
	}

	return nil
}
