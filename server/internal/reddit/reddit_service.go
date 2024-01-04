package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"redditwordcloud/internal/config"
	"redditwordcloud/pkg/retryhttp"
	"redditwordcloud/pkg/util"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

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
	creds        *Credentials
}

const (
	defaultBaseURL         = "https://oauth.reddit.com"
	defaultBaseURLReadonly = "https://reddit.com"
	defaultTokenURL        = "https://www.reddit.com/api/v1/access_token"
	maxMoreChildrenLimit   = 100
	getCommentArticleLimit = 4
	rps                    = 2
)

// Credentials are used to authenticate to make requests to the Reddit API.
type Credentials struct {
	ID       string
	Secret   string
	Username string
	Password string
}

func NewService(repository Repository) Service {
	c := retryhttp.NewRetryableClient()

	creds := &Credentials{
		ID:       config.GetEnv("REDDIT_CLIENT_ID", ""),
		Secret:   config.GetEnv("REDDIT_CLIENT_SECRET", ""),
		Username: config.GetEnv("REDDIT_USERNAME", ""),
		Password: config.GetEnv("REDDIT_PASSWORD", ""),
	}

	oauthTransport := oauthTransport(c, creds)
	c.Transport = oauthTransport

	return &service{
		Repository:   repository,
		timeout:      time.Duration(2) * time.Second,
		client:       retryhttp.NewRetryableClient(),
		redditClient: c,
		rl:           ratelimit.New(rps),
		creds:        creds,
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
	Children []RedditResponse `json:"children"`
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

func (svc *service) processRedditResponse(c context.Context, redditResponse RedditResponse, words *cmap.ConcurrentMap[string, int], link *Link, depth int) {

	if redditResponse.Data == nil || redditResponse.Kind == "t3" {
		return
	}

	var wg sync.WaitGroup

	switch redditResponse.Kind {
	case "Listing":
		children := redditResponse.Data.(*RedditListingObject).Children
		for _, child := range children {
			wg.Add(1)
			go func(ch RedditResponse) {
				defer wg.Done()
				svc.processRedditResponse(c, ch, words, link, depth)
			}(child)
		}
	case "more":
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
					svc.processRedditResponse(c, redditResponse, words, link, depth)
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

				for _, child := range MoreChildrenAPIResponse.JSON.Data.Things {

					if strings.HasPrefix(parentId, "t3") {
						// Parent is the link, thus we are getting the extra comments of the thread.
						svc.processRedditResponse(c, child, words, link, depth)
					} else {
						// Parent is a comment, thus we are getting the extra comments of the comment thread.
						svc.processRedditResponse(c, child, words, link, depth+1)

					}
				}

			}(chunk, parentId)
		}
	default:
		repliesResp := redditResponse.Data.(*RedditRepliesObject).Replies
		commentId := redditResponse.Data.(*RedditRepliesObject).Id

		wg.Add(1)
		go func(ch RedditResponse) {
			defer wg.Done()
			svc.processRedditResponse(c, ch, words, link, depth+1)
		}(repliesResp)

		if depth == getCommentArticleLimit {
			redditResponses := svc.getCommentArticleResp(commentId, link)
			for _, resp := range redditResponses {
				wg.Add(1)
				go func(ch RedditResponse) {
					defer wg.Done()
					svc.processRedditResponse(c, ch, words, link, depth+1)
				}(resp)
			}
			wg.Wait()
			return
		}

		body := redditResponse.Data.(*RedditRepliesObject).Body
		// if strings.Contains(body, "ever have become billionaires.") {
		// 	zap.S().Debug(repliesResp)
		// }
		// if strings.Contains(body, "Number is going to balloon here soon ") {
		// 	zap.S().Debug(repliesResp)
		// }
		// if strings.Contains(body, "It took 20 years to go from 30m deals to 13") {
		// 	zap.S().Debug(repliesResp)
		// }
		// if strings.Contains(body, "There are 300m+ deals already") {
		// 	zap.S().Debug(repliesResp)
		// }
		// if strings.Contains(body, "deal, not the annual salary.") {
		// 	zap.S().Debug(repliesResp)
		// }
		// if strings.Contains(body, "About 500m after taxes and before") {
		// 	zap.S().Debug(repliesResp)
		// }
		// if strings.Contains(body, "other projects, compound interest. where does that get") {
		// 	zap.S().Debug(repliesResp)
		// }
		// if strings.Contains(body, "the one who said they could get there through salary alone") {
		// 	zap.S().Debug(repliesResp)
		// }
		// if strings.Contains(body, " will be close to 1 billion in salary. I didn't say NET") {
		// 	zap.S().Debug(repliesResp)
		// }
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

func (svc *service) GetRedditThreadWordsByLink(c context.Context, req *GetRedditThreadWordsByLinkReq) (*GetRedditThreadWordsRes, error) {
	link := createLink(req.Link)
	linkStr := fmt.Sprintf("%s/%s/%s/comments/%s", link.Protocol, link.DomainName, link.Subreddit, link.CommentId)
	scid := fmt.Sprintf("%s/comments/%s", link.Subreddit, link.CommentId)

	zap.S().Debugf("Checking if scid %s exists in db...", scid)

	if wordDocument, err := svc.Repository.GetWordsFromLink(c, scid); err == nil {
		if wordDocument != nil {
			if util.IsInLastWeek(wordDocument.LastUpdated.Time()) {
				zap.S().Debugf("Retrieved word document for %s from MongoDB Words Collection.", scid)
				return &GetRedditThreadWordsRes{Words: wordDocument.Words, Link: link.CommentId}, nil
			}
		}
	}

	zap.S().Debugf("Scid %s does not exist in db.", scid)

	redditReq, err := http.NewRequest("GET", fmt.Sprintf("%s.json", linkStr), nil)

	if err != nil {
		zap.S().Errorf("Could not create reddit request: ", err)
	}

	svc.rl.Take()
	zap.S().Debugf("Getting thread words for link %s/%s/%s/comments/%s...", link.Protocol, link.DomainName, link.Subreddit, link.CommentId)
	redditReq.Header.Set("User-Agent", "redditwordcloud/1.0")

	res, err := svc.client.Do(redditReq)

	if err != nil {
		zap.S().Errorf("client: could not create request: %s\n", err)
		os.Exit(1)
	}

	zap.S().Debugf("Successful GET request to %s.", fmt.Sprintf("%s.json", linkStr))

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		zap.S().Error("Error reading the response body:", err)
		return nil, nil
	}

	var RedditResponses []RedditResponse

	err = json.Unmarshal(body, &RedditResponses)

	if err != nil {
		zap.S().Error("Error unmarshaling res to JSON:", err)
		zap.S().Debug(string(body))
		return nil, nil
	}

	words := cmap.New[int]()

	for _, rr := range RedditResponses {
		svc.processRedditResponse(c, rr, &words, link, 0)
	}

	m := words.Items()
	zap.S().Infof("Created Word Map with %d entries.", len(m))
	svc.Repository.InsertWords(c, &m, scid)

	return &GetRedditThreadWordsRes{Words: m, Link: link.CommentId}, nil
}

func (svc *service) getCommentArticleResp(commentId string, link *Link) []RedditResponse {
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
		return nil
	}

	zap.S().Debugf("Body: %s", body)

	var CommentArticleAPIResponse []RedditResponse

	err = json.Unmarshal(body, &CommentArticleAPIResponse)

	if err != nil {
		zap.S().Error("Error unmarshaling res to JSON:", err)
		zap.S().Debug(string(body))
		return svc.getCommentArticleResp(commentId, link)
	}

	return CommentArticleAPIResponse
}
