package reddit

import (
	"net/http"
	"net/url"
	"redditwordcloud/pkg/retryhttp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type Handler struct {
	Service
	httpClient *http.Client
}

func NewHandler(s Service) *Handler {
	return &Handler{
		Service:    s,
		httpClient: retryhttp.NewRetryableClient(),
	}
}

func (h *Handler) GetRedditThreadWordsByThreadIDHandler(c *gin.Context) {
	threadId := c.Param("threadId")

	// resp, err := h.httpClient.Get("https://google.com")

	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, "error")
	// }

	c.JSON(http.StatusOK, threadId)
}

func ValidateLink(fl validator.FieldLevel) bool {
	link := fl.Field().String()
	_, err := url.ParseRequestURI(link)
	return err == nil
}

func (h *Handler) GetRedditThreadWordsByLinkHandler(c *gin.Context) {
	txn := newrelic.FromContext(c)
	var req GetRedditThreadWordsByLinkReq

	segment := txn.StartSegment("BindJSON")
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	segment.End()

	segment = txn.StartSegment("GetRedditThreadWordsByLink Service")
	res, err := h.Service.GetRedditThreadWordsByLink(c, &req, txn)
	segment.End()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set the response content type to JSON
	c.Header("Content-Type", "application/json")

	// Write the JSON data to the response body
	c.JSON(http.StatusOK, res)
}
