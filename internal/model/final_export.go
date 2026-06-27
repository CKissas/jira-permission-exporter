package model

type FinalFlatExportRow struct {
	SchemeID          int64
	SchemeName        string
	SchemeDescription string
	ProjectID         string
	ProjectKey        string
	ProjectName       string
	ProjectType       string
	PermissionKey     string
	HolderType        string
	HolderParameter   string
	HolderValue       string
	RoleID            int64
	RoleName          string
	RoleDescription   string
	ActorID           int64
	ActorType         string
	ActorDisplayName  string
	ActorAccountID    string
	ActorGroupName    string
}
