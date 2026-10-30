package main

import (
	"asana-extractor/internal/config"
	"asana-extractor/internal/gorest"
	"asana-extractor/internal/scheduler"
	"asana-extractor/internal/storage"
	"context"
	"log"
	"path/filepath"
	"time"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	client := gorest.NewClient(conf)
	ctx := context.Background()

	task := func() error {
		startedAt := time.Now()
		log.Printf("Starting projects and users fetch at %s", startedAt.Format(time.RFC3339))

		projects, err := client.FetchAllProjects(ctx, conf.PerPage)
		if err != nil {
			log.Printf("Projects fetch failed after %s: %v", time.Since(startedAt).Round(time.Millisecond), err)
			return err
		}

		users, err := client.FetchAllUsers(ctx, conf.PerPage)
		if err != nil {
			log.Printf("Users fetch failed after %s: %v", time.Since(startedAt).Round(time.Millisecond), err)
			return err
		}

		snapshotDir, err := storage.NewSnapshotDir("data")
		if err != nil {
			return err
		}

		projectFiles, err := storage.SaveProjects(filepath.Join(snapshotDir, "projects"), projects)
		if err != nil {
			return err
		}
		userFiles, err := storage.SaveUsers(filepath.Join(snapshotDir, "users"), users)
		if err != nil {
			return err
		}

		finishedAt := time.Now()
		log.Printf(
			"Extraction completed: %d projects and %d users in %s; JSON files saved under %s",
			projectFiles,
			userFiles,
			finishedAt.Sub(startedAt).Round(time.Millisecond),
			snapshotDir,
		)
		log.Printf(
			"Next projects and users fetch in %s",
			conf.Interval,
		)
		return nil
	}

	if err := scheduler.Run(ctx, conf.Interval, task); err != nil {
		log.Fatal(err)
	}
}
