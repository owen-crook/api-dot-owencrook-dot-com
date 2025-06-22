// Purpose:
// Defines data structures and domain models related to users.
// Usually Go structs with tags for JSON and DB mapping.
// What to include:
// User struct with fields like ID, Name, Email, etc.
// Validation tags, serialization tags (e.g., json:"email" firestore:"email").
// Constants or enums related to user status/roles.

package boardgametracker

// Business Logic Models & Constants
type Game string

const (
	Wingspan Game = "wingspan"
	Wyrmspan Game = "wyrmspan"
)

// Storage Models
type ImageUploadMetadata struct {
	ID                    string  `firestore:"id" json:"id"`
	GoogleCloudStorageUrl string  `firestore:"google_cloud_storage_url" json:"google_cloud_storage_url"`
	LlmParsedContent      *string `firestore:"parsed_content" json:"parsed_content,omitempty"`
}

type GameScorecardDocument struct {
	ID                    string            `firestore:"id" json:"id"`
	ImageUploadMetadataID string            `firestore:"image_upload_metadata_id" json:"image_upload_metadata_id"`
	Game                  string            `firestore:"game" json:"game"`
	Date                  string            `firestore:"date" json:"date"`
	IsCompleted           bool              `firestore:"is_completed" json:"is_completed"`
	Location              *string           `firestore:"location" json:"location,omitempty"`
	PlayerScores          *[]map[string]any `firestore:"player_scores" json:"player_scores,omitempty"`
}
