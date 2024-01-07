package mongodb

import (
	"context"
	"time"

	"github.com/newrelic/go-agent/v3/integrations/nrmongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoDBConfig struct {
	ConnectionString string `env:"MONGODB_CONNECTION_STRING,required"`
	CollectionName   string `env:"WORDS_COLLECTION_NAME,required"`
	DatabaseName     string `env:"DATABASE_NAME,required"`
}

type MongoDBClient struct {
	Config MongoDBConfig
	Client *mongo.Client
}

func New(cfg MongoDBConfig) *MongoDBClient {

	zap.S().Info("Connecting to MongoDB...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nrMon := nrmongo.NewCommandMonitor(nil)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		cfg.ConnectionString,
	).SetMonitor(nrMon))

	if err != nil {
		zap.S().Errorf("could not instantiate mongodb client: %w", err)
		panic(err)
	}

	err = client.Ping(ctx, nil)

	if err != nil {
		zap.S().Error("There was a problem connecting to your Atlas cluster. Check that the URI includes a valid username and password, and that your IP address has been added to the access list. Error: ", err)
		panic(err)
	}

	zap.S().Info("Connected to MongoDB!")

	return &MongoDBClient{
		cfg,
		client,
	}
}

func (mdbc *MongoDBClient) Disconnect() {
	zap.S().Info("Disconnecting MongoDB client....")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mdbc.Client.Disconnect(ctx)

	zap.S().Info("Disconnected MongoDB client.")
}
