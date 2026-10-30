package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	BaseURL         string
	Token           string
	WorkspaceGID    string
	Interval        time.Duration
	PerPage         int
	OutputPath      string
	UsersOutputPath string
}

func Load() (Config, error) {
	token := os.Getenv("EXTRACTOR_TOKEN")
	if token == "" {
		return Config{}, fmt.Errorf("EXTRACTOR_TOKEN is not configured")
	}

	workspaceGID := os.Getenv("ASANA_WORKSPACE_GID")
	if workspaceGID == "" {
		return Config{}, fmt.Errorf("ASANA_WORKSPACE_GID is not configured")
	}

	intervalText := os.Getenv("PULL_INTERVAL")
	if intervalText == "" {
		intervalText = "30s"
	}

	interval, err := time.ParseDuration(intervalText)
	if err != nil || interval <= 0 {
		return Config{}, fmt.Errorf("invalid PULL_INTERVAL: %q", intervalText)
	}

	perPage := 10

	if value := os.Getenv("PER_PAGE"); value != "" {
		perPage, err = strconv.Atoi(value)
		if err != nil || perPage <= 0 || perPage > 100 {
			return Config{}, fmt.Errorf("invalid PER_PAGE: %q", value)
		}
	}

	baseURL := os.Getenv("EXTRACTOR_BASE_URL")
	if baseURL == "" {
		baseURL = "https://app.asana.com"
	}

	outputPath := os.Getenv("PROJECTS_OUTPUT_PATH")
	if outputPath == "" {
		outputPath = os.Getenv("OUTPUT_PATH")
	}
	if outputPath == "" {
		outputPath = "data/projects.json"
	}

	usersOutputPath := os.Getenv("USERS_OUTPUT_PATH")
	if usersOutputPath == "" {
		usersOutputPath = "data/users.json"
	}

	return Config{
		BaseURL:         baseURL,
		Token:           token,
		WorkspaceGID:    workspaceGID,
		Interval:        interval,
		PerPage:         perPage,
		OutputPath:      outputPath,
		UsersOutputPath: usersOutputPath,
	}, nil
}
