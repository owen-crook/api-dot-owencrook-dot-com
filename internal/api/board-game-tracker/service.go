// Purpose:
// Implements business logic of application.
// Validates, processes data, and orchestrates calls to repository and other services.
// Does not handle HTTP or database directly.

package boardgametracker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/gemini"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/helpers"
	"github.com/owen-crook/board-game-tracker-go-common/pkg/documents"
	"github.com/owen-crook/board-game-tracker-go-common/pkg/gamedata"
	"github.com/owen-crook/board-game-tracker-go-common/pkg/games"
)

type ScoreService struct {
	Repository   *Storage
	GeminiClient *gemini.Client
}

func GetTextFromLLM(ctx context.Context, service *ScoreService, game games.Game, image []byte) (string, error) {
	// grab standard prompt elements from critical functions to support
	// dynamic generate of the prompt
	categories, err := gamedata.GetScoringCategoriesByGame(game)
	if err != nil {
		return "", err
	}

	scorecardGeometry, err := gamedata.GetScorecardGeometryByGame(game)
	if err != nil {
		return "", err
	}

	exampleJson, err := gamedata.GetExampleJsonByGame(game)
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

func GenerateGameScorecardDocumentFromText(ctx context.Context, imageUploadMetadataId, game, text string, submittedDate time.Time, service *ScoreService) (*documents.ScorecardDocumentRaw, error) {
	// initialize final vars
	var finalDate time.Time
	var parsedDate time.Time
	var location string
	var playerScores []map[string]any

	// initialize bools for checks
	foundPlayerScores := false
	allItemsInPlayerScoresValid := false

	// set default values for required fields
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

	// determine expected columns on the players scores
	categories, err := gamedata.GetScoringCategoriesByGame(games.Game(game))
	if err != nil {
		return nil, err
	}
	var expectedKeys []string
	for _, category := range categories {
		expectedKeys = append(expectedKeys, category.ShortName)
	}
	expectedKeys = append(expectedKeys, "total")
	expectedKeys = append(expectedKeys, "name")

	// parse the player objects
	playersVal, ok := parsedJson["players"]
	if ok {
		playersSlice, ok := playersVal.([]any)
		if ok {
			foundPlayerScores = true
			validItems := 0
			for _, potentialPlayer := range playersSlice {
				player, ok := potentialPlayer.(map[string]any)
				if ok {
					// initialize a new map that will only contain expected keys
					// and our tracking slices
					playerClean := make(map[string]any)
					playerIsValid := true
					var missing []string
					var extra []string

					// convert expected to map for fast lookup
					expectedSet := make(map[string]bool)
					for _, key := range expectedKeys {
						expectedSet[key] = true
					}

					// find extra keys and build clean object
					for key, value := range player {
						if expectedSet[key] {
							if key == "name" { // name should be a string
								valueStr, ok := value.(string)
								if !ok {
									log.Printf("name is not string, using unknown")
									playerClean[key] = "unknown"
								} else {
									playerClean[key] = strings.ToLower(valueStr)
								}
							} else { // everything else should be an integer
								floatVal, ok := value.(float64)
								if !ok {
									log.Printf("score for key %s is not int, using 0", key)
									playerClean[key] = 0
								} else {
									if float64(int(floatVal)) == floatVal {
										playerClean[key] = int(floatVal)
									} else {
										log.Printf("score for key %s is float, not int, rounding", key)
										playerClean[key] = int(math.Round(floatVal))
									}
								}
							}
						} else {
							extra = append(extra, key)
						}
					}

					// find missing keys
					for _, key := range expectedKeys {
						if _, exists := player[key]; !exists {
							missing = append(missing, key)
						}
					}

					if len(extra) > 0 {
						playerIsValid = false
						log.Printf("found the following extra keys: %v", extra)
					}
					if len(missing) > 0 {
						playerIsValid = false
						log.Printf("found the following missing keys: %v", missing)
					}

					if playerIsValid {
						validItems++
					}

					playerClean["id"] = uuid.New().String()
					playerScores = append(playerScores, playerClean)
				}
			}
			if validItems == len(playersSlice) {
				allItemsInPlayerScoresValid = true
			}
		}
	}

	return &documents.ScorecardDocumentRaw{
		ImageUploadMetadataID: imageUploadMetadataId,
		Game:                  game,
		Date:                  finalDate,
		IsCompleted:           foundPlayerScores && allItemsInPlayerScoresValid,
		Location:              &location,
		PlayerScores:          &playerScores,
	}, nil
}
