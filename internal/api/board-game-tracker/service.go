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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/gemini"
)

type ScoreService struct {
	Repository   *Storage
	GeminiClient *gemini.Client
}

// func ValidateScorecard(sc Scorecard) error {
// 	fields, ok := GameScoreSchemas[sc.Game]
// 	if !ok {
// 		return fmt.Errorf("unsupported game type: %s", sc.Game)
// 	}

// 	for player, scores := range sc.Players {
// 		for _, f := range fields {
// 			if _, ok := scores[f]; !ok {
// 				return fmt.Errorf("player %s is missing required field %q", player, f)
// 			}
// 		}
// 	}
// 	return nil
// }

func GetTextFromLLM(ctx context.Context, service *ScoreService, image []byte) (string, error) {
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

	text, err := service.GeminiClient.GenerateFromTextAndImage(ctx, p, image)
	if err != nil {
		return "", fmt.Errorf("failed to generate text from image: %w", err)
	}
	return text, nil
}

func GenerateGameScorecardDocumentFromText(ctx context.Context, imageUploadMetadataId, game, text string, service *ScoreService) (*GameScorecardDocument, error) {
	// initialize final vars
	var id string
	var date string
	var location string
	var playerScores []map[string]any

	// initialize bools for checks
	foundPlayerScores := false
	allItemsInPlayerScoresValid := false

	// set default values for required fields
	id = uuid.New().String()
	location = "unknown"
	tzLocation, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return nil, err
	}
	date = time.Now().In(tzLocation).Format("2006-01-02")

	// the llm will likely reply with markdown formatting of ```json...```
	// we need to attempt to clean the string as best as possible prior
	// to marshalling into a struct
	text = strings.Trim(text, "`")
	text = strings.TrimSpace(text)

	jsonStart := strings.Index(text, "{")
	jsonEnd := strings.LastIndex(text, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd < jsonStart {
		// TODO: return partially completed scorecard instead of error
		return nil, fmt.Errorf("no valid JSON object found within LLM response")
	}

	likelyJsonString := text[jsonStart : jsonEnd+1]
	var parsedJson map[string]interface{}
	err = json.Unmarshal([]byte(likelyJsonString), &parsedJson)
	if err != nil {
		// TODO: return partially completed scorecard instead of error
		return nil, fmt.Errorf("unable to parse potential JSON string until actual JSON")
	}

	// at this point, we have a usable struct, we just need to validate its contents
	// we expect the keys date, location, and players
	// date -> allowed to not have a value, but we expect the key
	// location -> allowed to not have a value, but we expect the key
	// players -> should always exist, if missing or invalid, we mark as incomplete
	dateVal, ok := parsedJson["date"]
	if ok {
		if dateVal != nil {
			dateStr, ok := dateVal.(string)
			if ok {
				date = dateStr
			}
		}
	}

	locationVal, ok := parsedJson["location"]
	if ok {
		if locationVal != nil {
			locationStr, ok := locationVal.(string)
			if ok {
				location = locationStr
			}
		}
	}

	playersVal, ok := parsedJson["players"]
	if ok {
		playersSlice, ok := playersVal.([]any)
		if ok {
			foundPlayerScores = true
			validItems := 0
			for _, potentialPlayer := range playersSlice {
				player, ok := potentialPlayer.(map[string]any)
				if ok {
					validItems++
					playerScores = append(playerScores, player)
					// TODO: there will be actual expect structs to convert these
					//       to based on the game that we are playing. need to
					//       do another layer of checks once we write the function
					// TODO: instead of making this a list, key it by the name of the
					//       the player
				}
			}
			if validItems == len(playersSlice) {
				allItemsInPlayerScoresValid = true
			}
		}
	}

	return &GameScorecardDocument{
		ID:                    id,
		ImageUploadMetadataID: imageUploadMetadataId,
		Game:                  game,
		Date:                  date,
		IsCompleted:           foundPlayerScores && allItemsInPlayerScoresValid,
		Location:              &location,
		PlayerScores:          &playerScores,
	}, nil
}
