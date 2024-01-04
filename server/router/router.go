package router

import (
	"redditwordcloud/internal/health"
	"redditwordcloud/internal/reddit"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var r *gin.Engine

const (
	HealthPath                     = "/health"
	GetRedditThreadWordsByLinkPath = "/reddit/words/link"
)

func InitRouter(healthHandler *health.Handler, redditHandler *reddit.Handler) {
	r = gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:3000"
		},
		MaxAge: 12 * time.Hour,
	}))

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("ValidateLink", reddit.ValidateLink)
	}

	r.GET(HealthPath, healthHandler.GetHealth)
	// r.GET(GetRedditThreadWordsByThreadIDPath, redditHandler.GetRedditThreadWordsByThreadIDHandler)
	r.GET(GetRedditThreadWordsByLinkPath, redditHandler.GetRedditThreadWordsByLinkHandler)
}

func Start(addr string) error {
	return r.Run(addr)
}
