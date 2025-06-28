package router

import (
	"context"
	"log"

	boardgametracker "github.com/owen-crook/api-dot-owencrook-dot-com/internal/api/board-game-tracker"
	"github.com/owen-crook/api-dot-owencrook-dot-com/internal/auth"
	"github.com/owen-crook/api-dot-owencrook-dot-com/internal/config"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/firestore"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/gcs"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/gemini"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	ctx := context.Background()

	// initialize auth
	if err := auth.Init(ctx, cfg.GoogleClientID); err != nil {
		log.Fatalf("auth init failed: %v", err)
	}

	r := gin.Default()

	// health check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// setup route groups
	v1RouteGroup := r.Group("/api/v1")

	firestoreClient, err := firestore.NewFirestoreClient(ctx, cfg.GCPProjectID, cfg.FirestoreDatabaseID)
	if err != nil {
		log.Fatalf("failed to initialize Firestore: %v", err)
	}

	googleCloudStorageClient, err := gcs.NewGCSClient(ctx, "owencrook-dot-com")
	if err != nil {
		log.Fatalf("failed to initialize GCS: %v", err)
	}

	log.Println("Creating board game tracker repository")
	bgtRepository := boardgametracker.NewStorage(firestoreClient, googleCloudStorageClient)
	if bgtRepository == nil {
		log.Fatal("bgtRepository is nil!")
	}

	geminiClient, err := gemini.NewClient(ctx, cfg.GeminiToken, "gemini-2.0-flash")
	if err != nil {
		log.Fatal("gemini client is nil!")
	}

	bgtService := &boardgametracker.ScoreService{
		Repository:   bgtRepository,
		GeminiClient: geminiClient,
	}
	log.Println("Registering boardgametracker routes")
	boardgametracker.RegisterRoutes(cfg, v1RouteGroup, bgtService)
	return r
}
