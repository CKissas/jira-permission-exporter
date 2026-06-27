package model

type ProjectRolesMap map[string]string

type ProjectRoleDetail struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Actors      []ProjectRoleActor `json:"actors"`
}

type ProjectRoleActor struct {
	ID          int64       `json:"id"`
	DisplayName string      `json:"displayName"`
	Type        string      `json:"type"`
	ActorUser   *ActorUser  `json:"actorUser"`
	ActorGroup  *ActorGroup `json:"actorGroup"`
}

type ActorUser struct {
	AccountID string `json:"accountId"`
}

type ActorGroup struct {
	Name string `json:"name"`
}

type ProjectRoleActorRow struct {
	ProjectID        string
	ProjectKey       string
	ProjectName      string
	RoleID           int64
	RoleName         string
	RoleDescription  string
	ActorID          int64
	ActorType        string
	ActorDisplayName string
	ActorAccountID   string
	ActorGroupName   string
}
