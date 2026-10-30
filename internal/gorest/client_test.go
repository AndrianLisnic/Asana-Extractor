package gorest

import (
	"asana-extractor/internal/config"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestFetchAllProjectsPaginatesWithWorkspace(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api/1.0/projects" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected authorization header")
		}
		if r.URL.Query().Get("workspace") != "workspace-1" || r.URL.Query().Get("limit") != "2" {
			t.Errorf("unexpected query: %v", r.URL.Query())
		}
		if calls == 1 {
			fmt.Fprint(w, `{"data":[{"gid":"p1","resource_type":"project","name":"One"}],"next_page":{"offset":"cursor/+"}}`)
			return
		}
		if got := r.URL.Query().Get("offset"); got != "cursor/+" {
			t.Errorf("offset = %q", got)
		}
		fmt.Fprint(w, `{"data":[{"gid":"p2","resource_type":"project","name":"Two"}]}`)
	}))
	defer server.Close()

	client := NewClient(config.Config{BaseURL: server.URL, Token: "test-token", WorkspaceGID: "workspace-1"})
	client.MinRequestInterval = 0
	projects, err := client.FetchAllProjects(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 || projects[1].GID != "p2" || calls != 2 {
		t.Fatalf("projects=%+v calls=%d", projects, calls)
	}
}

func TestFetchAllUsersUsesWorkspaceEndpointAndOffset(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api/1.0/workspaces/workspace-1/users" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if calls == 1 {
			fmt.Fprint(w, `{"data":[{"gid":"u1","resource_type":"user","name":"Ada"}],"next_page":{"offset":"next"}}`)
			return
		}
		if r.URL.Query().Get("offset") != "next" {
			t.Errorf("offset = %q", r.URL.Query().Get("offset"))
		}
		fmt.Fprint(w, `{"data":[{"gid":"u2","resource_type":"user","name":"Lin"}]}`)
	}))
	defer server.Close()

	client := NewClient(config.Config{BaseURL: server.URL, Token: "token", WorkspaceGID: "workspace-1"})
	client.MinRequestInterval = 0
	users, err := client.FetchAllUsers(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 || users[1].GID != "u2" || calls != 2 {
		t.Fatalf("users=%+v calls=%d", users, calls)
	}
}

func TestFetchAllProjectsRejectsInvalidPageSize(t *testing.T) {
	client := &Client{}
	for _, pageSize := range []int{0, 101} {
		if _, err := client.FetchAllProjects(context.Background(), pageSize); err == nil {
			t.Errorf("expected error for page size %d", pageSize)
		}
	}
}

func TestRetryAfterParsesSeconds(t *testing.T) {
	response := &http.Response{Header: http.Header{"Retry-After": []string{"3"}}}
	if got := retryAfter(response); got != 3*time.Second {
		t.Errorf("retryAfter = %s", url.QueryEscape(got.String()))
	}
}
