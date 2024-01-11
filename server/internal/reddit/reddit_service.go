package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"redditwordcloud/pkg/retryhttp"
	"redditwordcloud/pkg/util"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	cmap "github.com/orcaman/concurrent-map/v2"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
)

type service struct {
	Repository
	timeout      time.Duration
	client       *http.Client
	redditClient *http.Client
	rl           ratelimit.Limiter
	rcfg         RedditConfig
}

const (
	defaultBaseURL         = "https://oauth.reddit.com"
	defaultBaseURLReadonly = "https://reddit.com"
	defaultTokenURL        = "https://www.reddit.com/api/v1/access_token"
	maxMoreChildrenLimit   = 100
	getCommentArticleLimit = 4
	redditRps              = 2
	NotFoundMessage        = "{\"message\": \"Not Found\", \"error\": 404}"
)

// Credentials are used to authenticate to make requests to the Reddit API.
type Credentials struct {
	ID       string
	Secret   string
	Username string
	Password string
}

func NewService(rcfg RedditConfig, repository Repository) Service {
	c := retryhttp.NewRetryableClient()
	oauthTransport := oauthTransport(c, rcfg)
	c.Transport = oauthTransport

	return &service{
		Repository:   repository,
		timeout:      time.Duration(2) * time.Second,
		client:       retryhttp.NewRetryableClient(),
		redditClient: c,
		rl:           ratelimit.New(redditRps),
		rcfg:         rcfg,
	}
}

func (svc *service) GetRedditThreadWordsByThreadID(c context.Context, req *GetRedditThreadWordsByThreadIDReq) (*GetRedditThreadWordsRes, error) {
	//TODO: Implement
	return nil, nil
}

type Link struct {
	Protocol   string
	DomainName string
	Subreddit  string
	CommentId  string
}

type RedditRepliesObject struct {
	Body    string         `json:"body"`
	Replies RedditResponse `json:"replies,omitempty"`
	Ups     int            `json:"ups"`
	Id      string         `json:"id"`
}

type RedditMoreObject struct {
	Children []string `json:"children"`
	ParentId string   `json:"parent_id"`
	Id       string   `json:"id"`
	Depth    int      `json:"depth"`
}

type RedditMoreChildrenObject struct {
	JSON struct {
		Errors []string `json:"errors"`
		Data   struct {
			Things []RedditResponse `json:"things"`
		} `json:"data"`
	} `json:"json"`
}

type RedditListingObject struct {
	Children []RedditResponse `json:"children,omitempty"`
}

type RedditResponse struct {
	Kind string       `json:"kind"`
	Data RedditObject `json:"data"`
}

type RedditObject interface{}

