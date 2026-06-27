package model

type SchemePermissionRow struct {
	SchemeID        int64
	SchemeName      string
	PermissionKey   string
	HolderType      string
	HolderParameter string
	HolderValue     string
}
