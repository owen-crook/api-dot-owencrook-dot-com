// Purpose:
// Defines data structures and domain models related to users.
// Usually Go structs with tags for JSON and DB mapping.
// What to include:
// User struct with fields like ID, Name, Email, etc.
// Validation tags, serialization tags (e.g., json:"email" firestore:"email").
// Constants or enums related to user status/roles.

package boardgametracker

const (
	GameWingspan         = "Wingspan"
	GameCatan            = "Catan"
	GameTerraformingMars = "TerraformingMars"
)

type Scorecard struct {
	ID      string                    `firestore:"id" json:"id"`
	Game    string                    `firestore:"game" json:"game"`
	Date    string                    `firestore:"date" json:"date"`
	Players map[string]map[string]int `firestore:"players" json:"players"`
}

type PlayerScoreEntry struct {
	Player string         `firestore:"player" json:"player"`
	Game   string         `firestore:"game" json:"game"`
	Date   string         `firestore:"date" json:"date"`
	GameID string         `firestore:"game_id" json:"game_id"`
	Scores map[string]int `firestore:"scores" json:"scores"`
}

var GameScoreSchemas = map[string][]string{
	GameWingspan:         {"printed", "eggs", "bonus", "total"},
	GameCatan:            {"roads", "cities", "longest_road", "total"},
	GameTerraformingMars: {"terraform", "milestones", "awards", "cards", "total"},
}
