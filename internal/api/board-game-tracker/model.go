// Purpose:
// Defines data structures and domain models related to users.
// Usually Go structs with tags for JSON and DB mapping.
// What to include:
// User struct with fields like ID, Name, Email, etc.
// Validation tags, serialization tags (e.g., json:"email" firestore:"email").
// Constants or enums related to user status/roles.

package boardgametracker

import "time"

// Business Logic Models & Constants
type Game string

const (
	Wingspan Game = "wingspan"
	Wyrmspan Game = "wyrmspan"
)

type ScoringCategory struct {
	ShortName string
	LongName  string
}

var (
	WyrmspanScoringCategories []ScoringCategory = []ScoringCategory{
		{ShortName: "guild", LongName: "markers on the dragon guild"},
		{ShortName: "tails-on-dragons", LongName: "printed on dragons"},
		{ShortName: "end-game-abilities", LongName: "from end-game abilities"},
		{ShortName: "eggs", LongName: "per egg*"},
		{ShortName: "cached-resources", LongName: "per cached resource*"},
		{ShortName: "public-objectives", LongName: "from public objectives (ties are friendly--see rulebook p.15)"},
		{ShortName: "remaining-coins-items", LongName: "from remaining coins & items 1 per coin 1 per 4 food, dragon card, cave card (in any combination) (round down)"},
	}

	WingspanScoringCategories []ScoringCategory = []ScoringCategory{
		{ShortName: "birds", LongName: "birds"},
		{ShortName: "bonus-cards", LongName: "bonus cards"},
		{ShortName: "end-of-round-goals", LongName: "end-of-round goals"},
		{ShortName: "eggs", LongName: "eggs"},
		{ShortName: "food-on-cards", LongName: "food on cards"},
		{ShortName: "tucked-cards", LongName: "tucked cards"},
	}
)

// Storage Models
type ImageUploadMetadata struct {
	ID                    string    `firestore:"id" json:"id"`
	GoogleCloudStorageUrl string    `firestore:"google_cloud_storage_url" json:"google_cloud_storage_url"`
	LlmParsedContent      *string   `firestore:"parsed_content" json:"parsed_content,omitempty"`
	CreatedBy             *string   `firestore:"created_by" json:"created_by,omitempty"`
	CreatedAt             time.Time `firestore:"created_at" json:"created_at"`
}

type GameScorecardDocument struct {
	ID                    string            `firestore:"id" json:"id"`
	ImageUploadMetadataID string            `firestore:"image_upload_metadata_id" json:"image_upload_metadata_id"`
	Game                  string            `firestore:"game" json:"game"`
	Date                  time.Time         `firestore:"date" json:"date"`
	IsCompleted           bool              `firestore:"is_completed" json:"is_completed"`
	Location              *string           `firestore:"location" json:"location,omitempty"`
	PlayerScores          *[]map[string]any `firestore:"player_scores" json:"player_scores,omitempty"`
	CreatedBy             *string           `firestore:"created_by" json:"created_by,omitempty"`
	CreatedAt             time.Time         `firestore:"created_at" json:"created_at"`
}

type GameScorecardDocumentUpdate struct {
	ID           string            `firestore:"id" json:"id"`
	Game         *string           `firestore:"game,omitempty" json:"game,omitempty"`
	Date         *time.Time        `firestore:"date,omitempty" json:"date,omitempty"`
	IsCompleted  *bool             `firestore:"is_completed,omitempty" json:"is_completed,omitempty"`
	Location     *string           `firestore:"location,omitempty" json:"location,omitempty"`
	PlayerScores *[]map[string]any `firestore:"player_scores,omitempty" json:"player_scores,omitempty"`
	UpdatedBy    string            `firestore:"updated_by,omitempty" json:"updated_by,omitempty"`
	UpdatedAt    time.Time         `firestore:"updated_at,omitempty" json:"updated_at,omitempty"`
}
