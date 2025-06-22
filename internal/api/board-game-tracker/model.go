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

type ScoringCategory struct {
	ShortName string
	LongName  string
}

var (
	WingspanScoringCategories []ScoringCategory = []ScoringCategory{
		{ShortName: "tails-on-dragons", LongName: "printed on dragons"},
		{ShortName: "end-game-abilities", LongName: "from end-game abilities"},
		{ShortName: "eggs", LongName: "per egg*"},
		{ShortName: "cached-resources", LongName: "per cached resources"},
		{ShortName: "public-objectives", LongName: "from public objectives (ties are friend-see rulebook p.15)"},
		{ShortName: "remaining-coins-items", LongName: "rom remaining coins * items 1 per coin 1 per 4 food, dragon card, cave card (in any combination) (round down)"},
	}

	WyrmspanScoringCategories []ScoringCategory = []ScoringCategory{
		{ShortName: "", LongName: ""},
		{ShortName: "", LongName: ""},
		{ShortName: "", LongName: ""},
		{ShortName: "", LongName: ""},
		{ShortName: "", LongName: ""},
	}
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
