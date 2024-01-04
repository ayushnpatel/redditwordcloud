package main

import (
	config "redditwordcloud/internal/config"
	"redditwordcloud/internal/health"
	"redditwordcloud/internal/reddit"
	"redditwordcloud/router"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	cfg := config.Load()

	if cfg.Env == config.Local || cfg.Env == config.Dev {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zap.ReplaceGlobals(zap.Must(config.Build()))
	}

	if cfg.Env == config.Production {
		config := zap.NewProductionConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zap.ReplaceGlobals(zap.Must(config.Build()))
	}

}

func main() {
	// dbConn, err := db.NewDatabase()
	// if err != nil {
	// 	log.Fatalf("could not initialize database connection: %s", err)
	// }

	zap.S().Info("Initializing config...")

	redditRep := reddit.NewRepository()
	defer redditRep.Disconnect()

	redditSvc := reddit.NewService(redditRep)
	redditHandler := reddit.NewHandler(redditSvc)

	healthHandler := health.NewHandler()

	router.InitRouter(healthHandler, redditHandler)
	router.Start("0.0.0.0:8080")
}
