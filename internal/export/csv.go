package export

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"jira-permission-exporter/internal/model"
)

const (
	PermissionSchemesFile = "permission_schemes.csv"
	SchemeProjectsFile    = "scheme_projects.csv"
	SchemePermissionsFile = "scheme_permissions.csv"
	ProjectRoleActorsFile = "project_role_actors.csv"
	FinalFlatExportFile   = "final_flat_export.csv"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func PermissionSchemesCSVPath(outputDir string) string {
	return filepath.Join(outputDir, PermissionSchemesFile)
}

func SchemeProjectsCSVPath(outputDir string) string {
	return filepath.Join(outputDir, SchemeProjectsFile)
}

func SchemePermissionsCSVPath(outputDir string) string {
	return filepath.Join(outputDir, SchemePermissionsFile)
}

func ProjectRoleActorsCSVPath(outputDir string) string {
	return filepath.Join(outputDir, ProjectRoleActorsFile)
}

func FinalFlatExportCSVPath(outputDir string) string {
	return filepath.Join(outputDir, FinalFlatExportFile)
}

func PermissionSchemesCSVExists(outputDir string) bool {
	return FileExists(PermissionSchemesCSVPath(outputDir))
}

func SchemeProjectsCSVExists(outputDir string) bool {
	return FileExists(SchemeProjectsCSVPath(outputDir))
}

func SchemePermissionsCSVExists(outputDir string) bool {
	return FileExists(SchemePermissionsCSVPath(outputDir))
}

func ProjectRoleActorsCSVExists(outputDir string) bool {
	return FileExists(ProjectRoleActorsCSVPath(outputDir))
}

func FinalFlatExportCSVExists(outputDir string) bool {
	return FileExists(FinalFlatExportCSVPath(outputDir))
}

func WritePermissionSchemesCSV(outputDir string, schemes []model.PermissionScheme) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	path := PermissionSchemesCSVPath(outputDir)

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create csv file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"scheme_id", "scheme_name", "description"}); err != nil {
		return "", fmt.Errorf("write csv header: %w", err)
	}

	for _, s := range schemes {
		record := []string{
			fmt.Sprintf("%d", s.ID),
			s.Name,
			s.Description,
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("write csv record: %w", err)
		}
	}

	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("flush csv writer: %w", err)
	}

	return path, nil
}

func ReadPermissionSchemesCSV(outputDir string) ([]model.PermissionScheme, error) {
	path := PermissionSchemesCSVPath(outputDir)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open permission schemes csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read permission schemes csv: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("permission schemes csv is empty")
	}

	var schemes []model.PermissionScheme

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			return nil, fmt.Errorf("invalid permission schemes csv row %d", i+1)
		}

		id, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse scheme_id on row %d: %w", i+1, err)
		}

		schemes = append(schemes, model.PermissionScheme{
			ID:          id,
			Name:        row[1],
			Description: row[2],
		})
	}

	return schemes, nil
}

func WriteSchemeProjectsCSV(outputDir string, rows []model.SchemeProjectRow) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	path := SchemeProjectsCSVPath(outputDir)

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create csv file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"scheme_id",
		"scheme_name",
		"project_id",
		"project_key",
		"project_name",
		"project_type",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("write csv header: %w", err)
	}

	for _, row := range rows {
		record := []string{
			fmt.Sprintf("%d", row.SchemeID),
			row.SchemeName,
			row.ProjectID,
			row.ProjectKey,
			row.ProjectName,
			row.ProjectType,
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("write csv record: %w", err)
		}
	}

	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("flush csv writer: %w", err)
	}

	return path, nil
}

func ReadSchemeProjectsCSV(outputDir string) ([]model.SchemeProjectRow, error) {
	path := SchemeProjectsCSVPath(outputDir)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open scheme projects csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read scheme projects csv: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("scheme projects csv is empty")
	}

	var result []model.SchemeProjectRow

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 6 {
			return nil, fmt.Errorf("invalid scheme projects csv row %d", i+1)
		}

		schemeID, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse scheme_id on row %d: %w", i+1, err)
		}

		result = append(result, model.SchemeProjectRow{
			SchemeID:    schemeID,
			SchemeName:  row[1],
			ProjectID:   row[2],
			ProjectKey:  row[3],
			ProjectName: row[4],
			ProjectType: row[5],
		})
	}

	return result, nil
}

