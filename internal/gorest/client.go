package gorest

import (
	"asana-extractor/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const apiPath = "/api/1.0"

type Client struct {
	BaseURL            string
	Token              string
	WorkspaceGID       string
	HTTPClient         *http.Client
	MaxRetries         int
	MinRequestInterval time.Duration

	rateMu      sync.Mutex
	lastRequest time.Time
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		BaseURL:            strings.TrimRight(cfg.BaseURL, "/"),
		Token:              cfg.Token,
		WorkspaceGID:       cfg.WorkspaceGID,
		HTTPClient:         &http.Client{Timeout: 15 * time.Second},
		MaxRetries:         3,
		MinRequestInterval: time.Second,
	}
}

func (c *Client) apiURL(path string) string {
	base := strings.TrimRight(c.BaseURL, "/")
	if strings.HasSuffix(base, apiPath) {
		return base + path
	}
	return base + apiPath + path
}

func (c *Client) projectsURL() string {
	return c.apiURL("/projects")
}

func (c *Client) usersURL() string {
	return c.apiURL("/workspaces/" + url.PathEscape(c.WorkspaceGID) + "/users")
}

func (c *Client) fetchPage(ctx context.Context, endpoint string, offset string, limit int, target any) error {
	requestURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parse API URL: %w", err)
	}

	query := requestURL.Query()
	query.Set("limit", strconv.Itoa(limit))
	if strings.HasSuffix(requestURL.Path, "/projects") {
		if c.WorkspaceGID == "" {
			return fmt.Errorf("ASANA_WORKSPACE_GID is required to paginate projects")
		}
		query.Set("workspace", c.WorkspaceGID)
	}
	if offset != "" {
		query.Set("offset", offset)
	}
	requestURL.RawQuery = query.Encode()

	retries := c.MaxRetries
	if retries < 1 {
		retries = 1
	}

	var lastErr error
	for attempt := 1; attempt <= retries; attempt++ {
		if err := c.waitForRateLimit(ctx); err != nil {
			return err
		}

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		request.Header.Set("Authorization", "Bearer "+c.Token)
		request.Header.Set("Accept", "application/json")

		response, err := c.HTTPClient.Do(request)
		if err != nil {
			lastErr = fmt.Errorf("send request: %w", err)
			if err := c.retryWait(ctx, attempt, 0); err != nil {
				return err
			}
			continue
		}

		if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500 {
			delay := retryAfter(response)
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
			lastErr = fmt.Errorf("temporary API response: status %d", response.StatusCode)
			if attempt == retries {
				break
			}
			if err := c.retryWait(ctx, attempt, delay); err != nil {
				return err
			}
			continue
		}

		if response.StatusCode != http.StatusOK {
			body, readErr := io.ReadAll(response.Body)
			response.Body.Close()
			if readErr != nil {
				return fmt.Errorf("API returned status %d; read body: %w", response.StatusCode, readErr)
			}
			return fmt.Errorf("API returned status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
		}

		decodeErr := json.NewDecoder(response.Body).Decode(target)
		response.Body.Close()
		if decodeErr != nil {
			return fmt.Errorf("decode API response: %w", decodeErr)
		}
		return nil
	}

	return fmt.Errorf("API request failed after %d attempts: %w", retries, lastErr)
}

func (c *Client) fetchProjectsPage(ctx context.Context, offset string, limit int) ([]Project, string, error) {
	var payload ProjectsResponse
	err := c.fetchPage(ctx, c.projectsURL(), offset, limit, &payload)
	if err != nil {
		return nil, "", err
	}
	return payload.Data, nextOffset(payload.NextPage), nil
}

func (c *Client) FetchAllProjects(ctx context.Context, perPage int) ([]Project, error) {
	if perPage < 1 || perPage > 100 {
		return nil, fmt.Errorf("perPage must be between 1 and 100: %d", perPage)
	}

	var allProjects []Project
	offset := ""
	for {
		projects, next, err := c.fetchProjectsPage(ctx, offset, perPage)
		if err != nil {
			return allProjects, err
		}
		allProjects = append(allProjects, projects...)
		if next == "" {
			return allProjects, nil
		}
		offset = next
	}
}

func (c *Client) fetchUsersPage(ctx context.Context, offset string, limit int) ([]User, string, error) {
	var payload UsersResponse
	err := c.fetchPage(ctx, c.usersURL(), offset, limit, &payload)
	if err != nil {
		return nil, "", err
	}
	return payload.Data, nextOffset(payload.NextPage), nil
}

func (c *Client) FetchAllUsers(ctx context.Context, perPage int) ([]User, error) {
	if perPage < 1 || perPage > 100 {
		return nil, fmt.Errorf("perPage must be between 1 and 100: %d", perPage)
	}

	var allUsers []User
	offset := ""
	for {
		users, next, err := c.fetchUsersPage(ctx, offset, perPage)
		if err != nil {
			return allUsers, err
		}
		allUsers = append(allUsers, users...)
		if next == "" {
			return allUsers, nil
		}
		offset = next
	}
}

func nextOffset(page *NextPage) string {
	if page == nil {
		return ""
	}
	return page.Offset
}

func (c *Client) waitForRateLimit(ctx context.Context) error {
	c.rateMu.Lock()
	wait := time.Duration(0)
	if !c.lastRequest.IsZero() {
		wait = time.Until(c.lastRequest.Add(c.MinRequestInterval))
	}
	if wait < 0 {
		wait = 0
	}
	c.lastRequest = time.Now().Add(wait)
	c.rateMu.Unlock()
	return sleepContext(ctx, wait)
}

func (c *Client) retryWait(ctx context.Context, attempt int, retryAfter time.Duration) error {
	if retryAfter <= 0 {
		retryAfter = time.Duration(attempt) * time.Second
	}
	return sleepContext(ctx, retryAfter)
}

func retryAfter(response *http.Response) time.Duration {
	value := strings.TrimSpace(response.Header.Get("Retry-After"))
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if retryAt, err := http.ParseTime(value); err == nil {
		if wait := time.Until(retryAt); wait > 0 {
			return wait
		}
	}
	return 0
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
