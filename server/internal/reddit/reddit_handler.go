package reddit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"redditwordcloud/pkg/retryhttp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

	if err != nil {
		return false
	}
	return true
}

func (h *Handler) GetRedditThreadWordsByLinkHandler(c *gin.Context) {
	var req GetRedditThreadWordsByLinkReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.Service.GetRedditThreadWordsByLink(c, &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Errorf("err: %w", err))
	}

	wordsJson, err := json.Marshal(res.Words)

	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Errorf("err: %w", err))
	}

	// Set the response content type to JSON
	c.Header("Content-Type", "application/json")

	// Write the JSON data to the response body
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(wordsJson)
}
