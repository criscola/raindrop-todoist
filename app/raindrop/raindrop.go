package raindrop

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/criscola/raindrop-todoist/logging"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var baseURL = url.URL{
	Scheme: "https",
	Host:   "api.raindrop.io",
	Path:   "/rest/v1/",
}

type PostponedReading struct {
	BookmarkId int64
	Domain string
	Url string
	Title string
}

// Client is a client for sending requests to the Raindrop public API.
type Client struct {
	c      *http.Client
	apiKey string
	postponedLabelName string
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
func New(apiKey string, postponedLabelName string, logger *logging.StandardLogger) *Client {
	c := &http.Client{ Timeout: 15 * time.Second, Transport: &authorizedTransport{apiKey} }

	return &Client{
		c:                  c,
		apiKey:             apiKey,
		postponedLabelName: postponedLabelName,
		logger:             logger,
	}
}

// GetPostponedReadings pulls the postponed readings marked with a configured label from Raindrop and returns them
// except raindrops with specified id
func (c *Client) GetPostponedReadings(exclusions []int64) (prs []PostponedReading, err error) {
	// TODO try with baseURL.String()
	req, err := http.NewRequest("GET","https://api.raindrop.io/rest/v1/raindrops/0" , nil)
	// TODO: Pass err and log error up
	if err != nil {
		log.Fatal().
			Err(err).
			Str("service", "Raindrop client").
			Str("function", "GetPostponedReadings").
			Msg("Error building request")
		return nil, err
	}

	// Raindrop API doesn't rigorously respect RFC 3986 so commas shouldn't be encoded, Go will always encode them so
	// the URL encoding must be handled specifically for this case
	req.URL.Opaque = `/rest/v1/raindrops/0?search=%5B%7B%22key%22:%22tag%22,%22val%22:%22` + c.postponedLabelName +
		`%22%7D%5D`

	res, err := c.c.Do(req)
	// TODO log error
	if err != nil {
		panic(fmt.Errorf("Fatal error sending request: %s \n", err))
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(fmt.Errorf("Error reading response body: %s \n", err))
	}

	fmt.Println(formatRequest(req))
	fmt.Println(string(body))
	fmt.Println(c.postponedLabelName)
	prs = []PostponedReading{}

	// Iterate over every item returned by request
	_, err = jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		bookmarkId, err := jsonparser.GetInt(value, "_id")

		if err != nil {
			panic(fmt.Errorf("Fatal error sending request: %s \n", err))
		}

		if len(exclusions) == 0 || !contains(exclusions, bookmarkId) {
			// If bookmark doesn't exist, create the relative Todoist task and add IDs to the db

			domain, err := jsonparser.GetString(value, "domain")
			if err != nil {
				panic(fmt.Errorf("Fatal error reading response domain: %s \n", err))
			}
			bookmarkUrl, err := jsonparser.GetString(value, "link")
			if err != nil {
				panic(fmt.Errorf("Fatal error reading response link: %s \n", err))
			}
			title, err := jsonparser.GetString(value, "title")
			if err != nil {
				panic(fmt.Errorf("Fatal error reading response title: %s \n", err))
			}

			prs = append(prs, PostponedReading{
				bookmarkId,
				domain,
				bookmarkUrl,
				title,
			})

		}
	}, "items")
	if err != nil {
		panic(fmt.Errorf("Error parsing response body: %s \n", err))
	}
	return prs, nil
}

func RemoveLabelFromBookmark(bookmarkId int64) error {
	return nil
}

func contains(s []int64, k int64) bool {
	for _, e := range s {
		if k == e {
			return true
		}
	}
	return false
}

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string // Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url) // Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host)) // Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}   // Return the request as a string
	return strings.Join(request, "\n")
}