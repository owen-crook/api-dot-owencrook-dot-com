// Purpose:
// Registers all HTTP routes related to the user API.
// Keeps route definitions separate for clarity and modularity.
// What to include:
// A function like RegisterRoutes(*gin.RouterGroup) that attaches handlers to endpoints.
// Grouping routes under /user or similar path.

package boardgametracker

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, service *ScoreService) {
	bgt := rg.Group("/board-game-tracker")
	bgt.POST("/parse-score-card/:game", HandleParseScoreCard(service))
}
