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
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/gemini"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/helpers"
)

type ScoreService struct {
	Repository   *Storage
	GeminiClient *gemini.Client
}

func GetSupportedGames() []Game {
	return []Game{Wingspan, Wyrmspan}
}

func IsSupportedGame(game Game) bool {
	for _, supportedGame := range GetSupportedGames() {
		if supportedGame == game {
			return true
		}
	}
	return false
}

func GetScoringCategoriesByGame(game Game) ([]ScoringCategory, error) {
	if !IsSupportedGame(game) {
		return nil, fmt.Errorf("invalid game: %s", game)
	}
	switch game {
	case Wyrmspan:
		return WyrmspanScoringCategories, nil
	case Wingspan:
		return WingspanScoringCategories, nil
	default:
		return nil, fmt.Errorf("unable to find scoring categories for game: %s", game)
	}
}

func GetScorecardGeometryByGame(game Game) (string, error) {
	if !IsSupportedGame(game) {
		return "", fmt.Errorf("invalid game: %s", game)
	}
	switch game {
	case Wyrmspan, Wingspan:
		return "The rows of the scorecard represent score categories and the columns represent different players of that game. The values at each row and column represent the score a given player acheived for that category. ", nil
	default:
		return "", fmt.Errorf("unable to find scorecard geometry for game: %s", game)
	}
}

