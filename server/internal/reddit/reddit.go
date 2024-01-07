package reddit

import (
	"context"

	"github.com/newrelic/go-agent/v3/newrelic"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WordDocument struct {
	ID                    primitive.ObjectID `bson:"_id"`
	SubredditAndCommentId string             `bson:"scid,omitempty"`
	Words                 map[string]int     `bson:"words"`
	LastUpdated           primitive.DateTime `bson:"last_updated"`
}

type GetRedditThreadWordsByThreadIDReq struct {
	ThreadID string `json:"threadId"`
}

type GetRedditThreadWordsByLinkReq struct {
	Link string `json:"link" binding:"required,ValidateLink"`
}

type GetRedditThreadWordsRes struct {
	Link    string `json:"link"`
	Words   map[string]int
	Success bool
}

type Repository interface {
	InsertWords(ctx context.Context, words map[string]int, scid string) (*WordDocument, error)
	GetWordsFromLink(ctx context.Context, scid string) (*WordDocument, error)
	Upsert(ctx context.Context, words map[string]int, link string) error
}

type RedditConfig struct {
	ID       string `env:"CLIENT_ID,required"`
	Secret   string `env:"CLIENT_SECRET,required"`
	Username string `env:"USERNAME,required"`
	Password string `env:"PASSWORD,required"`
}

type Service interface {
	GetRedditThreadWordsByThreadID(c context.Context, req *GetRedditThreadWordsByThreadIDReq) (*GetRedditThreadWordsRes, error)
	GetRedditThreadWordsByLink(c context.Context, req *GetRedditThreadWordsByLinkReq, txn *newrelic.Transaction) (*GetRedditThreadWordsRes, error)
}
