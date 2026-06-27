package main

import (
	"fmt"
	"os"
	"strconv"

	"jira-permission-exporter/internal/config"
	"jira-permission-exporter/internal/export"
	"jira-permission-exporter/internal/jira"
	"jira-permission-exporter/internal/model"
)

func main() {
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	client := jira.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)

	var schemes []model.PermissionScheme
	var schemeMap map[int64]string

	needsSchemes := cfg.Steps.PermissionSchemes || cfg.Steps.SchemeProjects || cfg.Steps.FinalFlatExport

	if needsSchemes {
		if cfg.Steps.PermissionSchemes {
			fmt.Println("Fetching permission schemes from Jira...")
			schemes, err = client.GetPermissionSchemes()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Jira API error fetching permission schemes: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Fetched %d permission schemes\n", len(schemes))

			outputPath, err := export.WritePermissionSchemesCSV(cfg.OutputDir, schemes)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Export error writing permission schemes CSV: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Permission schemes CSV exported: %s\n", outputPath)
		} else {
			fmt.Println("Loading permission schemes from existing CSV...")
			if !export.PermissionSchemesCSVExists(cfg.OutputDir) {
				fmt.Fprintf(
					os.Stderr,
					"Required file not found: %s. Either enable steps.permission_schemes or provide the existing CSV.\n",
					export.PermissionSchemesCSVPath(cfg.OutputDir),
				)
				os.Exit(1)
			}

			schemes, err = export.ReadPermissionSchemesCSV(cfg.OutputDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load permission schemes from CSV: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Loaded %d permission schemes from CSV\n", len(schemes))
		}

		schemeMap = make(map[int64]string)
		for _, s := range schemes {
			schemeMap[s.ID] = s.Name
		}
	} else {
		fmt.Println("Skipping permission schemes step")
	}

	var schemeProjects []model.SchemeProjectRow
	needsSchemeProjects := cfg.Steps.SchemeProjects || cfg.Steps.ProjectRoleActors || cfg.Steps.FinalFlatExport

	if needsSchemeProjects {
		if cfg.Steps.SchemeProjects {
			fmt.Println("Fetching all projects from Jira...")
			projects, err := client.GetAllProjects()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Jira API error fetching projects: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Fetched %d projects\n", len(projects))

			for i, p := range projects {
				fmt.Printf("Processing project %d/%d: %s\n", i+1, len(projects), p.Key)

				scheme, err := client.GetProjectPermissionScheme(p.Key)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to fetch permission scheme for project %s: %v\n", p.Key, err)
					continue
				}

				schemeName := scheme.Name
				if schemeName == "" && schemeMap != nil {
					if name, ok := schemeMap[scheme.ID]; ok {
						schemeName = name
					}
				}

				schemeProjects = append(schemeProjects, model.SchemeProjectRow{
					SchemeID:    scheme.ID,
					SchemeName:  schemeName,
					ProjectID:   p.ID,
					ProjectKey:  p.Key,
					ProjectName: p.Name,
					ProjectType: p.ProjectTypeKey,
				})
			}

			outputPath, err := export.WriteSchemeProjectsCSV(cfg.OutputDir, schemeProjects)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Export error writing scheme_projects CSV: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Scheme-project mapping CSV exported: %s\n", outputPath)
		} else {
			fmt.Println("Loading scheme-project mappings from existing CSV...")
			if !export.SchemeProjectsCSVExists(cfg.OutputDir) {
				fmt.Fprintf(
					os.Stderr,
					"Required file not found: %s. Either enable steps.scheme_projects or provide the existing CSV.\n",
					export.SchemeProjectsCSVPath(cfg.OutputDir),
				)
				os.Exit(1)
			}

			schemeProjects, err = export.ReadSchemeProjectsCSV(cfg.OutputDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load scheme-project mappings from CSV: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Loaded %d scheme-project mappings from CSV\n", len(schemeProjects))
		}
	} else {
		fmt.Println("Skipping scheme projects step")
	}

	if cfg.Steps.SchemePermissions {
		fmt.Println("Fetching expanded permission schemes for scheme permissions export...")
		expandedSchemes, err := client.GetPermissionSchemesExpanded()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Jira API error fetching expanded permission schemes: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Fetched %d expanded permission schemes\n", len(expandedSchemes))

		var rows []model.SchemePermissionRow

		for _, scheme := range expandedSchemes {
			for _, permission := range scheme.Permissions {
				rows = append(rows, model.SchemePermissionRow{
					SchemeID:        scheme.ID,
					SchemeName:      scheme.Name,
					PermissionKey:   permission.Permission,
					HolderType:      permission.Holder.Type,
					HolderParameter: permission.Holder.Parameter,
					HolderValue:     permission.Holder.Value,
				})
			}
		}

		outputPath, err := export.WriteSchemePermissionsCSV(cfg.OutputDir, rows)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Export error writing scheme_permissions CSV: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Scheme permissions CSV exported: %s\n", outputPath)
	} else {
		fmt.Println("Skipping scheme permissions step")
	}

	if cfg.Steps.ProjectRoleActors {
		fmt.Println("Collecting project role actors...")

		var rows []model.ProjectRoleActorRow

		for i, project := range schemeProjects {
			fmt.Printf("Processing project roles %d/%d: %s\n", i+1, len(schemeProjects), project.ProjectKey)

			rolesMap, err := client.GetProjectRoles(project.ProjectKey)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to fetch roles for project %s: %v\n", project.ProjectKey, err)
				continue
			}

			for _, roleURL := range rolesMap {
				roleDetail, err := client.GetProjectRoleDetail(project.ProjectKey, roleURL)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to fetch role detail for project %s: %v\n", project.ProjectKey, err)
					continue
				}

				if len(roleDetail.Actors) == 0 {
					continue
				}

				for _, actor := range roleDetail.Actors {
					accountID := ""
					groupName := ""

					if actor.ActorUser != nil {
						accountID = actor.ActorUser.AccountID
					}
					if actor.ActorGroup != nil {
						groupName = actor.ActorGroup.Name
					}

					rows = append(rows, model.ProjectRoleActorRow{
						ProjectID:        project.ProjectID,
						ProjectKey:       project.ProjectKey,
						ProjectName:      project.ProjectName,
						RoleID:           roleDetail.ID,
						RoleName:         roleDetail.Name,
						RoleDescription:  roleDetail.Description,
						ActorID:          actor.ID,
						ActorType:        actor.Type,
						ActorDisplayName: actor.DisplayName,
						ActorAccountID:   accountID,
						ActorGroupName:   groupName,
					})
				}
			}
		}

		outputPath, err := export.WriteProjectRoleActorsCSV(cfg.OutputDir, rows)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Export error writing project_role_actors CSV: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Project role actors CSV exported: %s\n", outputPath)
	} else {
		fmt.Println("Skipping project role actors step")
	}

	if cfg.Steps.FinalFlatExport {
		fmt.Println("Loading data for final flat export...")

		if !export.PermissionSchemesCSVExists(cfg.OutputDir) {
			fmt.Fprintf(os.Stderr, "Required file not found: %s\n", export.PermissionSchemesCSVPath(cfg.OutputDir))
			os.Exit(1)
		}
		if !export.SchemeProjectsCSVExists(cfg.OutputDir) {
			fmt.Fprintf(os.Stderr, "Required file not found: %s\n", export.SchemeProjectsCSVPath(cfg.OutputDir))
			os.Exit(1)
		}
		if !export.SchemePermissionsCSVExists(cfg.OutputDir) {
			fmt.Fprintf(os.Stderr, "Required file not found: %s\n", export.SchemePermissionsCSVPath(cfg.OutputDir))
			os.Exit(1)
		}
		if !export.ProjectRoleActorsCSVExists(cfg.OutputDir) {
			fmt.Fprintf(os.Stderr, "Required file not found: %s\n", export.ProjectRoleActorsCSVPath(cfg.OutputDir))
			os.Exit(1)
		}

		permissionSchemes, err := export.ReadPermissionSchemesCSV(cfg.OutputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read permission schemes CSV: %v\n", err)
			os.Exit(1)
		}

		schemeProjectsRows, err := export.ReadSchemeProjectsCSV(cfg.OutputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read scheme projects CSV: %v\n", err)
			os.Exit(1)
		}

		schemePermissionRows, err := export.ReadSchemePermissionsCSV(cfg.OutputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read scheme permissions CSV: %v\n", err)
			os.Exit(1)
		}

		projectRoleActorRows, err := export.ReadProjectRoleActorsCSV(cfg.OutputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read project role actors CSV: %v\n", err)
			os.Exit(1)
		}

		schemeDescriptionMap := make(map[int64]string)
		for _, scheme := range permissionSchemes {
			schemeDescriptionMap[scheme.ID] = scheme.Description
		}

		projectsBySchemeID := make(map[int64][]model.SchemeProjectRow)
		for _, row := range schemeProjectsRows {
			projectsBySchemeID[row.SchemeID] = append(projectsBySchemeID[row.SchemeID], row)
		}

		roleActorsByProjectAndRole := make(map[string][]model.ProjectRoleActorRow)
		for _, row := range projectRoleActorRows {
			key := row.ProjectKey + "|" + strconv.FormatInt(row.RoleID, 10)
			roleActorsByProjectAndRole[key] = append(roleActorsByProjectAndRole[key], row)
		}

		var finalRows []model.FinalFlatExportRow

		for _, permissionRow := range schemePermissionRows {
			projects := projectsBySchemeID[permissionRow.SchemeID]
			if len(projects) == 0 {
				continue
			}

			for _, project := range projects {
				if permissionRow.HolderType == "projectRole" {
					roleID, err := strconv.ParseInt(permissionRow.HolderParameter, 10, 64)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: invalid projectRole holder_parameter for scheme %d permission %s: %s\n",
							permissionRow.SchemeID, permissionRow.PermissionKey, permissionRow.HolderParameter)
						continue
					}

					key := project.ProjectKey + "|" + strconv.FormatInt(roleID, 10)
					actors := roleActorsByProjectAndRole[key]

					for _, actor := range actors {
						finalRows = append(finalRows, model.FinalFlatExportRow{
							SchemeID:          permissionRow.SchemeID,
							SchemeName:        permissionRow.SchemeName,
							SchemeDescription: schemeDescriptionMap[permissionRow.SchemeID],
							ProjectID:         project.ProjectID,
							ProjectKey:        project.ProjectKey,
							ProjectName:       project.ProjectName,
							ProjectType:       project.ProjectType,
							PermissionKey:     permissionRow.PermissionKey,
							HolderType:        permissionRow.HolderType,
							HolderParameter:   permissionRow.HolderParameter,
							HolderValue:       permissionRow.HolderValue,
							RoleID:            actor.RoleID,
							RoleName:          actor.RoleName,
							RoleDescription:   actor.RoleDescription,
							ActorID:           actor.ActorID,
							ActorType:         actor.ActorType,
							ActorDisplayName:  actor.ActorDisplayName,
							ActorAccountID:    actor.ActorAccountID,
							ActorGroupName:    actor.ActorGroupName,
						})
					}
				} else {
					finalRows = append(finalRows, model.FinalFlatExportRow{
						SchemeID:          permissionRow.SchemeID,
						SchemeName:        permissionRow.SchemeName,
						SchemeDescription: schemeDescriptionMap[permissionRow.SchemeID],
						ProjectID:         project.ProjectID,
						ProjectKey:        project.ProjectKey,
						ProjectName:       project.ProjectName,
						ProjectType:       project.ProjectType,
						PermissionKey:     permissionRow.PermissionKey,
						HolderType:        permissionRow.HolderType,
						HolderParameter:   permissionRow.HolderParameter,
						HolderValue:       permissionRow.HolderValue,
						RoleID:            0,
						RoleName:          "",
						RoleDescription:   "",
						ActorID:           0,
						ActorType:         "",
						ActorDisplayName:  "",
						ActorAccountID:    "",
						ActorGroupName:    "",
					})
				}
			}
		}

		outputPath, err := export.WriteFinalFlatExportCSV(cfg.OutputDir, finalRows)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Export error writing final flat export CSV: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Final flat export CSV exported: %s\n", outputPath)
	} else {
		fmt.Println("Skipping final flat export step")
	}

	if cfg.Steps.SplitFinalFlatExport {
		fmt.Println("Splitting final flat export into grouped files...")

		if !export.FinalFlatExportCSVExists(cfg.OutputDir) {
			fmt.Fprintf(
				os.Stderr,
				"Required file not found: %s. Either enable steps.final_flat_export or provide the existing CSV.\n",
				export.FinalFlatExportCSVPath(cfg.OutputDir),
			)
			os.Exit(1)
		}

		if err := export.SplitFinalFlatExport(cfg.OutputDir); err != nil {
			fmt.Fprintf(os.Stderr, "Split export error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Split exports created under:\n- %s\n- %s\n",
			cfg.OutputDir+"/project_grouping",
			cfg.OutputDir+"/holder_type",
		)
	} else {
		fmt.Println("Skipping split final flat export step")
	}
}
