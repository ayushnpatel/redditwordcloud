package reddit

import (
	"context"

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
	Link  string `json:"link"`
	Words map[string]int
}

type Repository interface {
	InsertWords(ctx context.Context, words *map[string]int, link string) (*WordDocument, error)
	GetWordsFromLink(ctx context.Context, link string) (*WordDocument, error)
	Disconnect()
}

type Service interface {
	GetRedditThreadWordsByThreadID(c context.Context, req *GetRedditThreadWordsByThreadIDReq) (*GetRedditThreadWordsRes, error)
	GetRedditThreadWordsByLink(c context.Context, req *GetRedditThreadWordsByLinkReq) (*GetRedditThreadWordsRes, error)
}
