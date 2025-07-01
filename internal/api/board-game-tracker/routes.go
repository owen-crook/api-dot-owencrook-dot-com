// Purpose:
// Registers all HTTP routes related to the user API.
// Keeps route definitions separate for clarity and modularity.
// What to include:
// A function like RegisterRoutes(*gin.RouterGroup) that attaches handlers to endpoints.
// Grouping routes under /user or similar path.

package boardgametracker

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/owen-crook/api-dot-owencrook-dot-com/internal/auth"
	"github.com/owen-crook/api-dot-owencrook-dot-com/internal/config"
)

func RegisterRoutes(cfg *config.Config, rg *gin.RouterGroup, service *ScoreService) {
	// setup different route groups for various auth levels
	boardGameTrackerGroup := rg.Group("/board-game-tracker")                                                           // public
	boardGameTrackerAuthNGroup := boardGameTrackerGroup.Group("/", auth.RequireAuth(nil))                              // authN
	boardGameTrackerAuthZAdminGroup := boardGameTrackerGroup.Group("/", auth.RequireAuth(config.GetAdminEmails(*cfg))) // authZ

	// mount authN groups
	// TODO: delete dummy route once actual routes are in place
	boardGameTrackerAuthNGroup.GET("/dummy", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "found a dummy", "dummy": "you"})
	})

	// mount admin routes
	boardGameTrackerAuthZAdminGroup.POST("/parse-score-card/:game", HandleParseScoreCard(service))
	boardGameTrackerAuthZAdminGroup.PATCH("/update-score-card/:documentId", HandleUpdateScoreCard(service))
}
