package main

import (
	"redditwordcloud/internal/config"
	"redditwordcloud/internal/health"
	"redditwordcloud/internal/mongodb"
	"redditwordcloud/internal/newrelic"
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

	zap.S().Info("Config and Logger initialized!")

}

func main() {
	// dbConn, err := db.NewDatabase()
	// if err != nil {
	// 	log.Fatalf("could not initialize database connection: %s", err)
	// }

	cfg := config.Load()
	nrc := newrelic.New(cfg.NewRelicConfig)
	mdbc := mongodb.New(cfg.MongoDBConfig)
	defer mdbc.Disconnect()

	redditRep := reddit.NewRepository(mdbc, nrc)
	redditSvc := reddit.NewService(cfg.RedditConfig, redditRep)
	redditHandler := reddit.NewHandler(redditSvc)

	healthHandler := health.NewHandler()

	router.InitRouter(healthHandler, redditHandler, nrc)
	router.Start("0.0.0.0:8080")
}
