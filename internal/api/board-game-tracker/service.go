// Purpose:
// Implements business logic of your application.
// Validates, processes data, and orchestrates calls to repository and other services.
// Does not handle HTTP or database directly.
// What to include:
// Functions like GetUserByID, CreateUser with core app rules.
// Input/output transformations if needed.
// Calls repository layer to get/save data.

package boardgametracker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/gemini"
)

type ScoreService struct {
	Repository   *Storage
	GeminiClient *gemini.Client
}

func ValidateScorecard(sc Scorecard) error {
	fields, ok := GameScoreSchemas[sc.Game]
	if !ok {
		return fmt.Errorf("unsupported game type: %s", sc.Game)
	}

	for player, scores := range sc.Players {
		for _, f := range fields {
			if _, ok := scores[f]; !ok {
				return fmt.Errorf("player %s is missing required field %q", player, f)
			}
		}
	}
	return nil
}

func ParseScoreCard(ctx context.Context, service *ScoreService, img []byte) (*Scorecard, error) {
	if service == nil || service.GeminiClient == nil {
		return nil, fmt.Errorf("HuggingFace client not available")
	}

	// game := "wyrmspan"
	// categories := []string{
	// 	"markers-on-dragon-guild",
	// 	"tails-on-dragons",
	// 	"end-game-abilities",
	// 	"eggs",
	// 	"cached-resources",
	// 	"tucked-cards",
	// 	"public-objectives",
	// 	"remaining-coins-items",
	// }

	// TODO: this prompt is pretty good
	// just generate it more dynamically based on which game we have (map full name to short name)
	// and then render the example from an actual struct
	p := `
	This image contains a boardgame scorecard for the game Wyrmspan.
	The rows of the scorecard represent score categories and the columns represent different players of that game.
	The values at each row and column represent the score a given player acheived for that category.
	The score categories are listed below (long-form :: short-form):
	- printed on dragons :: tails-on-dragons
	- from end-game abilities :: end-game-abilities
	- per egg* :: eggs
	- per cached resourced :: cached-resources
	- from public objectives (ties are friend-see rulebook p.15) :: public-objectives
	- from remaining coins * items 1 per coin 1 per 4 food, dragon card, cave card (in any combination) (round down) :: remaining-coins-items

	Your job is to generate a plain JSON object containing information about the game as follows:
	1. if a date is written anywhere on the scorecard (typically outside the primary scoring area), the value should be stored in the json key "date". if not visible, please leave the "date" key with a null value.
	2. if a location is written anywhere on the scorecard (typically outside the primary scoring area), the value should be stored in the json key "location" if not visible, please leave the "location" key with a null value.
	3. include a "players" key that contains an array of the scores from each player. make sure the players name is included, as well as the scores associated with the short form categories described above.

	Here is an example of the expected response:
	` + "```json" + `
	{
		"date": "2025-06-18",
		"location": "minty",
		"players": [
			{
				"name": "OC",
				"markers-on-dragon-guild": 3,
				"tails-on-dragons": 32,
				"end-game-abilities": 13,
				"eggs": 12,
				"cached-resources": 1,
				"tucked-cards": 5,
				"public-objectives": 15,
				"remaining-coins-items": 0,
				"total": 81
			},
			{
				"name": "RH"
				"markers-on-dragon-guild": 6,
				"tails-on-dragons": 45,
				"end-game-abilities": 3,
				"eggs": 7,
				"cached-resources": 6,
				"tucked-cards": 2,
				"public-objectives": 12,
				"remaining-coins-items": 0,
				"total": 81
			}
		]
	}
	` + "```"
	fmt.Printf("prompt:\n %s", p)

	text, err := service.GeminiClient.GenerateFromTextAndImage(ctx, p, img)
	if err != nil {
		return nil, fmt.Errorf("failed to generate text from image: %w", err)
	}

	fmt.Printf("\nreponse:\n %s", text)

	// TODO: go from the text returned into meaningful datastructures :)
	// lets go :)

	// Parse the model's response as JSON into a Scorecard
	var sc Scorecard
	if err := json.Unmarshal([]byte(text), &sc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal model output: %w", err)
	}

	// Validate the scorecard
	if err := ValidateScorecard(sc); err != nil {
		return nil, fmt.Errorf("scorecard validation failed: %w", err)
	}

	// Save to Firestore (assumes Store has SaveScorecard method)
	if service.Repository != nil {
		if err := service.Repository.SaveScorecard(ctx, &sc); err != nil {
			return nil, fmt.Errorf("failed to save scorecard: %w", err)
		}
	}

	// Return the parsed scorecard
	return &sc, nil
}
