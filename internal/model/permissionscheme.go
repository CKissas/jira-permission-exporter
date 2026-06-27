package model

type PermissionSchemesResponse struct {
	PermissionSchemes []PermissionScheme `json:"permissionSchemes"`
}

type PermissionScheme struct {
	ID          int64                 `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Permissions []PermissionGrantItem `json:"permissions"`
}

type PermissionGrantItem struct {
	ID         int64            `json:"id"`
	Permission string           `json:"permission"`
	Holder     PermissionHolder `json:"holder"`
}

type PermissionHolder struct {
	Type      string `json:"type"`
	Parameter string `json:"parameter"`
	Value     string `json:"value"`
}
