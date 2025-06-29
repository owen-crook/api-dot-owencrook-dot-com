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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/owen-crook/api-dot-owencrook-dot-com/internal/auth"
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
			CreatedAt:             time.Now(),
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
		document, err := GenerateGameScorecardDocumentFromText(c.Request.Context(), md.ID, user.Email, game, text, s)
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
