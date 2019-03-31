package clients

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/appengine/urlfetch"
)

// HTTPClient generic interface
type HTTPClient interface {
	Get(string) (*http.Response, error)
}

type googleHTTPClient struct {
	client *http.Client
}

// NewGoogleHTTPClient implementation
func NewGoogleHTTPClient(ctx context.Context) HTTPClient {
	return &googleHTTPClient{client: urlfetch.Client(ctx)}
}

func (c *googleHTTPClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

type goHTTPClient struct {
	client *http.Client
}

// NewGoPClient implementation
func NewGoPClient() HTTPClient {
	return &goHTTPClient{client: &http.Client{Timeout: 5 * time.Second}}
}

func (c *goHTTPClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}
