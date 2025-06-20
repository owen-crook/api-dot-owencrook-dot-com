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
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleParseScoreCard(service *ScoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.ParseMultipartForm(10 << 20) // 10MB

		file, _, err := c.Request.FormFile("image")
		if err != nil {
			log.Printf("FormFile error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "image is required"})
			return
		}
		defer file.Close()

		// Read first 512 bytes to detect content type
		header := make([]byte, 512)
		n, err := file.Read(header)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read image header"})
			return
		}

		contentType := http.DetectContentType(header[:n])
		log.Printf("Detected content type: %s", contentType)

		// Rewind reader before full read
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rewind file"})
			return
		}

		imgBytes, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read image"})
			return
		}

		// attempt to save the file, in theory we get back a url
		url, err := service.Repository.SaveImage(c.Request.Context(), imgBytes, contentType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		// TODO: GetRawTextFromLlm

		// TODO: ParseTextToDataStructures

		// TODO: SaveDataStructuresAsDocuments

		// output, err := ParseScoreCard(c.Request.Context(), service, imgBytes)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }

		c.JSON(http.StatusOK, url)
	}
}
