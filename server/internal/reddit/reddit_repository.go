package reddit

import (
	"context"
	"fmt"
	"redditwordcloud/internal/mongodb"
	"redditwordcloud/internal/newrelic"
	"redditwordcloud/pkg/util"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type repository struct {
	upsertMu        sync.Mutex
	wordsCollection *mongo.Collection
	nrc             *newrelic.NewRelicClient
}

func NewRepository(mdbc *mongodb.MongoDBClient, nrc *newrelic.NewRelicClient) Repository {

	collection := mdbc.Client.Database(mdbc.Config.DatabaseName).Collection(mdbc.Config.CollectionName)

	return &repository{
		wordsCollection: collection,
		nrc:             nrc,
	}
}

func (r *repository) InsertWords(ctx context.Context, words map[string]int, scid string) (*WordDocument, error) {
	defer r.nrc.Client.StartTransaction(fmt.Sprintf("%s: InsertWords", scid)).End()
	wordDoc := WordDocument{
		ID:                    primitive.NewObjectID(),
		SubredditAndCommentId: scid,
		Words:                 words,
		LastUpdated:           primitive.NewDateTimeFromTime(time.Now()),
	}

	r.upsertMu.Lock()
	insertResult, err := r.wordsCollection.InsertOne(context.TODO(), wordDoc)
	r.upsertMu.Unlock()

	if err != nil {
		zap.S().Errorf("Failed to insert word document into collection with scid %s: %w", scid, err)
		return nil, err
	}

	zap.S().Infof("%s document successfully inserted.", insertResult.InsertedID)

	return nil, nil
}

func (r *repository) GetWordsFromLink(ctx context.Context, scid string) (*WordDocument, error) {
	defer r.nrc.Client.StartTransaction(fmt.Sprintf("%s: GetWordsFromLink", scid)).End()
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

func (r *repository) Upsert(ctx context.Context, words map[string]int, scid string) error {
	defer r.nrc.Client.StartTransaction(fmt.Sprintf("%s: Upsert", scid)).End()
	filter := bson.D{{Key: "scid", Value: scid}}

	r.upsertMu.Lock()
	wordDoc, err := r.GetWordsFromLink(ctx, scid)

	if err != nil {
		return err
	}

	if wordDoc != nil {
		wordDoc.Words = util.CombineMaps(wordDoc.Words, words)
	}

	update := bson.M{
		"$set": wordDoc,
	}

	if _, err = r.wordsCollection.UpdateOne(ctx, filter, update); err != nil {
		zap.S().Errorf("Error upserting WordDocument %s to MongoDb: %w", scid, err)
		return fmt.Errorf("could not upsert: %w", err)
	}
	r.upsertMu.Unlock()

	return nil
}
