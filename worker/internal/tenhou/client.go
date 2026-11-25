package tenhou

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

const (
	rawIndexPath = "/sc/raw/list.cgi"
	rawFilePath  = "/sc/raw/dat/"
	logPath      = "/0/log/?"
	maxBody      = 32 << 20
)

// Client talks to tenhou.net with one request at a time and a minimum gap between requests
// (Tenhou allows a single concurrent session). Share one Client per process.
type Client struct {
	Base      string // https://tenhou.net (overridable in tests)
	HTTP      *http.Client
	UserAgent string
	MinGap    time.Duration

	mu   sync.Mutex
	last time.Time
}

func NewClient(userAgent string) *Client {
	return &Client{Base: "https://tenhou.net", HTTP: &http.Client{Timeout: 60 * time.Second}, UserAgent: userAgent, MinGap: 2 * time.Second}
}

func (c *Client) get(ctx context.Context, u string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if wait := c.MinGap - time.Since(c.last); wait > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
	defer func() { c.last = time.Now() }()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{URL: u, Status: resp.StatusCode}
	}
	return body, nil
}

type HTTPError struct {
	URL    string
	Status int
}

func (e *HTTPError) Error() string { return fmt.Sprintf("GET %s: HTTP %d", e.URL, e.Status) }

// Index fetches the sc/raw index (list.cgi). Callers must poll at most every 20 minutes.
func (c *Client) Index(ctx context.Context) ([]RawFile, error) {
	b, err := c.get(ctx, c.Base+rawIndexPath)
	if err != nil {
		return nil, err
	}
	return ParseIndex(b)
}

var safeFile = regexp.MustCompile(`^[a-z]{3}\d{8,10}\.(log|html)\.gz$`)

// RawFile downloads and gunzips one .gz file named in the index (URLs always built from /sc/raw/dat/).
func (c *Client) RawFile(ctx context.Context, name string) ([]byte, error) {
	if !safeFile.MatchString(name) {
		return nil, fmt.Errorf("refusing unexpected raw file name %q", name)
	}
	b, err := c.get(ctx, c.Base+rawFilePath+name)
	if err != nil {
		return nil, err
	}
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	defer zr.Close()
	return io.ReadAll(io.LimitReader(zr, maxBody))
}

// Log downloads a game's mjlog XML.
func (c *Client) Log(ctx context.Context, logID string) ([]byte, error) {
	return c.get(ctx, c.Base+logPath+url.QueryEscape(logID))
}