func WriteSchemePermissionsCSV(outputDir string, rows []model.SchemePermissionRow) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	path := SchemePermissionsCSVPath(outputDir)

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create csv file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"scheme_id",
		"scheme_name",
		"permission_key",
		"holder_type",
		"holder_parameter",
		"holder_value",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("write csv header: %w", err)
	}

	for _, row := range rows {
		record := []string{
			fmt.Sprintf("%d", row.SchemeID),
			row.SchemeName,
			row.PermissionKey,
			row.HolderType,
			row.HolderParameter,
			row.HolderValue,
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("write csv record: %w", err)
		}
	}

	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("flush csv writer: %w", err)
	}

	return path, nil
}

func ReadSchemePermissionsCSV(outputDir string) ([]model.SchemePermissionRow, error) {
	path := SchemePermissionsCSVPath(outputDir)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open scheme permissions csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read scheme permissions csv: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("scheme permissions csv is empty")
	}

	var result []model.SchemePermissionRow

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 6 {
			return nil, fmt.Errorf("invalid scheme permissions csv row %d", i+1)
		}

		schemeID, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse scheme_id on row %d: %w", i+1, err)
		}

		result = append(result, model.SchemePermissionRow{
			SchemeID:        schemeID,
			SchemeName:      row[1],
			PermissionKey:   row[2],
			HolderType:      row[3],
			HolderParameter: row[4],
			HolderValue:     row[5],
		})
	}

	return result, nil
}

func WriteProjectRoleActorsCSV(outputDir string, rows []model.ProjectRoleActorRow) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	path := ProjectRoleActorsCSVPath(outputDir)

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create csv file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"project_id",
		"project_key",
		"project_name",
		"role_id",
		"role_name",
		"role_description",
		"actor_id",
		"actor_type",
		"actor_display_name",
		"actor_account_id",
		"actor_group_name",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("write csv header: %w", err)
	}

	for _, row := range rows {
		record := []string{
			row.ProjectID,
			row.ProjectKey,
			row.ProjectName,
			fmt.Sprintf("%d", row.RoleID),
			row.RoleName,
			row.RoleDescription,
			fmt.Sprintf("%d", row.ActorID),
			row.ActorType,
			row.ActorDisplayName,
			row.ActorAccountID,
			row.ActorGroupName,
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("write csv record: %w", err)
		}
	}

	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("flush csv writer: %w", err)
	}

	return path, nil
}

func ReadProjectRoleActorsCSV(outputDir string) ([]model.ProjectRoleActorRow, error) {
	path := ProjectRoleActorsCSVPath(outputDir)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open project role actors csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read project role actors csv: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("project role actors csv is empty")
	}

	var result []model.ProjectRoleActorRow

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 11 {
			return nil, fmt.Errorf("invalid project role actors csv row %d", i+1)
		}

		roleID, err := strconv.ParseInt(row[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse role_id on row %d: %w", i+1, err)
		}

		actorID, err := strconv.ParseInt(row[6], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse actor_id on row %d: %w", i+1, err)
		}

		result = append(result, model.ProjectRoleActorRow{
			ProjectID:        row[0],
			ProjectKey:       row[1],
			ProjectName:      row[2],
			RoleID:           roleID,
			RoleName:         row[4],
			RoleDescription:  row[5],
			ActorID:          actorID,
			ActorType:        row[7],
			ActorDisplayName: row[8],
			ActorAccountID:   row[9],
			ActorGroupName:   row[10],
		})
	}

	return result, nil
}

