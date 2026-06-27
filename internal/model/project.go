package model

type ProjectSearchResponse struct {
	Values     []ProjectSummary `json:"values"`
	IsLast     bool             `json:"isLast"`
	StartAt    int              `json:"startAt"`
	MaxResults int              `json:"maxResults"`
	Total      int              `json:"total"`
}

type ProjectSummary struct {
	ID             string `json:"id"`
	Key            string `json:"key"`
	Name           string `json:"name"`
	ProjectTypeKey string `json:"projectTypeKey"`
}

type ProjectPermissionScheme struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SchemeProjectRow struct {
	SchemeID    int64
	SchemeName  string
	ProjectID   string
	ProjectKey  string
	ProjectName string
	ProjectType string
}
