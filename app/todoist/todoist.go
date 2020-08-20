package todoist

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/criscola/raindrop-todoist/logging"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const InvalidId = -1

// Don't change this outside of New()
var baseURL = url.URL{
	Scheme: "https",
	Host:   "api.todoist.com",
	Path:   "/sync/v8/sync",
}

// Client is a client for sending requests to the Raindrop public API.
type Client struct {
	c      *http.Client
	apiKey string
	taskFormat string
	syncToken string
	logger *logging.StandardLogger
}

// New creates a new Go client for the Raindrop public API.
func New(apiKey string, logger *logging.StandardLogger) *Client {
	c := &http.Client{Timeout: 15 * time.Second}

	q := baseURL.Query() // Get a copy of the query values.
	q.Add("token", apiKey) // Add a new value to the set.
	baseURL.RawQuery = q.Encode() // Encode and assign back to the original query.

	return &Client{
		c:      c,
		apiKey: apiKey,
		syncToken: "*",
		logger: logger,
	}
}

// GetCompletedReadings gets the new tasks since the last request which have been checked (i.e. completed) by the user
// in the meantime. Lookup https://developer.todoist.com/sync/v8/#sync
/*
func (c *Client) GetCompletedReadings() ([]int64, error) {
	req, err := http.NewRequest("GET", "/sync/v8/sync", nil)
	if err != nil {
		return nil, err
	}

	parameters := url.Values{}
	parameters.Add("")
}*/

func (c *Client) NewReadingTask(title, bookmarkUrl, domain string) (int64, error) {
	taskContent := `Read [` + title  + `](` + bookmarkUrl + `) on `+ domain + `}]`
	cmdUuid := uuid.New().String()
	tempId := uuid.New().String()

	q := baseURL.Query() // Get a copy of the query values.
	q.Add("commands", `[{"type": "item_add", "temp_id": "` + tempId + `", "uuid": "` + cmdUuid + `", "args": {"content": "` + taskContent + `"}}]`)

	endpoint := baseURL.ResolveReference(&url.URL{RawQuery: q.Encode()})

	// Add task to todoist and save the id
	// TODO: Cut strings too long
	req, err := http.NewRequest("POST", endpoint.String(), nil)
	if err != nil {
		// TODO logging
		panic(fmt.Errorf("Fatal error in building request: %s \n", err))
		return InvalidId, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

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
	defer res.Body.Close()

	// Get task id from the response
	todoId, err := jsonparser.GetInt(taskCreatedRes, "temp_id_mapping", tempId)

	// TODO logging
	if err != nil {
		panic(fmt.Errorf("Fatal error reading id from request: %s \n", err))
		return InvalidId, err
	}

	c.syncToken, err = jsonparser.GetString(taskCreatedRes, "sync_token")
	if err != nil {
		panic(fmt.Errorf("Fatal error reading sync_id from request: %s \n", err))
		return InvalidId, err
	}

	return todoId, nil
}

// TODO enable string customization through a config string
func formatTask(taskFormat string, attributes... string) string {
	return ""
}