func WriteFinalFlatExportCSV(outputDir string, rows []model.FinalFlatExportRow) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	path := FinalFlatExportCSVPath(outputDir)

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create csv file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"scheme_id",
		"scheme_name",
		"scheme_description",
		"project_id",
		"project_key",
		"project_name",
		"project_type",
		"permission_key",
		"holder_type",
		"holder_parameter",
		"holder_value",
		"role_id",
		"role_name",
		"role_description",
		"actor_id",
		"actor_type",
		"actor_display_name",
		"actor_account_id",
		"actor_group_name",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("write csv header: %w", err)
	}

	for _, row := range rows {
		record := []string{
			fmt.Sprintf("%d", row.SchemeID),
			row.SchemeName,
			row.SchemeDescription,
			row.ProjectID,
			row.ProjectKey,
			row.ProjectName,
			row.ProjectType,
			row.PermissionKey,
			row.HolderType,
			row.HolderParameter,
			row.HolderValue,
			fmt.Sprintf("%d", row.RoleID),
			row.RoleName,
			row.RoleDescription,
			fmt.Sprintf("%d", row.ActorID),
			row.ActorType,
			row.ActorDisplayName,
			row.ActorAccountID,
			row.ActorGroupName,
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("write csv record: %w", err)
		}
	}

	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("flush csv writer: %w", err)
	}

	return path, nil
}

func sanitizeFileName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "empty"
	}

	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	value = replacer.Replace(value)

	return value
}

func SplitFinalFlatExport(outputDir string) error {
	inputPath := FinalFlatExportCSVPath(outputDir)
	if !FileExists(inputPath) {
		return fmt.Errorf("final flat export file does not exist: %s", inputPath)
	}

	projectDir := filepath.Join(outputDir, "project_grouping")
	holderTypeDir := filepath.Join(outputDir, "holder_type")

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("create project grouping directory: %w", err)
	}
	if err := os.MkdirAll(holderTypeDir, 0755); err != nil {
		return fmt.Errorf("create holder type directory: %w", err)
	}

	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open final flat export file: %w", err)
	}
	defer inputFile.Close()

	reader := csv.NewReader(bufio.NewReader(inputFile))

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read final flat export header: %w", err)
	}

	const (
		projectKeyIndex = 4
		holderTypeIndex = 8
	)

	projectFiles := make(map[string]*os.File)
	projectWriters := make(map[string]*csv.Writer)

	holderFiles := make(map[string]*os.File)
	holderWriters := make(map[string]*csv.Writer)

	closeAll := func() {
		for _, w := range projectWriters {
			w.Flush()
		}
		for _, f := range projectFiles {
			f.Close()
		}
		for _, w := range holderWriters {
			w.Flush()
		}
		for _, f := range holderFiles {
			f.Close()
		}
	}

	defer closeAll()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read final flat export row: %w", err)
		}

		if len(record) <= holderTypeIndex {
			return fmt.Errorf("invalid final flat export row: expected at least %d columns, got %d", holderTypeIndex+1, len(record))
		}

		projectKey := sanitizeFileName(record[projectKeyIndex])
		holderType := sanitizeFileName(record[holderTypeIndex])

		if _, exists := projectWriters[projectKey]; !exists {
			path := filepath.Join(projectDir, projectKey+".csv")
			file, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("create project grouping file %s: %w", path, err)
			}

			writer := csv.NewWriter(file)
			if err := writer.Write(header); err != nil {
				file.Close()
				return fmt.Errorf("write project grouping header %s: %w", path, err)
			}

			projectFiles[projectKey] = file
			projectWriters[projectKey] = writer
		}

		if _, exists := holderWriters[holderType]; !exists {
			path := filepath.Join(holderTypeDir, holderType+".csv")
			file, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("create holder type file %s: %w", path, err)
			}

			writer := csv.NewWriter(file)
			if err := writer.Write(header); err != nil {
				file.Close()
				return fmt.Errorf("write holder type header %s: %w", path, err)
			}

			holderFiles[holderType] = file
			holderWriters[holderType] = writer
		}

		if err := projectWriters[projectKey].Write(record); err != nil {
			return fmt.Errorf("write project grouping row for %s: %w", projectKey, err)
		}

		if err := holderWriters[holderType].Write(record); err != nil {
			return fmt.Errorf("write holder type row for %s: %w", holderType, err)
		}
	}

	for key, writer := range projectWriters {
		writer.Flush()
		if err := writer.Error(); err != nil {
			return fmt.Errorf("flush project grouping writer %s: %w", key, err)
		}
	}

	for key, writer := range holderWriters {
		writer.Flush()
		if err := writer.Error(); err != nil {
			return fmt.Errorf("flush holder type writer %s: %w", key, err)
		}
	}

	return nil
}
