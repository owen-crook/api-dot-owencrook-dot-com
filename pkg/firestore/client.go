package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

func NewFirestoreClient(ctx context.Context, projectID, credentialsFile string) (*firestore.Client, error) {
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}
	return client, nil
}
