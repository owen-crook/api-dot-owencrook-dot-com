package router

import (
	// "github.com/owen-crook/api-dot-owencrook-dot-com/internal/api/billing"
	// "github.com/owen-crook/api-dot-owencrook-dot-com/internal/api/user"

	"context"
	"log"

	boardgametracker "github.com/owen-crook/api-dot-owencrook-dot-com/internal/api/board-game-tracker"
	"github.com/owen-crook/api-dot-owencrook-dot-com/internal/config"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/firestore"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/gcs"
	"github.com/owen-crook/api-dot-owencrook-dot-com/pkg/gemini"

	// "github.com/owen-crook/api-dot-owencrook-dot-com/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	ctx := context.Background()
	r := gin.Default()

	// Global middleware
	// r.Use(middleware.Logger())

	// Health check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	firestoreClient, err := firestore.NewFirestoreClient(ctx, cfg.GCPProjectID, cfg.GCPCredentialsFile)
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

	// log.Println("Creating huggingface client")
	// hfClient := huggingface.NewClient(cfg.HuggingFaceToken)
	// if hfClient == nil {
	// 	log.Fatal("huggingface client is nil!")
	// }

	geminiClient, err := gemini.NewClient(ctx, cfg.GeminiToken, "gemini-2.0-flash")
	if err != nil {
		log.Fatal("gemini client is nil!")
	}

	api := r.Group("/api/v1")

	bgtService := &boardgametracker.ScoreService{
		Repository:   bgtRepository,
		GeminiClient: geminiClient,
	}
	log.Println("Registering boardgametracker routes")
	boardgametracker.RegisterRoutes(api, bgtService)
	return r
}
