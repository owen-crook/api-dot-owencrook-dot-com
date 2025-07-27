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

type Storage struct {
	FirestoreClient *firestore.Client
	GCSClient       *gcs.Client
}

func NewStorage(fsClient *firestore.Client, gcsGlient *gcs.Client) *Storage {
	return &Storage{FirestoreClient: fsClient, GCSClient: gcsGlient}
}

func (s *Storage) SaveImage(ctx context.Context, image []byte, contentType string) (string, error) {
	// determine file extension
	exts, err := mime.ExtensionsByType(contentType)
	if err != nil {
		return "", err
	}
	if len(exts) == 0 {
		return "", errors.New("no file extension found for content type")
	}
	ext := helpers.NormalizeExtension(exts[0])
	// generate a uuid for the image to determine the path
	imageId := uuid.New().String()
	path := fmt.Sprintf("board-game-tracker/uploads/images/%s%s", imageId, ext)
	reader := bytes.NewReader(image)
	res, err := s.GCSClient.UploadFile(ctx, path, reader, contentType)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s *Storage) SaveImageUpload(ctx context.Context, metadata *documents.ImageUploadCreate) error {
	_, err := s.FirestoreClient.Collection("board-game-image-uploads").Doc(metadata.ID).Set(ctx, metadata)
	return err
}

func (s *Storage) SaveGameScorecardDocument(ctx context.Context, doc *documents.ScorecardDocumentCreate) error {
	_, err := s.FirestoreClient.Collection("board-game-scorecards").Doc(doc.ID).Set(ctx, doc)
	return err
}

func (s *Storage) CheckDocumentExists(ctx context.Context, collection, documentId string) (bool, error) {
	reference := s.FirestoreClient.Collection(collection).Doc(documentId)
	snapshot, err := reference.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to find document: %w", err)
	}
	return snapshot.Exists(), nil
}
