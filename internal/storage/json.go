package storage

import (
	"asana-extractor/internal/gorest"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NewSnapshotDir creates a unique timestamped directory under baseDir for one
// complete extraction cycle.
func NewSnapshotDir(baseDir string) (string, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	for attempt := 0; attempt < 100; attempt++ {
		name := time.Now().Format("20060102-150405.000000000")
		if attempt > 0 {
			name = fmt.Sprintf("%s-%02d", name, attempt)
		}
		dir := filepath.Join(baseDir, name)
		if err := os.Mkdir(dir, 0755); err == nil {
			return dir, nil
		} else if !os.IsExist(err) {
			return "", fmt.Errorf("create snapshot directory: %w", err)
		}
	}
	return "", fmt.Errorf("could not create a unique snapshot directory under %q", baseDir)
}

func SaveProjects(dir string, projects []gorest.Project) (int, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("create projects directory: %w", err)
	}
	for _, project := range projects {
		if err := saveJSONByGID(dir, project.GID, project); err != nil {
			return 0, fmt.Errorf("save project %q: %w", project.GID, err)
		}
	}
	return len(projects), nil
}

func SaveUsers(dir string, users []gorest.User) (int, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("create users directory: %w", err)
	}
	for _, user := range users {
		if err := saveJSONByGID(dir, user.GID, user); err != nil {
			return 0, fmt.Errorf("save user %q: %w", user.GID, err)
		}
	}
	return len(users), nil
}

func saveJSONByGID(dir, gid string, value any) error {
	if gid == "" || filepath.Base(gid) != gid || strings.ContainsAny(gid, `/\\`) {
		return fmt.Errorf("invalid or empty GID %q", gid)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, gid+".json"), data, 0644); err != nil {
		return fmt.Errorf("write JSON file: %w", err)
	}
	return nil
}
