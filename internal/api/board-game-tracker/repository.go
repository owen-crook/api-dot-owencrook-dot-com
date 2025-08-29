// Purpose:
// Handles data persistence and retrieval from databases or external APIs.
// Abstracts the storage details from the rest of the app.

package boardgametracker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime"

	firestore "cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/gcs"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/helpers"
	"github.com/owen-crook/board-game-tracker-go-common/pkg/documents"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const bgtBucket = "owencrook-dot-com"

type Storage struct {
	FirestoreClient *firestore.Client
	GCSClient       *gcs.Client
}

func NewStorage(fsClient *firestore.Client, gcsGlient *gcs.Client) *Storage {
	return &Storage{FirestoreClient: fsClient, GCSClient: gcsGlient}
}

func constructImagePath(imageId, ext string) string {
	return fmt.Sprintf("board-game-tracker/uploads/images/%s%s", imageId, ext)
}

func (s *Storage) SaveImage(ctx context.Context, image []byte, contentType string) (string, string, error) {
	// determine file extension
	exts, err := mime.ExtensionsByType(contentType)
	if err != nil {
		return "", "", err
	}
	if len(exts) == 0 {
		return "", "", errors.New("no file extension found for content type")
	}
	ext := helpers.NormalizeExtension(exts[0])
	// generate a uuid for the image to determine the path
	imageId := uuid.New().String()
	path := constructImagePath(imageId, ext)
	reader := bytes.NewReader(image)
	err = s.GCSClient.UploadFile(ctx, bgtBucket, path, reader, contentType)
	if err != nil {
		return "", "", err
	}
	return bgtBucket, path, nil
}

func (s *Storage) DeleteImage(ctx context.Context, path string) error {
	return s.GCSClient.DeleteFile(ctx, bgtBucket, path)
}

func (s *Storage) SaveImageUpload(ctx context.Context, metadata *documents.ImageUploadCreate) error {
	_, err := s.FirestoreClient.Collection("board-game-image-uploads").Doc(metadata.ID).Set(ctx, metadata)
	return err
}

func (s *Storage) SaveGameScorecardDocument(ctx context.Context, doc *documents.ScorecardDocumentCreate) error {
	_, err := s.FirestoreClient.Collection("board-game-scorecards").Doc(doc.ID).Set(ctx, doc)
	return err
}

func (s *Storage) GetDocument(ctx context.Context, collection, documentId string) (*firestore.DocumentSnapshot, error) {
	reference := s.FirestoreClient.Collection(collection).Doc(documentId)
	snapshot, err := reference.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return snapshot, nil
}

func (s *Storage) CheckDocumentExists(ctx context.Context, collection, documentId string) (bool, error) {
	snapshot, err := s.GetDocument(ctx, collection, documentId)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to find document: %w", err)
	}
	return snapshot.Exists(), nil
}

func (s *Storage) DeleteDocument(ctx context.Context, collection, documentId string) error {
	reference := s.FirestoreClient.Collection(collection).Doc(documentId)
	_, err := reference.Delete(ctx)
	return err
}
