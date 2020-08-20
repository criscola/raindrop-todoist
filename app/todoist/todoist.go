package todoist

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/criscola/raindrop-todoist/logging"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const InvalidId = -1

var baseURL = url.URL{
	Scheme: "https",
	Host:   "api.todoist.com",
	Path:   "/rest/v1/",
}

// Client is a client for sending requests to the Raindrop public API.
type Client struct {
	c      *http.Client
	apiKey string
	taskFormat string
	syncToken string
	logger *logging.StandardLogger
}

type authorizedTransport struct {
	apiKey string
}

func (t *authorizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer " + t.apiKey)

	return http.DefaultTransport.RoundTrip(req)
}

// New creates a new Go client for the Raindrop public API.
func New(apiKey string, logger *logging.StandardLogger) *Client {
	c := &http.Client{Timeout: 15 * time.Second, Transport: &authorizedTransport{apiKey}}

	return &Client{
		c:      c,
		logger: logger,
	}
}

// GetCompletedReadings gets checked readings ids on Todoist, if there's any
func (c *Client) GetCompletedReadings() ([]int64, error) {
	return nil, nil
}

func (c *Client) NewReadingTask(title, bookmarkUrl, domain string) (int64, error) {
	postBody := strings.NewReader(`{"content": "Read [` + title  + `](` + bookmarkUrl + `) on `+ domain + `"}`)

	// Add task to todoist and save the id
	// TODO: Cut strings too long
	endpoint := baseURL.ResolveReference(&url.URL{Path: "tasks"})
	req, err := http.NewRequest("POST", endpoint.String(), postBody)
	if err != nil {
		// TODO logging
		panic(fmt.Errorf("Fatal error in building request: %s \n", err))
		return InvalidId, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := c.c.Do(req)
	// TODO logging
	if err != nil {
		panic(fmt.Errorf("Fatal error sending request: %s \n", err))
		return InvalidId, err
	}

	taskCreatedRes, err := ioutil.ReadAll(res.Body)
	// TODO logging
	if err != nil {
		panic(fmt.Errorf("Fatal error reading HTTP response: %s \n", err))
		return InvalidId, err
	}
	// Get task id from the response
	todoId, err := jsonparser.GetInt(taskCreatedRes, "id")
	// TODO logging
	if err != nil {
		panic(fmt.Errorf("Fatal error reading id from request: %s \n", err))
		return InvalidId, err
	}
	defer res.Body.Close()

	return todoId, nil
}

// TODO enable string customization through a config string
func formatTask(taskFormat string, attributes... string) string {
	return ""
}