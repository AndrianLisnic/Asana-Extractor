package storage

import (
	"asana-extractor/internal/gorest"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveProjectsWritesOneJSONPerGID(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "snapshot", "projects")
	projects := []gorest.Project{
		{Resource: gorest.Resource{GID: "project-1", ResourceType: "project"}, Name: "First"},
		{Resource: gorest.Resource{GID: "project-2", ResourceType: "project"}, Name: "Second"},
	}

	count, err := SaveProjects(dir, projects)
	if err != nil {
		t.Fatal(err)
	}
	if count != len(projects) {
		t.Fatalf("count = %d, want %d", count, len(projects))
	}
	assertJSONFile(t, filepath.Join(dir, "project-1.json"), "First")
	assertJSONFile(t, filepath.Join(dir, "project-2.json"), "Second")
}

func TestSaveUsersWritesOneJSONPerGID(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "snapshot", "users")
	users := []gorest.User{{Resource: gorest.Resource{GID: "user-1", ResourceType: "user"}, Name: "Ada"}}

	count, err := SaveUsers(dir, users)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
	assertJSONFile(t, filepath.Join(dir, "user-1.json"), "Ada")
}

func TestSaveProjectsRejectsInvalidGID(t *testing.T) {
	_, err := SaveProjects(t.TempDir(), []gorest.Project{{Resource: gorest.Resource{GID: "../escape"}}})
	if err == nil {
		t.Fatal("expected invalid GID error")
	}
}

func assertJSONFile(t *testing.T, path, wantName string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["name"] != wantName {
		t.Errorf("name = %v, want %q", got["name"], wantName)
	}
}
