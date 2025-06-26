package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

func NewFirestoreClient(ctx context.Context, projectID, databaseID string) (*firestore.Client, error) {
	client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}
	return client, nil
}
