// Purpose:
// Handles incoming HTTP requests and sends HTTP responses.
// Acts as the controller layer in MVC terms.
// What to include:
// Gin handler functions (func(c *gin.Context)) for each endpoint, e.g., GetUser, CreateUser.
// Input validation (request parameters, JSON bodies).
// Call methods from the service layer to execute business logic.
// Format responses and error handling.

package boardgametracker

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/owen-crook/api-dot-owencrook-dot-com/internal/auth"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/helpers"
)

func HandleParseScoreCard(s *ScoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// parse user making request (they should be authZ'd to get here, just want data for logging)
		user, err := auth.GetUserFromRequest(c.Request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse user from request"})
		}

		// parse and validate the game
		game := c.Param("game")
		if !IsSupportedGame(Game(game)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported game: %s", game)})
		}

		// parse image
		c.Request.ParseMultipartForm(10 << 20) // 10MB
		file, _, err := c.Request.FormFile("image")
		if err != nil {
			log.Printf("FormFile error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "image is required"})
			return
		}
		defer file.Close()

		// parse the date, falling back to today
		date := helpers.TimeAsCalendarDateOnly(time.Now())
		dateStrFromRequest := c.PostForm("date")
		if dateStrFromRequest != "" {
			parsedDate, err := helpers.ParseFlexibleDate(dateStrFromRequest)
			if err != nil {
				log.Printf("Unable to parse date from request: %v", err)
			} else {
				date = helpers.TimeAsCalendarDateOnly(parsedDate)
			}
		}

		// read first 512 bytes to detect content type
		header := make([]byte, 512)
		n, err := file.Read(header)
		if err != nil {
			log.Printf("Header error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read image header"})
			return
		}

		contentType := http.DetectContentType(header[:n])

		// Rewind reader before full read
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			log.Printf("File rewinder error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rewind file"})
			return
		}

		imgBytes, err := io.ReadAll(file)
		if err != nil {
			log.Printf("Image read error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read image"})
			return
		}

		// save the file to GCS, getting back the url
		url, err := s.Repository.SaveImage(c.Request.Context(), imgBytes, contentType)
		if err != nil {
			log.Printf("Image save error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// start building out metadata struct that we will upload no matter what
		md := ImageUploadMetadata{
			ID:                    uuid.New().String(),
			GoogleCloudStorageUrl: url,
			CreatedBy:             &user.Email,
			CreatedAt:             time.Now().In(time.UTC),
		}

		// send the request to Gemini
		text, err := GetTextFromLLM(c.Request.Context(), s, Game(game), imgBytes)
		if err != nil {
			log.Printf("LLM Text Parsing Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			md.LlmParsedContent = &text
		}

		// save the image upload metadata
		err = s.Repository.SaveImageUploadMetadata(c.Request.Context(), &md)
		if err != nil {
			log.Printf("Image upload metadata error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// TODO: if GetTextFromLLM fails, we may want to exit the flow after we save
		//       the metadata to prevent GenerateGameScorecardDocumentFromText from failing

		// parse the content from the string into known struct
		document, err := GenerateGameScorecardDocumentFromText(c.Request.Context(), md.ID, user.Email, game, text, date, s)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		// save the content to db
		err = s.Repository.SaveGameScorecardDocument(c.Request.Context(), document)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, document)
	}
}

func HandleUpdateScoreCard(s *ScoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// parse user making request (they should be authZ'd to get here, just want data for logging)
		user, err := auth.GetUserFromRequest(c.Request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse user from request"})
		}

		// parse and validate the documentId
		documentId := c.Param("documentId")
		exists, err := s.Repository.CheckDocumentExists(c.Request.Context(), "board-game-scorecards", documentId)
		if err != nil {
			log.Printf("Error checking document existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check document existence"})
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Document with ID %s not found", documentId)})
			return
		}

		// parse the request body into GameScorecardDocumentUpdate
		var update GameScorecardDocumentUpdate
		if err := c.ShouldBindJSON(&update); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		// handle updates
		var updates []firestore.Update
		if update.Game != nil {
			if !IsSupportedGame(Game(*update.Game)) {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported game: %s", *update.Game)})
				return
			}
			updates = append(updates, firestore.Update{Path: "game", Value: *update.Game})
		}
		if update.Date != nil {
			*update.Date = helpers.TimeAsCalendarDateOnly(*update.Date)
			updates = append(updates, firestore.Update{Path: "date", Value: *update.Date})
		}
		if update.IsCompleted != nil {
			updates = append(updates, firestore.Update{Path: "is_completed", Value: *update.IsCompleted})
		}
		if update.Location != nil {
			updates = append(updates, firestore.Update{Path: "location", Value: *update.Location})

		}
		if update.PlayerScores != nil {
			// TODO (OC-24): pass through validation to ensure we have all the players and that each player score
			// is valid, as if it was being passed through the LLM. To do this, I need to parse out
			// the logic from GenerateGameScorecardDocumentFromText and
			// put it into a function that can be called here.
			updates = append(updates, firestore.Update{Path: "player_scores", Value: *update.PlayerScores})
		}
		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no updates provided"})
			return
		}

		// add updated by and updated at
		updates = append(updates, firestore.Update{Path: "updated_by", Value: user.Email})
		updates = append(updates, firestore.Update{Path: "updated_at", Value: time.Now().In(time.UTC)})

		// perform the update
		_, err = s.Repository.FirestoreClient.Collection("board-game-scorecards").Doc(documentId).Update(c.Request.Context(), updates)
		if err != nil {
			log.Printf("Error updating document: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update scorecard"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Document updated successfully"})
	}
}