func GetExampleJsonByGame(game Game) (map[string]interface{}, error) {
	if !IsSupportedGame(game) {
		return nil, fmt.Errorf("invalid game: %s", game)
	}
	switch game {
	case Wyrmspan:
		return map[string]interface{}{
			"date":     "2025-06-22",
			"location": "minty",
			"players": []map[string]interface{}{
				{
					"name":                    "SM",
					"markers-on-dragon-guild": 3,
					"tails-on-dragons":        32,
					"end-game-abilities":      13,
					"eggs":                    12,
					"cached-resources":        1,
					"tucked-cards":            5,
					"public-objectives":       15,
					"remaining-coins-items":   0,
					"total":                   81,
				},
				{
					"name":                    "WS",
					"markers-on-dragon-guild": 7,
					"tails-on-dragons":        24,
					"end-game-abilities":      16,
					"eggs":                    12,
					"cached-resources":        7,
					"tucked-cards":            6,
					"public-objectives":       9,
					"remaining-coins-items":   1,
					"total":                   82,
				},
			},
		}, nil
	case Wingspan:
		return map[string]interface{}{
			"date":     "2025-06-23",
			"location": "crook nook",
			"players": []map[string]interface{}{
				{
					"name":               "OC",
					"birds":              31,
					"bonus-cards":        10,
					"end-of-round-goals": 18,
					"eggs":               18,
					"food-on-cards":      5,
					"tucked-cards":       2,
					"total":              84,
				},
				{
					"name":               "JB",
					"birds":              49,
					"bonus-cards":        4,
					"end-of-round-goals": 17,
					"eggs":               9,
					"food-on-cards":      4,
					"tucked-cards":       7,
					"total":              90,
				},
				{
					"name":               "MS",
					"birds":              34,
					"bonus-cards":        11,
					"end-of-round-goals": 13,
					"eggs":               6,
					"food-on-cards":      2,
					"tucked-cards":       1,
					"total":              67,
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unable to find example json for game: %s", game)
	}
}

func GetTextFromLLM(ctx context.Context, service *ScoreService, game Game, image []byte) (string, error) {
	// grab standard prompt elements from critical functions to support
	// dynamic generate of the prompt
	categories, err := GetScoringCategoriesByGame(game)
	if err != nil {
		return "", err
	}

	scorecardGeometry, err := GetScorecardGeometryByGame(game)
	if err != nil {
		return "", err
	}

	exampleJson, err := GetExampleJsonByGame(game)
	if err != nil {
		return "", err
	}
	exampleJsonBytes, err := json.MarshalIndent(exampleJson, "", "  ")
	if err != nil {
		return "", err
	}

	var promptBuilder strings.Builder
	fmt.Fprintf(&promptBuilder, "This image contains a boardgame scorecard for the game %s. ", string(game))
	promptBuilder.WriteString(scorecardGeometry)
	promptBuilder.WriteString("The score categories are listed below in the form 'long-form :: short-form':\n")
	for _, category := range categories {
		fmt.Fprintf(&promptBuilder, "- %s :: %s\n", category.LongName, category.ShortName)
	}
	promptBuilder.WriteString("\nYour job is to generate a plain JSON object containing information about the game as follows:\n")
	promptBuilder.WriteString("1. if a date is written anywhere on the scorecard (typically outside the primary scoring area), the value should be stored in the json key 'date'. if not visible, please leave the 'date' key with a null value.\n")
	promptBuilder.WriteString("2. if a location is written anywhere on the scorecard (typically outside the primary scoring area), the value should be stored in the json key 'location' if not visible, please leave the 'location' key with a null value.\n")
	promptBuilder.WriteString("3. include a 'players' key that contains an array of the scores from each player. Each object in the array should contain the players name under the 'name' key, a key for each of the short-form categories above, and 'total' key for their overall score.\n")
	promptBuilder.WriteString("\nHere is an example of the expected response:\n")
	promptBuilder.WriteString("```json\n")
	promptBuilder.WriteString(string(exampleJsonBytes))
	promptBuilder.WriteString("\n```")
	prompt := promptBuilder.String()

	text, err := service.GeminiClient.GenerateFromTextAndImage(ctx, prompt, image)
	if err != nil {
		return "", fmt.Errorf("failed to generate text from image: %w", err)
	}
	return text, nil
}

func GenerateGameScorecardDocumentFromText(ctx context.Context, imageUploadMetadataId, creator, game, text string, submittedDate time.Time, service *ScoreService) (*GameScorecardDocument, error) {
	// initialize final vars
	var id string
	var finalDate time.Time
	var parsedDate time.Time
	var location string
	var playerScores []map[string]any

	// initialize bools for checks
	foundPlayerScores := false
	allItemsInPlayerScoresValid := false

	// set default values for required fields
	id = uuid.New().String()
	location = "unknown"
	finalDate = submittedDate

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
	err := json.Unmarshal([]byte(likelyJsonString), &parsedJson)
	if err != nil {
		// TODO: return partially completed scorecard instead of error
		return nil, fmt.Errorf("unable to parse potential JSON string until actual JSON")
	}

	// at this point, we have a usable struct, we just need to validate its contents
	// we expect the keys date, location, and players
	// date -> key should always exist, but value can be null
	// location -> key should always exist, but value can be null
	// players -> should always exist, if missing or invalid, we mark as incomplete
	parsedDateVal, ok := parsedJson["date"]
	if ok {
		if parsedDateVal != nil {
			parsedDateStr, ok := parsedDateVal.(string)
			if ok {
				log.Printf("found datestring %s", parsedDateStr)
				parsedDate, err = helpers.ParseFlexibleDate(parsedDateStr)
				if err != nil {
					log.Printf("unable to parse string to date for %s", parsedDateStr)
				} else {
					finalDate = helpers.TimeAsCalendarDateOnly(parsedDate)
				}
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
					// TODO [blocking]: there will be actual expect structs to convert these
					//       to based on the game that we are playing. need to
					//       do another layer of checks once we write the function
					// TODO [maybe]: instead of making this a list, key it by the name of the
					//       				 the player. need to consider how we want to query/render
					//							 the final data before any decisions are made
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
		Date:                  finalDate,
		IsCompleted:           foundPlayerScores && allItemsInPlayerScoresValid,
		Location:              &location,
		PlayerScores:          &playerScores,
		CreatedBy:             &creator,
		CreatedAt:             time.Now().In(time.UTC),
	}, nil
}
