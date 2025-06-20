// Purpose:
// Handles data persistence and retrieval from databases or external APIs.
// Abstracts the storage details from the rest of the app.
// What to include:
// Functions like FindUserByID, SaveUser, DeleteUser.
// Calls to Firestore (or your chosen DB client).
// Manages DB queries and error handling.

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
	ext := exts[0]
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

func (s *Storage) SaveScorecard(ctx context.Context, sc *Scorecard) error {
	_, err := s.FirestoreClient.Collection("scorecards").Doc(sc.ID).Set(ctx, sc)
	return err
}

// SavePlayerScoreEntry saves a PlayerScoreEntry to Firestore under the "player_scores" collection.
func (s *Storage) SavePlayerScoreEntry(ctx context.Context, pse *PlayerScoreEntry) error {
	_, err := s.FirestoreClient.Collection("player_scores").Doc(pse.Player+"_"+pse.GameID).Set(ctx, pse)
	return err
}

func GetTextFromPix2Struct(imageData []byte) (string, error) {
	return "", nil
}