func (rr *RedditResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		Kind string          `json:"kind"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	rr.Kind = raw.Kind

	switch raw.Kind {
	case "Listing":
		var listing RedditListingObject
		if err := json.Unmarshal(raw.Data, &listing); err != nil {
			return err
		}

		rr.Data = &listing
	case "more":
		var more RedditMoreObject
		if err := json.Unmarshal(raw.Data, &more); err != nil {
			return err
		}
		rr.Data = &more
	default:
		var replies RedditRepliesObject
		if err := json.Unmarshal(raw.Data, &replies); err != nil {
			return err
		}
		rr.Data = &replies
	}

	return nil
}

func (rro *RedditRepliesObject) UnmarshalJSON(data []byte) error {
	// Define an auxiliary type to use for unmarshaling, to avoid recursion
	type RedditChildrenDataObjectNoReplies struct {
		Body string `json:"body"`
		Id   string `json:"id"`
		Ups  int    `json:"ups"`
	}

	type RedditChildrenDataObjectAux RedditRepliesObject
	var raw json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	rawString := string(raw)
	index := strings.Index(rawString, "\"replies\": {")

	if index != -1 {
		var aux RedditChildrenDataObjectAux
		if err := json.Unmarshal(data, &aux); err != nil {
			return err
		}
		rro.Body = aux.Body
		rro.Ups = aux.Ups
		rro.Id = aux.Id
		rro.Replies = aux.Replies
	} else {
		var auxNoReplies RedditChildrenDataObjectNoReplies
		if err := json.Unmarshal(data, &auxNoReplies); err != nil {
			return err
		}
		rro.Body = auxNoReplies.Body
		rro.Ups = auxNoReplies.Ups
		rro.Id = auxNoReplies.Id
		rro.Replies = RedditResponse{}

	}

	return nil
}

func (svc *service) upsertWords(c context.Context, words cmap.ConcurrentMap[string, int], link *Link) {
	scid := fmt.Sprintf("%s/comments/%s", link.Subreddit, link.CommentId)
	m := words.Items()

	if len(m) != 0 {
		if err := svc.Repository.Upsert(c, m, scid); err != nil {
			zap.S().Errorf("could not upsert words for linkId: %s\n", link.CommentId)
			return
		}
		zap.S().Debugf("Upserted Word Map with %d entries.", len(m))
	}
}

func (svc *service) processRedditResponse(c context.Context, redditResponse RedditResponse, words cmap.ConcurrentMap[string, int], link *Link) {

	if redditResponse.Data == nil || redditResponse.Kind == "t3" {
		return
	}

	var wg sync.WaitGroup

	switch redditResponse.Kind {
	case "Listing":
		children := redditResponse.Data.(*RedditListingObject).Children
		words := cmap.New[int]()
		for _, child := range children {
			svc.processRedditResponse(c, child, words, link)
		}
		svc.upsertWords(c, words, link)
	case "more":
		if len(redditResponse.Data.(*RedditMoreObject).Children) == 0 {
			wg.Add(1)
			words := cmap.New[int]()
			go func(words cmap.ConcurrentMap[string, int]) {
				defer wg.Done()
				parentIdNoPrefix := strings.Split(redditResponse.Data.(*RedditMoreObject).ParentId, "_")[1]
				redditResponses, _ := svc.getCommentArticleResp(parentIdNoPrefix, link)
				for _, resp := range redditResponses {
					svc.processRedditResponse(c, resp, words, link)
				}
			}(words)
			wg.Wait()
			return
		}

		childrenChunks := util.ChunkStringSlice(redditResponse.Data.(*RedditMoreObject).Children, maxMoreChildrenLimit)
		parentId := redditResponse.Data.(*RedditMoreObject).ParentId

		for _, chunk := range childrenChunks {
			wg.Add(1)
			go func(children []string, parentId string) {
				defer wg.Done()

				redditReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/morechildren", defaultBaseURL), nil)

				if err != nil {
					zap.S().Errorf("Could not create reddit request: ", err)
				}

				redditReq.Header.Set("User-Agent", "redditwordcloud/1.0")

				q := redditReq.URL.Query()

				q.Add("link_id", fmt.Sprintf("t3_%s", link.CommentId))
				q.Add("children", strings.Join(children, ","))
				q.Add("api_type", "json")

				redditReq.URL.RawQuery = q.Encode()
				svc.rl.Take()
				res, err := svc.redditClient.Do(redditReq)

				if err != nil {
					zap.S().Debugf("Error c.Do: %w, reprocessing reddit response...", err)
					svc.processRedditResponse(c, redditResponse, cmap.New[int](), link)
				}

				zap.S().Debugf("Successful GET request.")

				body, err := io.ReadAll(res.Body)

				if err != nil {
					zap.S().Error("Error reading the response body:", err)
					wg.Wait()
					return
				}

				var MoreChildrenAPIResponse RedditMoreChildrenObject

				err = json.Unmarshal(body, &MoreChildrenAPIResponse)

				if err != nil {
					zap.S().Error("Error unmarshaling res to JSON:", err)
					zap.S().Debug(string(body))
					wg.Wait()
					return
				}

				words := cmap.New[int]()
				for _, child := range MoreChildrenAPIResponse.JSON.Data.Things {
					svc.processRedditResponse(c, child, words, link)
				}
				svc.upsertWords(c, words, link)
			}(chunk, parentId)
		}
		wg.Wait()
		return
	default:
		repliesResp := redditResponse.Data.(*RedditRepliesObject).Replies
		svc.processRedditResponse(c, repliesResp, words, link)

		body := redditResponse.Data.(*RedditRepliesObject).Body
		htmlUnescapedBody := html.UnescapeString(body)
		cleanedBody := cleanBody(strconv.Quote(htmlUnescapedBody))

		for _, word := range strings.Split(cleanedBody, " ") {
			lowerCaseWord := strings.ToLower(word)
			cwc, b := words.Get(lowerCaseWord)
			if b {
				words.Set(lowerCaseWord, cwc+1)
			} else {
				words.Set(lowerCaseWord, 1)
			}
		}
	}
	wg.Wait()

}

func cleanBody(body string) string {
	re := regexp.MustCompile(`\\(.)`) // Matches any escaped character

	// Replace escaped characters with an empty string
	cleanedBody := re.ReplaceAllString(body, "")
	// Removes quotes at beginning and end of string
	cleanedBody = strings.Trim(cleanedBody, "\"")

	// zap.S().Debugf("Current cleaned body: %s", cleanedBody)

	re = regexp.MustCompile(`(.{0,1})([^\w\s'])(.{0,1})`)
	isFloat := regexp.MustCompile(`\d([^\w\s])\d`)

	parts := re.FindAllString(cleanedBody, -1)

	for _, part := range parts {
		isAFloat := isFloat.MatchString(part)
		if !isAFloat {
			newPart := re.ReplaceAllString(part, "$1$3")
			cleanedBody = strings.Replace(cleanedBody, part, newPart, 1)
			cleanedBody = strings.ReplaceAll(cleanedBody, "’", "'") // Replace curly apostrophe with straight apostrophe
			cleanedBody = strings.ReplaceAll(cleanedBody, "`", "'") // Replace backtick with straight apostrophe
		}
	}

	return cleanedBody
}

func createLink(link string) *Link {
	parts := strings.Split(link, "/")

	protocol := strings.Join([]string{parts[0], "/"}, "")
	domainName := parts[2]
	subreddit := strings.Join([]string{parts[3], "/", parts[4]}, "")
	commentId := parts[6]

	return &Link{
		Protocol:   protocol,
		DomainName: domainName,
		Subreddit:  subreddit,
		CommentId:  commentId,
	}
}

func (svc *service) GetRedditThreadWordsByLink(c context.Context, req *GetRedditThreadWordsByLinkReq, txn *newrelic.Transaction) (*GetRedditThreadWordsRes, error) {
	link := createLink(req.Link)
	linkStr := fmt.Sprintf("%s/%s/%s/comments/%s", link.Protocol, link.DomainName, link.Subreddit, link.CommentId)
	scid := fmt.Sprintf("%s/comments/%s", link.Subreddit, link.CommentId)

	svc.rl.Take()
	redditResponses, err := svc.getCommentArticleResp(link.CommentId, link)

	if len(redditResponses) == 0 || err != nil {
		if err != nil {
			return nil, fmt.Errorf("could not get comments for link: %s, err: %w", linkStr, err)
		}
	}

	zap.S().Debugf("Checking if scid %s exists in db...", scid)

	segment := txn.StartSegment(fmt.Sprintf("scid %s check", scid))
	if wordDocument, err := svc.Repository.GetWordsFromLink(c, scid); err == nil {
		if wordDocument != nil {
			if util.IsInLastWeek(wordDocument.LastUpdated.Time()) {
				zap.S().Debugf("Retrieved word document for %s from MongoDB Words Collection.", scid)
				return &GetRedditThreadWordsRes{Words: wordDocument.Words, Success: true, Link: scid}, nil
			}
		} else {
			zap.S().Debugf("Scid %s does not exist in db. Inserting with empty map.", scid)
			if _, err := svc.Repository.InsertWords(c, make(map[string]int), scid); err != nil {
				zap.S().Errorf("Could not insert empty map into MongoDB: %w", err)
				return nil, fmt.Errorf("could not insert empty map into MongoDB: %w", err)
			}
			zap.S().Debug("Created Word Map with 0 entries.")
		}
	}
	segment.End()

	defer txn.StartSegment(fmt.Sprintf("Process RedditResponses for article %s", scid))
	for _, rr := range redditResponses {
		go func(rr RedditResponse) {
			svc.processRedditResponse(c, rr, cmap.New[int](), link)
		}(rr)
	}

	return &GetRedditThreadWordsRes{Success: true, Words: nil, Link: scid}, nil
}

func (svc *service) getCommentArticleResp(commentId string, link *Link) ([]RedditResponse, error) {
	redditReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/comments/article", defaultBaseURL, link.Subreddit), nil)

	if err != nil {
		zap.S().Errorf("Could not create reddit request: ", err)
	}

	redditReq.Header.Set("User-Agent", "redditwordcloud/1.0")

	q := redditReq.URL.Query()

	q.Add("article", link.CommentId)
	q.Add("comment", commentId)

	redditReq.URL.RawQuery = q.Encode()
	svc.rl.Take()
	res, err := svc.redditClient.Do(redditReq)

	if err != nil {
		zap.S().Debugf("Error c.Do: %w, reprocessing reddit response...", err)
		svc.getCommentArticleResp(commentId, link)
	}

	zap.S().Debugf("Successful GET request.")

	body, err := io.ReadAll(res.Body)

	if err != nil {
		zap.S().Error("Error reading the response body:", err)
		return nil, err
	}

	if string(body) == NotFoundMessage {
		return nil, fmt.Errorf("could not find reddit thread %s", link.CommentId)
	}

	// zap.S().Debugf("Body: %s", body)

	var CommentArticleAPIResponse []RedditResponse

	err = json.Unmarshal(body, &CommentArticleAPIResponse)

	if err != nil {
		zap.S().Error("Error unmarshaling res to JSON:", err)
		zap.S().Debug(string(body))
		return nil, fmt.Errorf("error unmarshaling res to JSON: %w", err)
	}

	return CommentArticleAPIResponse, nil
}
