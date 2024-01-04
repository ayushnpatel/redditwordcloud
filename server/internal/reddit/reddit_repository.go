package reddit

import (
	"context"
	"redditwordcloud/internal/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoDBClient interface {
	Connect()
	Disconnect()
}

type repository struct {
	client          *mongo.Client
	wordsCollection *mongo.Collection
}

const (
	MongoDBConnectionString = "MONGODB_CONNECTION_STRING"
	WordsCollectionName     = "WORDS_COLLECTION_NAME"
	DatabaseName            = "DATABASE_NAME"
)

func NewRepository() Repository {
	var mongoUri = config.GetEnv(MongoDBConnectionString, "")

	if mongoUri == "" {
		return nil
	}

	zap.S().Info("Connecting to MongoDB...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		mongoUri,
	))

	err = client.Ping(ctx, nil)

	if err != nil {
		zap.S().Error("There was a problem connecting to your Atlas cluster. Check that the URI includes a valid username and password, and that your IP address has been added to the access list. Error: ", err)
		panic(err)
	}

	dbName, collectionName := config.GetEnv(DatabaseName, ""), config.GetEnv(WordsCollectionName, "")

	collection := client.Database(dbName).Collection(collectionName)

	zap.S().Info("Connected to MongoDB!")
	return &repository{
		client:          client,
		wordsCollection: collection,
	}
}

func (r *repository) InsertWords(ctx context.Context, words *map[string]int, scid string) (*WordDocument, error) {

	wordDoc := WordDocument{
		ID:                    primitive.NewObjectID(),
		SubredditAndCommentId: scid,
		Words:                 *words,
		LastUpdated:           primitive.NewDateTimeFromTime(time.Now()),
	}

	insertResult, err := r.wordsCollection.InsertOne(context.TODO(), wordDoc)

	if err != nil {
		zap.S().Errorf("Failed to insert word document into collection with scid %s: %w", scid, err)
		return nil, err
	}

	zap.S().Infof("%s document successfully inserted.", insertResult.InsertedID)

	return nil, nil
}

func (r *repository) GetWordsFromLink(ctx context.Context, scid string) (*WordDocument, error) {

	filter := bson.D{{Key: "scid", Value: scid}}

	var result WordDocument
	err := r.wordsCollection.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		zap.S().Errorf("Error getting WordDocument from MongoDb: %w", err)
		return nil, err
	}

	return &result, nil
}

func (r *repository) Disconnect() {
	zap.S().Info("Disconnecting MongoDB client....")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r.client.Disconnect(ctx)

	zap.S().Info("Disconnected MongoDB client.")
}
