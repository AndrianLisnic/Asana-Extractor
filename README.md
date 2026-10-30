# Asana Extractor

A Go command-line program that periodically extracts projects and users from an Asana workspace and saves each resource as an individual JSON file.

## Features

- Fetches projects from `GET /api/1.0/projects`.
- Fetches workspace users from `GET /api/1.0/workspaces/{workspace_gid}/users`.
- Uses Bearer token authentication and cursor-based pagination.
- Respects a minimum interval between API requests and retries rate-limit (`429`) and server (`5xx`) responses, honoring `Retry-After` when provided.
- Saves each resource as `<gid>.json` under a timestamped extraction directory.
- Runs an extraction immediately on startup and repeats after the configured interval.

## Requirements

- Go 1.27 or later.
- An Asana Personal Access Token with permission to read projects and users in the target workspace. The token is sent as `Authorization: Bearer <token>`.
- The GID of the Asana workspace or organization.

The workspace users endpoint returns compact user records and is limited to 2,000 users. For larger workspaces, use Asana's `/users` endpoint and filter/paginate appropriately.

## Configuration

Set these environment variables in PowerShell before running the program:

```powershell
$env:EXTRACTOR_TOKEN="your_asana_personal_access_token"
$env:ASANA_WORKSPACE_GID="your_workspace_gid"
$env:PULL_INTERVAL="30s"
$env:PER_PAGE="50"
```

Optional variables:

```powershell
$env:EXTRACTOR_BASE_URL="https://app.asana.com"
$env:PROJECTS_OUTPUT_PATH="data/projects/projects.json"
$env:USERS_OUTPUT_PATH="data/users/users.json"
```

`PULL_INTERVAL` accepts Go duration values such as `30s` or `5m`; the default is `30s`. `PER_PAGE` must be between 1 and 100; the default is 10. `PROJECTS_OUTPUT_PATH` and `USERS_OUTPUT_PATH` are currently retained as configuration defaults for compatibility; output is organized under `data/<timestamp>/`.

## Run

```powershell
go run .\cmd\gorest\
```

The program performs the first extraction immediately, then waits for the configured interval after each completed cycle. It logs the start, failures, completion duration, resource counts, and output directory.

## Output

Each extraction creates a new timestamped directory under `data`. Projects and users are separated into subdirectories, and each JSON file is named by its Asana GID:

```text
data/
└── 20261030-103713.660808300/
    ├── projects/
    │   ├── 1220049786700509.json
    │   └── <project-gid>.json
    └── users/
        ├── <user-gid>.json
        └── ...
```

## Tests

Run all unit tests with:

```powershell
go test ./...
```

The tests cover API pagination and request parameters, JSON-per-resource storage, GID validation, retry header parsing, and scheduler behavior. API tests use an in-process HTTP test server and do not require an Asana account or live token.

## Scaling notes

The client requests pages sequentially using Asana's `next_page.offset` cursor and applies a shared minimum request interval across project and user requests. This avoids page-number assumptions and keeps API traffic bounded. The workspace users endpoint has a documented 2,000-user limit; use the general users endpoint for larger organizations. For very large snapshots, consider streaming pages directly to disk instead of keeping every resource in memory before saving.
