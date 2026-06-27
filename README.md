# Jira Permission Exporter

A local Go application that exports Jira permission scheme data into CSV files.

It supports a step-based execution model, so you can:
- run the full export flow
- rerun only specific steps
- reuse previously generated CSV files for later steps

## Features

The exporter can collect:

- all permission schemes
- projects using each permission scheme
- permission grants within each scheme
- project role actors (users/groups assigned to roles)
- a final flattened export joining all of the above
- grouped split exports from the final flat export

## Output Files

The application writes CSV files under the `output/` directory.

Main outputs:

- `output/permission_schemes.csv`
- `output/scheme_projects.csv`
- `output/scheme_permissions.csv`
- `output/project_role_actors.csv`
- `output/final_flat_export.csv`

Optional split outputs:

- `output/project_grouping/`
- `output/holder_type/`

---

## Requirements

Before running the application, make sure you have:

- [Go](https://go.dev/) installed
- access to Jira Cloud
- a Jira API token

Recommended Go version:

- Go 1.22+
- the project may also work on nearby versions

You can verify your Go installation with:

```bash
go version
```

---

## Jira API Token

You need a Jira Cloud API token.

Generate it from your Atlassian account:

- Account Settings
- Security
- API tokens
- Create API token

You will use:

- your Atlassian email
- your API token

> Important: never commit your real token to Git.

---

## Project Setup

### 1. Clone the repository

```bash
git clone https://github.com/CKissas/jira-permission-exporter.git
cd jira-permission-exporter
```

### 2. Install Go dependencies

If dependencies are not already downloaded, run:

```bash
go mod tidy
```

This will install the required Go packages.

If you want, you can also explicitly install the YAML package:

```bash
go get gopkg.in/yaml.v3
```

In most cases, `go mod tidy` is enough.

---

## Configuration

### 1. Copy the example config

Create your local config file from the example:

```bash
cp config.example.yaml config.yaml
```

### 2. Edit `config.yaml`

Open the file and update it with your Jira details.

Example:

```yaml
base_url: "https://your-domain.atlassian.net"
email: "your-email@example.com"
api_token: "your-jira-api-token"
output_dir: "./output"

steps:
  permission_schemes: true
  scheme_projects: true
  scheme_permissions: true
  project_role_actors: true
  final_flat_export: true
  split_final_flat_export: false
```

### Config options

#### `base_url`

Your Jira Cloud base URL.

Example:

```yaml
base_url: "https://your-domain.atlassian.net"
```

#### `email`

Your Atlassian account email.

Example:

```yaml
email: "your-email@example.com"
```

#### `api_token`

Your Jira API token.

Example:

```yaml
api_token: "your-jira-api-token"
```

#### `output_dir`

Directory where CSV files will be written.

Example:

```yaml
output_dir: "./output"
```

---

## Steps Configuration

The app runs based on the `steps` section.

Each step can be enabled or disabled independently.

```yaml
steps:
  permission_schemes: true
  scheme_projects: true
  scheme_permissions: true
  project_role_actors: true
  final_flat_export: true
  split_final_flat_export: false
```

### Available steps

#### `permission_schemes`

Fetches all permission schemes from Jira and exports:

- `output/permission_schemes.csv`

#### `scheme_projects`

Fetches all projects and maps each project to its permission scheme.

Exports:

- `output/scheme_projects.csv`

#### `scheme_permissions`

Fetches expanded permission scheme data and exports scheme permission grants.

Exports:

- `output/scheme_permissions.csv`

#### `project_role_actors`

Fetches project roles and role actors for each project.

Exports:

- `output/project_role_actors.csv`

#### `final_flat_export`

Builds the final combined export using previously collected datasets.

Exports:

- `output/final_flat_export.csv`

#### `split_final_flat_export`

Reads `output/final_flat_export.csv` and splits it into smaller grouped files.

Exports:

- `output/project_grouping/*.csv`
- `output/holder_type/*.csv`

---

## Important Step Behavior

Some later steps depend on earlier outputs.

If you disable earlier steps, the application will try to load the required CSV files from `output/`.

Example:

```yaml
steps:
  permission_schemes: false
  scheme_projects: false
  scheme_permissions: false
  project_role_actors: false
  final_flat_export: true
  split_final_flat_export: false
```

This works only if these files already exist:

- `output/permission_schemes.csv`
- `output/scheme_projects.csv`
- `output/scheme_permissions.csv`
- `output/project_role_actors.csv`

If a required file is missing, the application will fail with a clear error message.

---

## How to Run

### Run with the default config file

```bash
go run ./cmd/exporter
```

By default, the app looks for:

```text
config.yaml
```

### Run with a custom config file

```bash
go run ./cmd/exporter ./my-config.yaml
```

---

## Example Execution Scenarios

### 1. Full export from Jira

```yaml
steps:
  permission_schemes: true
  scheme_projects: true
  scheme_permissions: true
  project_role_actors: true
  final_flat_export: true
  split_final_flat_export: false
```

Run:

```bash
go run ./cmd/exporter
```

### 2. Only build the final flat export from existing CSV files

```yaml
steps:
  permission_schemes: false
  scheme_projects: false
  scheme_permissions: false
  project_role_actors: false
  final_flat_export: true
  split_final_flat_export: false
```

Run:

```bash
go run ./cmd/exporter
```

### 3. Only split an already generated final flat export

```yaml
steps:
  permission_schemes: false
  scheme_projects: false
  scheme_permissions: false
  project_role_actors: false
  final_flat_export: false
  split_final_flat_export: true
```

Run:

```bash
go run ./cmd/exporter
```

---

## Output Structure

Typical structure after execution:

```text
output/
├── permission_schemes.csv
├── scheme_projects.csv
├── scheme_permissions.csv
├── project_role_actors.csv
├── final_flat_export.csv
├── project_grouping/
│   ├── ABC.csv
│   ├── XYZ.csv
│   └── ...
└── holder_type/
    ├── projectRole.csv
    ├── group.csv
    ├── user.csv
    └── ...
```

---

## Notes

- The exporter is intentionally sequential for simplicity and easier debugging.
- Large Jira instances can produce very large CSV files.
- The final flat export may contain millions of rows depending on the number of projects, permissions, and role assignments.
- The split export step helps make the final output easier to consume.

---

## Security Notes

- `config.yaml` is ignored by Git and should remain local only.
- Do not commit real credentials or generated output files.
- If a real token was ever accidentally committed, revoke it and generate a new one.

---

## Troubleshooting

### Missing required CSV file

If you enable a later step but disable earlier steps, the exporter may fail because it cannot find a required CSV file.

Fix:

- either enable the earlier step
- or make sure the required CSV already exists in `output/`

### Jira authentication issues

Check:

- `base_url`
- `email`
- `api_token`

Make sure:

- the API token is valid
- your account has permission to access the Jira data being exported

### Slow execution

This is expected for large Jira instances, especially for:

- project permission scheme mapping
- project role actor collection
- final flat export generation

---

## Development

Run formatting:

```bash
go fmt ./...
```

Run dependency cleanup:

```bash
go mod tidy
```

---

## License

See [LICENSE](./LICENSE)